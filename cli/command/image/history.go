package image

import (
	"context"
	"fmt"

	"github.com/containerd/platforms"
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/completion"
	"github.com/docker/cli/cli/command/formatter"
	flagsHelper "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types/image"
	"github.com/spf13/cobra"
)

type historyOptions struct {
	image    string
	platform string

	human   bool
	quiet   bool
	noTrunc bool
	format  string
}

// NewHistoryCommand creates a new `docker history` command
func NewHistoryCommand(dockerCli command.Cli) *cobra.Command {
	var opts historyOptions

	cmd := &cobra.Command{
		Use:   "history [OPTIONS] IMAGE",
		Short: "Show the history of an image",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.image = args[0]
			return runHistory(cmd.Context(), dockerCli, opts)
		},
		ValidArgsFunction: completion.ImageNames(dockerCli),
		Annotations: map[string]string{
			"aliases": "docker image history, docker history",
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&opts.human, "human", "H", true, "Print sizes and dates in human readable format")
	flags.BoolVarP(&opts.quiet, "quiet", "q", false, "Only show image IDs")
	flags.BoolVar(&opts.noTrunc, "no-trunc", false, "Don't truncate output")
	flags.StringVar(&opts.format, "format", "", flagsHelper.FormatHelp)
	flags.StringVar(&opts.platform, "platform", "",
		`Specify a platform from a multi-platform image to show the history for.
If the platform is not specified, the host platform is preferred if it's available, otherwise any available platform is used.

Format: "os[/arch[/variant]]"
Example: "docker image history --platform=linux/amd64"`)

	return cmd
}

func runHistory(ctx context.Context, dockerCli command.Cli, opts historyOptions) error {
	var options image.HistoryOptions
	if opts.platform != "" {
		p, err := platforms.Parse(opts.platform)
		if err != nil {
			_, _ = fmt.Fprintf(dockerCli.Err(), "Invalid platform %s", opts.platform)
			return err
		}
		options.Platform = &p
	}

	history, err := dockerCli.Client().ImageHistory(ctx, opts.image, options)
	if err != nil {
		return err
	}

	format := opts.format
	if len(format) == 0 {
		format = formatter.TableFormatKey
	}

	historyCtx := formatter.Context{
		Output: dockerCli.Out(),
		Format: NewHistoryFormat(format, opts.quiet, opts.human),
		Trunc:  !opts.noTrunc,
	}
	return HistoryWrite(historyCtx, opts.human, history)
}
