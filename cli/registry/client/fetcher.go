package client

import (
	"context"
	"encoding/json"

	"github.com/distribution/reference"
	"github.com/docker/cli/cli/manifest/types"
	"github.com/docker/cli/internal/registry"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/ocischema"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/registry/api/errcode"
	v2 "github.com/docker/distribution/registry/api/v2"
	distclient "github.com/docker/distribution/registry/client"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// fetchManifest pulls a manifest from a registry and returns it. An error
// is returned if no manifest is found matching namedRef.
func fetchManifest(ctx context.Context, repo distribution.Repository, ref reference.Named) (types.ImageManifest, error) {
	manifest, err := getManifest(ctx, repo, ref)
	if err != nil {
		return types.ImageManifest{}, err
	}

	switch v := manifest.(type) {
	// Removed Schema 1 support
	case *schema2.DeserializedManifest:
		return pullManifestSchemaV2(ctx, ref, repo, *v)
	case *ocischema.DeserializedManifest:
		return pullManifestOCISchema(ctx, ref, repo, *v)
	case *manifestlist.DeserializedManifestList:
		return types.ImageManifest{}, errors.Errorf("%s is a manifest list", ref)
	}
	return types.ImageManifest{}, errors.Errorf("%s is not a manifest", ref)
}

func fetchList(ctx context.Context, repo distribution.Repository, ref reference.Named) ([]types.ImageManifest, error) {
	manifest, err := getManifest(ctx, repo, ref)
	if err != nil {
		return nil, err
	}

	switch v := manifest.(type) {
	case *manifestlist.DeserializedManifestList:
		return pullManifestList(ctx, ref, repo, *v)
	default:
		return nil, errors.Errorf("unsupported manifest format: %v", v)
	}
}

func getManifest(ctx context.Context, repo distribution.Repository, ref reference.Named) (distribution.Manifest, error) {
	manSvc, err := repo.Manifests(ctx)
	if err != nil {
		return nil, err
	}

	dgst, opts, err := getManifestOptionsFromReference(ref)
	if err != nil {
		return nil, errors.Errorf("image manifest for %q does not exist", ref)
	}
	return manSvc.Get(ctx, dgst, opts...)
}

func pullManifestSchemaV2(ctx context.Context, ref reference.Named, repo distribution.Repository, mfst schema2.DeserializedManifest) (types.ImageManifest, error) {
	manifestDesc, err := validateManifestDigest(ref, mfst)
	if err != nil {
		return types.ImageManifest{}, err
	}
	configJSON, err := pullManifestSchemaV2ImageConfig(ctx, mfst.Target().Digest, repo)
	if err != nil {
		return types.ImageManifest{}, err
	}

	if manifestDesc.Platform == nil {
		manifestDesc.Platform = &ocispec.Platform{}
	}

	// Fill in os and architecture fields from config JSON
	if err := json.Unmarshal(configJSON, manifestDesc.Platform); err != nil {
		return types.ImageManifest{}, err
	}

	return types.NewImageManifest(ref, manifestDesc, &mfst), nil
}

func pullManifestOCISchema(ctx context.Context, ref reference.Named, repo distribution.Repository, mfst ocischema.DeserializedManifest) (types.ImageManifest, error) {
	manifestDesc, err := validateManifestDigest(ref, mfst)
	if err != nil {
		return types.ImageManifest{}, err
	}
	configJSON, err := pullManifestSchemaV2ImageConfig(ctx, mfst.Target().Digest, repo)
	if err != nil {
		return types.ImageManifest{}, err
	}

	if manifestDesc.Platform == nil {
		manifestDesc.Platform = &ocispec.Platform{}
	}

	// Fill in os and architecture fields from config JSON
	if err := json.Unmarshal(configJSON, manifestDesc.Platform); err != nil {
		return types.ImageManifest{}, err
	}

	return types.NewOCIImageManifest(ref, manifestDesc, &mfst), nil
}

func pullManifestSchemaV2ImageConfig(ctx context.Context, dgst digest.Digest, repo distribution.Repository) ([]byte, error) {
	blobs := repo.Blobs(ctx)
	configJSON, err := blobs.Get(ctx, dgst)
	if err != nil {
		return nil, err
	}

	verifier := dgst.Verifier()
	if _, err := verifier.Write(configJSON); err != nil {
		return nil, err
	}
	if !verifier.Verified() {
		return nil, errors.Errorf("image config verification failed for digest %s", dgst)
	}
	return configJSON, nil
}

// validateManifestDigest computes the manifest digest, and, if pulling by
// digest, ensures that it matches the requested digest.
func validateManifestDigest(ref reference.Named, mfst distribution.Manifest) (ocispec.Descriptor, error) {
	mediaType, canonical, err := mfst.Payload()
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	desc := ocispec.Descriptor{
		Digest:    digest.FromBytes(canonical),
		Size:      int64(len(canonical)),
		MediaType: mediaType,
	}

	// If pull by digest, then verify the manifest digest.
	if digested, isDigested := ref.(reference.Canonical); isDigested && digested.Digest() != desc.Digest {
		return ocispec.Descriptor{}, errors.Errorf("manifest verification failed for digest %s", digested.Digest())
	}

	return desc, nil
}

// pullManifestList handles "manifest lists" which point to various
// platform-specific manifests.
func pullManifestList(ctx context.Context, ref reference.Named, repo distribution.Repository, mfstList manifestlist.DeserializedManifestList) ([]types.ImageManifest, error) {
	if _, err := validateManifestDigest(ref, mfstList); err != nil {
		return nil, err
	}

	infos := make([]types.ImageManifest, 0, len(mfstList.Manifests))
	for _, manifestDescriptor := range mfstList.Manifests {
		manSvc, err := repo.Manifests(ctx)
		if err != nil {
			return nil, err
		}
		manifest, err := manSvc.Get(ctx, manifestDescriptor.Digest)
		if err != nil {
			return nil, err
		}

		manifestRef, err := reference.WithDigest(ref, manifestDescriptor.Digest)
		if err != nil {
			return nil, err
		}

		var imageManifest types.ImageManifest
		switch v := manifest.(type) {
		case *schema2.DeserializedManifest:
			imageManifest, err = pullManifestSchemaV2(ctx, manifestRef, repo, *v)
		case *ocischema.DeserializedManifest:
			imageManifest, err = pullManifestOCISchema(ctx, manifestRef, repo, *v)
		default:
			err = errors.Errorf("unsupported manifest type: %T", manifest)
		}
		if err != nil {
			return nil, err
		}

		// Replace platform from config
		p := manifestDescriptor.Platform
		imageManifest.Descriptor.Platform = types.OCIPlatform(&p)

		infos = append(infos, imageManifest)
	}
	return infos, nil
}

func continueOnError(err error) bool {
	switch v := err.(type) {
	case errcode.Errors:
		if len(v) == 0 {
			return true
		}
		return continueOnError(v[0])
	case errcode.Error:
		switch e := err.(errcode.Error); e.Code {
		case errcode.ErrorCodeUnauthorized, v2.ErrorCodeManifestUnknown, v2.ErrorCodeNameUnknown:
			return true
		default:
			return false
		}
	case *distclient.UnexpectedHTTPResponseError:
		return true
	}
	return false
}

func (c *client) iterateEndpoints(ctx context.Context, namedRef reference.Named, each func(context.Context, distribution.Repository, reference.Named) (bool, error)) error {
	endpoints, err := allEndpoints(namedRef, c.insecureRegistry)
	if err != nil {
		return err
	}

	repoName := reference.TrimNamed(namedRef)
	repoInfo := registry.ParseRepositoryInfo(namedRef)
	indexInfo := repoInfo.Index

	confirmedTLSRegistries := make(map[string]bool)
	for _, endpoint := range endpoints {
		if endpoint.URL.Scheme != "https" {
			if _, confirmedTLS := confirmedTLSRegistries[endpoint.URL.Host]; confirmedTLS {
				logrus.Debugf("skipping non-TLS endpoint %s for host/port that appears to use TLS", endpoint.URL)
				continue
			}
		}

		if c.insecureRegistry {
			endpoint.TLSConfig.InsecureSkipVerify = true
		}
		repoEndpoint := repositoryEndpoint{
			repoName:  repoName,
			indexInfo: indexInfo,
			endpoint:  endpoint,
		}
		repo, err := c.getRepositoryForReference(ctx, namedRef, repoEndpoint)
		if err != nil {
			logrus.Debugf("error %s with repo endpoint %+v", err, repoEndpoint)
			var protoErr httpProtoError
			if errors.As(err, &protoErr) {
				continue
			}
			return err
		}

		if endpoint.URL.Scheme == "http" && !c.insecureRegistry {
			logrus.Debugf("skipping non-tls registry endpoint: %s", endpoint.URL)
			continue
		}
		done, err := each(ctx, repo, namedRef)
		if err != nil {
			if continueOnError(err) {
				if endpoint.URL.Scheme == "https" {
					confirmedTLSRegistries[endpoint.URL.Host] = true
				}
				logrus.Debugf("continuing on error (%T) %s", err, err)
				continue
			}
			logrus.Debugf("not continuing on error (%T) %s", err, err)
			return err
		}
		if done {
			return nil
		}
	}
	return newNotFoundError(namedRef.String())
}

// allEndpoints returns a list of endpoints ordered by priority (v2, http).
func allEndpoints(namedRef reference.Named, insecure bool) ([]registry.APIEndpoint, error) {
	var serviceOpts registry.ServiceOptions
	if insecure {
		logrus.Debugf("allowing insecure registry for: %s", reference.Domain(namedRef))
		serviceOpts.InsecureRegistries = []string{reference.Domain(namedRef)}
	}
	registryService, err := registry.NewService(serviceOpts)
	if err != nil {
		return nil, err
	}
	repoInfo := registry.ParseRepositoryInfo(namedRef)
	endpoints, err := registryService.Endpoints(context.TODO(), reference.Domain(repoInfo.Name))
	logrus.Debugf("endpoints for %s: %v", namedRef, endpoints)
	return endpoints, err
}

func newNotFoundError(ref string) *notFoundError {
	return &notFoundError{err: errors.New("no such manifest: " + ref)}
}

type notFoundError struct {
	err error
}

func (n *notFoundError) Error() string {
	return n.err.Error()
}

// NotFound satisfies interface github.com/docker/docker/errdefs.ErrNotFound
func (notFoundError) NotFound() {}
