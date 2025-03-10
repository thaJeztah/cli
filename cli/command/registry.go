package command

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/distribution/reference"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/credentials"
	configtypes "github.com/docker/cli/cli/config/types"
	"github.com/docker/cli/cli/hints"
	"github.com/docker/cli/cli/streams"
	"github.com/docker/cli/internal/prompt"
	"github.com/docker/cli/internal/tui"
	registrytypes "github.com/docker/docker/api/types/registry"
	"github.com/morikuni/aec"
	"github.com/pkg/errors"
)

const (
	registerSuggest = "Log in with your Docker ID or email address to push and pull images from Docker Hub. " +
		"If you don't have a Docker ID, head over to https://hub.docker.com/ to create one."
	patSuggest = "You can log in with your password or a Personal Access " +
		"Token (PAT). Using a limited-scope PAT grants better security and is required " +
		"for organizations using SSO. Learn more at https://docs.docker.com/go/access-tokens/"
)

// authConfigKey is the key used to store credentials for Docker Hub. It is
// a copy of [registry.IndexServer].
//
// [registry.IndexServer]: https://pkg.go.dev/github.com/docker/docker/registry#IndexServer
const authConfigKey = "https://index.docker.io/v1/"

// NewAuthRequester returns a RequestPrivilegeFunc for the specified registry
// and the given cmdName (used as informational message to the user).
//
// The returned function is a [registrytypes.RequestAuthConfig] to prompt the user
// for credentials if needed. It is called as fallback if the credentials (if any)
// used for the initial operation did not work.
//
// TODO(thaJeztah): cli Cli could be a Streams if it was not for cli.SetIn to be needed?
// TODO(thaJeztah): ideally, this would accept reposName / imageRef as a regular string (we can parse it if needed!), or .. maybe generics and accept either?
func NewAuthRequester(cli Cli, indexServer string, promptMsg string) registrytypes.RequestAuthConfig {
	configKey := getAuthConfigKey(indexServer)
	return newPrivilegeFunc(cli, configKey, promptMsg)
}

// RegistryAuthenticationPrivilegedFunc returns a RequestPrivilegeFunc from the specified registry index info
// for the given command.
//
// Deprecated: this function was only used internally and will be removed in the next release. Use
func RegistryAuthenticationPrivilegedFunc(cli Cli, index *registrytypes.IndexInfo, cmdName string) registrytypes.RequestAuthConfig {
	indexServer := index.Name
	configKey := getAuthConfigKey(indexServer)
	if index.Official {
		configKey = authConfigKey
	}
	promptMsg := fmt.Sprintf("Login prior to %s:", cmdName)
	return newPrivilegeFunc(cli, configKey, promptMsg)
}

