package command

import (
	"io"
	"os"
	"strconv"

	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/context/docker"
	"github.com/docker/cli/cli/context/store"
	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/client"
	"github.com/moby/term"
	"github.com/pkg/errors"
)

// DockerCliOption applies a modification on a DockerCli.
type DockerCliOption func(cli *DockerCli) error

// InitializeOpt is the type of the functional options passed to DockerCli.Initialize
// TODO combine InitializeOpt and DockerCliOption, as they have the same signature
type InitializeOpt func(dockerCli *DockerCli) error

// WithStandardStreams sets a cli in, out and err streams with the standard streams.
func WithStandardStreams() DockerCliOption {
	return func(cli *DockerCli) error {
		// Set terminal emulation based on platform as required.
		stdin, stdout, stderr := term.StdStreams()
		cli.in = streams.NewIn(stdin)
		cli.out = streams.NewOut(stdout)
		cli.err = stderr
		return nil
	}
}

// WithCombinedStreams uses the same stream for the output and error streams.
func WithCombinedStreams(combined io.Writer) DockerCliOption {
	return func(cli *DockerCli) error {
		cli.out = streams.NewOut(combined)
		cli.err = combined
		return nil
	}
}

// WithInputStream sets a cli input stream.
func WithInputStream(in io.ReadCloser) DockerCliOption {
	return func(cli *DockerCli) error {
		cli.in = streams.NewIn(in)
		return nil
	}
}

// WithOutputStream sets a cli output stream.
func WithOutputStream(out io.Writer) DockerCliOption {
	return func(cli *DockerCli) error {
		cli.out = streams.NewOut(out)
		return nil
	}
}

// WithErrorStream sets a cli error stream.
func WithErrorStream(err io.Writer) DockerCliOption {
	return func(cli *DockerCli) error {
		cli.err = err
		return nil
	}
}

// WithContentTrustFromEnv enables content trust on a cli from environment variable DOCKER_CONTENT_TRUST value.
func WithContentTrustFromEnv() DockerCliOption {
	return func(cli *DockerCli) error {
		cli.contentTrust = false
		if e := os.Getenv("DOCKER_CONTENT_TRUST"); e != "" {
			if t, err := strconv.ParseBool(e); t || err != nil {
				// treat any other value as true
				cli.contentTrust = true
			}
		}
		return nil
	}
}

// WithContentTrust enables content trust on a cli.
func WithContentTrust(enabled bool) DockerCliOption {
	return func(cli *DockerCli) error {
		cli.contentTrust = enabled
		return nil
	}
}

// WithContextEndpointType add support for an additional typed endpoint in the context store
// Plugins should use this to store additional endpoints configuration in the context store
func WithContextEndpointType(endpointName string, endpointType store.TypeGetter) DockerCliOption {
	return func(cli *DockerCli) error {
		switch endpointName {
		case docker.DockerEndpoint:
			return errors.Errorf("cannot change %q endpoint type", endpointName)
		}
		cli.contextStoreConfig.SetEndpoint(endpointName, endpointType)
		return nil
	}
}

// WithDefaultContextStoreConfig configures the cli to use the default context store configuration.
func WithDefaultContextStoreConfig() DockerCliOption {
	return func(cli *DockerCli) error {
		cli.contextStoreConfig = DefaultContextStoreConfig()
		return nil
	}
}

// WithInitializeClient is passed to DockerCli.Initialize by callers who wish to set a particular API Client for use by the CLI.
func WithInitializeClient(makeClient func(dockerCli *DockerCli) (client.APIClient, error)) InitializeOpt {
	return func(dockerCli *DockerCli) error {
		var err error
		dockerCli.client, err = makeClient(dockerCli)
		return err
	}
}

// WithAPIClientFromEndpoint initializes the CLI's API client if no client has
// been set yet. The API client is based on the CLI's current configuration and
// the active endpoint, or returns an error  if no configuration is loaded. If
// the CLI already has an API client set, this function is a no-op.
func WithAPIClientFromEndpoint() InitializeOpt {
	return func(cli *DockerCli) error {
		if cli.client != nil {
			return nil
		}
		if cli.configFile == nil {
			return errors.New("failed to initialize API client, because no config file is loaded")
		}

		var err error
		cli.client, err = newAPIClientFromEndpoint(cli.dockerEndpoint, cli.configFile)
		return err
	}
}

// WithDefaultConfigFile loads the default configuration, and sets the credential
// store, if no store has been set yet.
func WithDefaultConfigFile() InitializeOpt {
	return func(cli *DockerCli) error {
		// TODO LoadDefaultConfigFile should return an error instead of writing it as warning
		cli.configFile = cliconfig.LoadDefaultConfigFile(cli.err)
		return nil
	}
}
