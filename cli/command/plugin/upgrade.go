package plugin

import (
	"context"
	"fmt"
	"strings"

	"github.com/distribution/reference"
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/internal/jsonstream"
	"github.com/docker/cli/internal/prompt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newUpgradeCommand(dockerCli command.Cli) *cobra.Command {
	var options pluginOptions
	cmd := &cobra.Command{
		Use:   "upgrade [OPTIONS] PLUGIN [REMOTE]",
		Short: "Upgrade an existing plugin",
		Args:  cli.RequiresRangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.localName = args[0]
			if len(args) == 2 {
				options.remote = args[1]
			}
			return runUpgrade(cmd.Context(), dockerCli, options)
		},
		Annotations: map[string]string{"version": "1.26"},
	}

	flags := cmd.Flags()
	loadPullFlags(dockerCli, &options, flags)
	flags.BoolVar(&options.skipRemoteCheck, "skip-remote-check", false, "Do not check if specified remote plugin matches existing plugin image")
	return cmd
}

func runUpgrade(ctx context.Context, dockerCLI command.Cli, opts pluginOptions) error {
	p, _, err := dockerCLI.Client().PluginInspectWithRaw(ctx, opts.localName)
	if err != nil {
		return errors.Errorf("error reading plugin data: %v", err)
	}

	if p.Enabled {
		return errors.Errorf("the plugin must be disabled before upgrading")
	}

	opts.localName = p.Name
	if opts.remote == "" {
		opts.remote = p.PluginReference
	}
	remote, err := reference.ParseNormalizedNamed(opts.remote)
	if err != nil {
		return errors.Wrap(err, "error parsing remote upgrade image reference")
	}
	remote = reference.TagNameOnly(remote)

	old, err := reference.ParseNormalizedNamed(p.PluginReference)
	if err != nil {
		return errors.Wrap(err, "error parsing current image reference")
	}
	old = reference.TagNameOnly(old)

	_, _ = fmt.Fprintf(dockerCLI.Out(), "Upgrading plugin %s from %s to %s\n", p.Name, reference.FamiliarString(old), reference.FamiliarString(remote))
	if !opts.skipRemoteCheck && remote.String() != old.String() {
		r, err := prompt.Confirm(ctx, dockerCLI.In(), dockerCLI.Out(), "Plugin images do not match, are you sure?")
		if err != nil {
			return err
		}
		if !r {
			return cancelledErr{errors.New("plugin upgrade has been cancelled")}
		}
	}

	options, err := buildPullConfig(ctx, dockerCLI, opts)
	if err != nil {
		return err
	}

	responseBody, err := dockerCLI.Client().PluginUpgrade(ctx, opts.localName, options)
	if err != nil {
		if strings.Contains(err.Error(), "target is image") {
			return errors.New(err.Error() + " - Use `docker image pull`")
		}
		return err
	}
	defer func() {
		_ = responseBody.Close()
	}()
	if err := jsonstream.Display(ctx, responseBody, dockerCLI.Out()); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(dockerCLI.Out(), "Upgraded plugin %s to %s\n", opts.localName, opts.remote) // todo: return proper values from the API for this result
	return nil
}

type cancelledErr struct{ error }

func (cancelledErr) Cancelled() {}