func newPrivilegeFunc(cli Cli, indexServer string, promptMsg string) registrytypes.RequestAuthConfig {
	return func(ctx context.Context) (string, error) {
		// TODO(thaJeztah): can we make the prompt an argument? ("prompt func()" or "prompt func()?
		_, _ = fmt.Fprint(cli.Out(), "\n"+promptMsg+"\n")
		isDefaultRegistry := indexServer == authConfigKey
		authConfig, err := GetDefaultAuthConfig(cli.ConfigFile(), true, indexServer, isDefaultRegistry)
		if err != nil {
			_, _ = fmt.Fprintf(cli.Err(), "Unable to retrieve stored credentials for %s, error: %s.\n", authConfigKey, err)
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		authConfig, err = PromptUserForCredentials(ctx, cli, "", "", authConfig.Username, authConfigKey)
		if err != nil {
			return "", err
		}
		return registrytypes.EncodeAuthConfig(authConfig)
	}
}

// ResolveAuthConfig returns auth-config for the given registry from the
// credential-store. It returns an empty AuthConfig if no credentials were
// found.
//
// It is similar to [registry.ResolveAuthConfig], but uses the credentials-
// store, instead of looking up credentials from a map.
func ResolveAuthConfig(cfg *configfile.ConfigFile, index *registrytypes.IndexInfo) registrytypes.AuthConfig {
	indexServer := index.Name
	configKey := getAuthConfigKey(indexServer)
	if index.Official {
		configKey = authConfigKey
	}

	a, _ := cfg.GetAuthConfig(configKey)
	return registrytypes.AuthConfig(a)
}

// GetDefaultAuthConfig gets the default auth config given a serverAddress
// If credentials for given serverAddress exists in the credential store, the configuration will be populated with values in it
func GetDefaultAuthConfig(cfg *configfile.ConfigFile, checkCredStore bool, serverAddress string, isDefaultRegistry bool) (registrytypes.AuthConfig, error) {
	if !isDefaultRegistry {
		// FIXME(thaJeztah): should the same logic be used for getAuthConfigKey ?? Looks like we're normalizing things here, but not elsewhere? why?
		serverAddress = credentials.ConvertToHostname(serverAddress)
	}
	authconfig := configtypes.AuthConfig{}
	var err error
	if checkCredStore {
		authconfig, err = cfg.GetAuthConfig(serverAddress)
		if err != nil {
			return registrytypes.AuthConfig{
				ServerAddress: serverAddress,
			}, err
		}
	}
	authconfig.ServerAddress = serverAddress
	authconfig.IdentityToken = ""
	return registrytypes.AuthConfig(authconfig), nil
}

// ConfigureAuth handles prompting of user's username and password if needed.
//
// Deprecated: use [PromptUserForCredentials] instead.
func ConfigureAuth(ctx context.Context, cli Cli, flUser, flPassword string, authConfig *registrytypes.AuthConfig, _ bool) error {
	defaultUsername := authConfig.Username
	serverAddress := authConfig.ServerAddress

	newAuthConfig, err := PromptUserForCredentials(ctx, cli, flUser, flPassword, defaultUsername, serverAddress)
	if err != nil {
		return err
	}

	authConfig.Username = newAuthConfig.Username
	authConfig.Password = newAuthConfig.Password
	return nil
}

// PromptUserForCredentials handles the CLI prompt for the user to input
// credentials.
// If argUser is not empty, then the user is only prompted for their password.
// If argPassword is not empty, then the user is only prompted for their username
// If neither argUser nor argPassword are empty, then the user is not prompted and
// an AuthConfig is returned with those values.
// If defaultUsername is not empty, the username prompt includes that username
// and the user can hit enter without inputting a username  to use that default
// username.
//
// TODO(thaJeztah): cli Cli could be a Streams if it was not for cli.SetIn to be needed?
func PromptUserForCredentials(ctx context.Context, cli Cli, argUser, argPassword, defaultUsername, serverAddress string) (registrytypes.AuthConfig, error) {
	// On Windows, force the use of the regular OS stdin stream.
	//
	// See:
	// - https://github.com/moby/moby/issues/14336
	// - https://github.com/moby/moby/issues/14210
	// - https://github.com/moby/moby/pull/17738
	//
	// TODO(thaJeztah): we need to confirm if this special handling is still needed, as we may not be doing this in other places.
	if runtime.GOOS == "windows" {
		cli.SetIn(streams.NewIn(os.Stdin))
	}

	argUser = strings.TrimSpace(argUser)
	if argUser == "" {
		if serverAddress == authConfigKey {
			// When signing in to the default (Docker Hub) registry, we display
			// hints for creating an account, and (if hints are enabled), using
			// a token instead of a password.
			_, _ = fmt.Fprintln(cli.Out(), registerSuggest)
			if hints.Enabled() {
				_, _ = fmt.Fprintln(cli.Out(), patSuggest)
				_, _ = fmt.Fprintln(cli.Out())
			}
		}

		var msg string
		defaultUsername = strings.TrimSpace(defaultUsername)
		if defaultUsername == "" {
			msg = "Username: "
		} else {
			msg = fmt.Sprintf("Username (%s): ", defaultUsername)
		}

		var err error
		argUser, err = prompt.ReadInput(ctx, cli.In(), cli.Out(), msg)
		if err != nil {
			return registrytypes.AuthConfig{}, err
		}
		if argUser == "" {
			argUser = defaultUsername
		}
		if argUser == "" {
			return registrytypes.AuthConfig{}, errors.Errorf("Error: Non-null Username Required")
		}
	}

	argPassword = strings.TrimSpace(argPassword)
	if argPassword == "" {
		restoreInput, err := prompt.DisableInputEcho(cli.In())
		if err != nil {
			return registrytypes.AuthConfig{}, err
		}
		defer func() {
			if err := restoreInput(); err != nil {
				// TODO(thaJeztah): we should consider printing instructions how
				//  to restore this manually (other than restarting the shell).
				//  e.g., 'run stty echo' when in a Linux or macOS shell, but
				//  PowerShell and CMD.exe may need different instructions.
				_, _ = fmt.Fprintln(cli.Err(), "Error: failed to restore terminal state to echo input:", err)
			}
		}()

		out := tui.NewOutput(cli.Err())
		out.PrintNote("A Personal Access Token (PAT) can be used instead.\n" +
			"To create a PAT, visit " + aec.Underline.Apply("https://app.docker.com/settings") + "\n\n")
		argPassword, err = prompt.ReadInput(ctx, cli.In(), cli.Out(), "Password: ")
		if err != nil {
			return registrytypes.AuthConfig{}, err
		}
		_, _ = fmt.Fprintln(cli.Out())
		if argPassword == "" {
			return registrytypes.AuthConfig{}, errors.Errorf("Error: Password Required")
		}
	}

	return registrytypes.AuthConfig{
		Username:      argUser,
		Password:      argPassword,
		ServerAddress: serverAddress,
	}, nil
}

// RetrieveAuthTokenFromImage retrieves an encoded auth token given a complete
// image. The auth configuration is serialized as a base64url encoded RFC4648,
// section 5) JSON string for sending through the X-Registry-Auth header.
//
// For details on base64url encoding, see:
// - RFC4648, section 5:   https://tools.ietf.org/html/rfc4648#section-5
//
// FIXME(thaJeztah): do we need a variant like this, but with "indexServer" (domainName) as input?
func RetrieveAuthTokenFromImage(cfg *configfile.ConfigFile, image string) (string, error) {
	// Retrieve encoded auth token from the image reference
	authConfig, err := resolveAuthConfigFromImage(cfg, image)
	if err != nil {
		return "", err
	}
	// FIXME(thaJeztah); implement stringer or "MarshalText / UnmarshalText" on AuthConfig, so that we can pass it around as-is, and encode where it needs to be encoded (when making requests).
	encodedAuth, err := registrytypes.EncodeAuthConfig(authConfig)
	if err != nil {
		return "", err
	}
	return encodedAuth, nil
}

// resolveAuthConfigFromImage retrieves that AuthConfig using the image string
//
// TODO(thaJeztah): export this and/or accept an image ref-type, and use instead of ResolveAuthConfig
func resolveAuthConfigFromImage(cfg *configfile.ConfigFile, image string) (registrytypes.AuthConfig, error) {
	registryRef, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return registrytypes.AuthConfig{}, err
	}
	configKey := getAuthConfigKey(reference.Domain(registryRef))
	a, err := cfg.GetAuthConfig(configKey)
	if err != nil {
		return registrytypes.AuthConfig{}, err
	}
	return registrytypes.AuthConfig(a), nil
}

// getAuthConfigKey special-cases using the full index address of the official
// index as the AuthConfig key, and uses the (host)name[:port] for private indexes.
//
// It is similar to [registry.GetAuthConfigKey], but does not require on
// [registrytypes.IndexInfo] as intermediate.
//
// [registry.GetAuthConfigKey]: https://pkg.go.dev/github.com/docker/docker/registry#GetAuthConfigKey
// [registrytypes.IndexInfo]:https://pkg.go.dev/github.com/docker/docker/api/types/registry#IndexInfo
func getAuthConfigKey(domainName string) string {
	if domainName == "docker.io" || domainName == "index.docker.io" {
		return authConfigKey
	}
	return domainName
}
