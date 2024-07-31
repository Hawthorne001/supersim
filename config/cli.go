package config

import (
	"fmt"
	"strings"

	opservice "github.com/ethereum-optimism/optimism/op-service"

	registry "github.com/ethereum-optimism/superchain-registry/superchain"

	"github.com/urfave/cli/v2"
)

const (
	ForkCommandName = "fork"

	L1ForkHeightFlagName = "l1.fork.height"
	L1PortFlagName       = "l1.port"

	ChainsFlagName         = "chains"
	NetworkFlagName        = "network"
	L2StartingPortFlagName = "l2.starting.port"
)

func BaseCLIFlags(envPrefix string) []cli.Flag {
	return []cli.Flag{
		&cli.Uint64Flag{
			Name:    L1PortFlagName,
			Usage:   "Listening port for the L1 instance. `0` binds to any available port",
			Value:   8545,
			EnvVars: opservice.PrefixEnvVar(envPrefix, "L1_PORT"),
		},
		&cli.Uint64Flag{
			Name:    L2StartingPortFlagName,
			Usage:   "Starting port to increment from for L2 chains. `0` binds each chain to any available port",
			Value:   9545,
			EnvVars: opservice.PrefixEnvVar(envPrefix, "L2_STARTING_PORT"),
		},
	}
}

func ForkCLIFlags(envPrefix string) []cli.Flag {
	networks := strings.Join(superchainNetworks(), ", ")
	mainnetMembers := strings.Join(superchainMemberChains(registry.Superchains["mainnet"]), ", ")
	return []cli.Flag{
		&cli.Uint64Flag{
			Name:    L1ForkHeightFlagName,
			Usage:   "L1 height to fork the superchain (bounds L2 time). `0` for latest",
			Value:   0,
			EnvVars: opservice.PrefixEnvVar(envPrefix, "L1_FORK_HEIGHT"),
		},
		&cli.StringSliceFlag{
			Name:     ChainsFlagName,
			Usage:    fmt.Sprintf("chains to fork in the superchain, mainnet options: [%s]", mainnetMembers),
			Required: true,
			EnvVars:  opservice.PrefixEnvVar(envPrefix, "CHAINS"),
		},
		&cli.StringFlag{
			Name:    NetworkFlagName,
			Value:   "mainnet",
			Usage:   fmt.Sprintf("superchain network. options: %s", networks),
			EnvVars: opservice.PrefixEnvVar(envPrefix, "NETWORK"),
		},
	}
}

type ForkCLIConfig struct {
	L1ForkHeight uint64
	Network      string
	Chains       []string
}

type CLIConfig struct {
	L1Port         uint64
	L2StartingPort uint64

	ForkConfig *ForkCLIConfig
}

func ReadCLIConfig(ctx *cli.Context) (*CLIConfig, error) {
	cfg := &CLIConfig{
		L1Port:         ctx.Uint64(L1PortFlagName),
		L2StartingPort: ctx.Uint64(L2StartingPortFlagName),
	}

	if ctx.Command.Name == ForkCommandName {
		cfg.ForkConfig = &ForkCLIConfig{
			L1ForkHeight: ctx.Uint64(L1ForkHeightFlagName),
			Network:      ctx.String(NetworkFlagName),
			Chains:       ctx.StringSlice(ChainsFlagName),
		}
	}

	return cfg, cfg.Check()
}

// Check runs validatation on the cli configuration
func (c *CLIConfig) Check() error {
	if c.ForkConfig != nil {
		forkCfg := c.ForkConfig
		superchain, ok := registry.Superchains[forkCfg.Network]
		if !ok {
			return fmt.Errorf("unrecognized superchain network `%s`, available networks: [%s]",
				forkCfg.Network, strings.Join(superchainNetworks(), ", "))
		}

		// ensure every chain is apart of the network
		for _, chain := range forkCfg.Chains {
			if !isInSuperchain(chain, superchain) {
				return fmt.Errorf("unrecognized chain `%s` in %s superchain, available chains: [%s]",
					chain, forkCfg.Network, strings.Join(superchainMemberChains(superchain), ", "))
			}
		}
	}

	return nil
}