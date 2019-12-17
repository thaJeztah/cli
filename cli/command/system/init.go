package system

import (
	"os"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"
)

const dockerDir = ".docker"

type initOptions struct {
	template string
}

// NewInitCommand creates a new cobra.Command for `docker init`
func NewInitCommand(dockerCli command.Cli) *cobra.Command {
	var opts initOptions

	cmd := &cobra.Command{
		Use:   "init [OPTIONS]",
		Short: "Initialize a docker project",
		Args:  cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, dockerCli, &opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.template, "template", "t", "default", "Template to use for initializing docker")

	return cmd
}

func runInit(cmd *cobra.Command, dockerCli command.Cli, opts *initOptions) error {
	return os.Mkdir(dockerDir, 0777)
}
