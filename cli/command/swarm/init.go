package swarm

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types/swarm"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type initOptions struct {
	swarmOptions
	listenAddr NodeAddrOption
	// Not a NodeAddrOption because it has no default port.
	advertiseAddr   string
	dataPathAddr    string
	forceNewCluster bool
	availability    string
	defaultAddrPool string
}

func newInitCommand(dockerCli command.Cli) *cobra.Command {
	opts := initOptions{
		listenAddr: NewListenAddrOption(),
	}

	cmd := &cobra.Command{
		Use:   "init [OPTIONS]",
		Short: "Initialize a swarm",
		Args:  cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(dockerCli, cmd.Flags(), opts)
		},
	}

	flags := cmd.Flags()
	flags.Var(&opts.listenAddr, flagListenAddr, "Listen address (format: <ip|interface>[:port])")
	flags.StringVar(&opts.advertiseAddr, flagAdvertiseAddr, "", "Advertised address (format: <ip|interface>[:port])")
	flags.StringVar(&opts.dataPathAddr, flagDataPathAddr, "", "Address or interface to use for data path traffic (format: <ip|interface>)")
	flags.SetAnnotation(flagDataPathAddr, "version", []string{"1.31"})
	flags.BoolVar(&opts.forceNewCluster, "force-new-cluster", false, "Force create a new cluster from current state")
	flags.BoolVar(&opts.autolock, flagAutolock, false, "Enable manager autolocking (requiring an unlock key to start a stopped manager)")
	flags.StringVar(&opts.availability, flagAvailability, "active", `Availability of the node ("active"|"pause"|"drain")`)
	flags.StringVar(&opts.defaultAddrPool, flagDefaultAddrPool, "", "List of default subnet addresses followed by subnet size (format: <subnet,subnet,..:subnet-size)")
	flags.SetAnnotation(flagDefaultAddrPool, "version", []string{"1.39"})
	addSwarmFlags(flags, &opts.swarmOptions)
	return cmd
}

// getDefaultAddrPool extracts info from address pool string and fills PoolsOpt struct
func getDefaultAddrPool(defaultAddrPool string) ([]string, int, error) {
	var (
		size int
		err  error
	)
	if defaultAddrPool == "" {
		// defaultAddrPool is not defined
		return nil, 0, nil
	}

	result := strings.Split(defaultAddrPool, ":")
	if len(result) > 2 {
		return nil, 0, fmt.Errorf("Invalid default address pool format. Expected format CIDR[,CIDR]*:SUBNET-SIZE")
	}
	// if size is not specified default size is 24
	size = 24
	if len(result) == 2 {
		// trim leading and trailing white spaces
		result[1] = strings.TrimSpace(result[1])

		// get the size from the slice
		size, err = strconv.Atoi(result[1])

		if err != nil || size <= 0 {
			return nil, 0, fmt.Errorf("error in DefaultAddressPool subnet size %s", defaultAddrPool)
		}
	}

	// get subnet list
	subnetlist := strings.Split(result[0], ",")

	for i := range subnetlist {
		// trim leading and trailing white spaces
		subnetlist[i] = strings.TrimSpace(subnetlist[i])
		_, b, err := net.ParseCIDR(subnetlist[i])
		if err != nil {
			return nil, 0, fmt.Errorf("invalid base pool %q: %v", subnetlist[i], err)
		}
		ones, _ := b.Mask.Size()
		if size < ones {
			return nil, 0, fmt.Errorf("subnet size is too small for pool: %d", size)
		}
	}
	return subnetlist, size, nil
}

func runInit(dockerCli command.Cli, flags *pflag.FlagSet, opts initOptions) error {
	var (
		size            int
		defaultAddrPool []string
		err             error
	)
	client := dockerCli.Client()
	ctx := context.Background()

	if opts.defaultAddrPool != "" {
		defaultAddrPool, size, err = getDefaultAddrPool(opts.defaultAddrPool)
		if err != nil {
			return err
		}
	}

	req := swarm.InitRequest{
		ListenAddr:       opts.listenAddr.String(),
		AdvertiseAddr:    opts.advertiseAddr,
		DataPathAddr:     opts.dataPathAddr,
		DefaultAddrPool:  defaultAddrPool,
		ForceNewCluster:  opts.forceNewCluster,
		Spec:             opts.swarmOptions.ToSpec(flags),
		AutoLockManagers: opts.swarmOptions.autolock,
		SubnetSize:       uint32(size),
	}
	if flags.Changed(flagAvailability) {
		availability := swarm.NodeAvailability(strings.ToLower(opts.availability))
		switch availability {
		case swarm.NodeAvailabilityActive, swarm.NodeAvailabilityPause, swarm.NodeAvailabilityDrain:
			req.Availability = availability
		default:
			return errors.Errorf("invalid availability %q, only active, pause and drain are supported", opts.availability)
		}
	}

	nodeID, err := client.SwarmInit(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "could not choose an IP address to advertise") || strings.Contains(err.Error(), "could not find the system's IP address") {
			return errors.New(err.Error() + " - specify one with --advertise-addr")
		}
		return err
	}

	fmt.Fprintf(dockerCli.Out(), "Swarm initialized: current node (%s) is now a manager.\n\n", nodeID)

	if err := printJoinCommand(ctx, dockerCli, nodeID, true, false); err != nil {
		return err
	}

	fmt.Fprint(dockerCli.Out(), "To add a manager to this swarm, run 'docker swarm join-token manager' and follow the instructions.\n\n")

	if req.AutoLockManagers {
		unlockKeyResp, err := client.SwarmGetUnlockKey(ctx)
		if err != nil {
			return errors.Wrap(err, "could not fetch unlock key")
		}
		printUnlockCommand(dockerCli.Out(), unlockKeyResp.UnlockKey)
	}

	return nil
}
