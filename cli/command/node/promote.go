package node

import (
	"context"
	"fmt"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types/swarm"
	"github.com/spf13/cobra"
)

func newPromoteCommand(dockerCli command.Cli) *cobra.Command {
	return &cobra.Command{
		Use:   "promote NODE [NODE...]",
		Short: "Promote one or more nodes to manager in the swarm",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPromote(dockerCli, args)
		},
	}
}

func runPromote(dockerCli command.Cli, nodes []string) error {
	promote := func(node *swarm.Node) error {
		if node.Spec.Role == swarm.NodeRoleManager {
			fmt.Fprintf(dockerCli.Out(), "Node %s is already a manager.\n", node.ID)
			return errNoRoleChange
		}
		node.Spec.Role = swarm.NodeRoleManager
		return nil
	}
	success := func(nodeID string) {
		fmt.Fprintf(dockerCli.Out(), "Node %s promoted to a manager in the swarm.\n", nodeID)
	}
	if err := updateNodes(dockerCli, nodes, promote, success); err != nil {
		return err
	}

	client := dockerCli.Client()
	ctx := context.Background()
	info, err := client.Info(ctx)
	if err != nil {
		return err
	}

	if command.GetManagerCount(info.Swarm) == 2 {
		command.PrintManagerWarning(dockerCli)
	}
	return nil

}
