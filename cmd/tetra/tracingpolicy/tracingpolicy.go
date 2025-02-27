// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Hubble

package tracingpolicy

import (
	"fmt"
	"os"
	"strings"

	"github.com/cilium/tetragon/api/v1/tetragon"
	"github.com/cilium/tetragon/cmd/tetra/common"
	"github.com/cilium/tetragon/cmd/tetra/tracingpolicy/generate"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	tpCmd := &cobra.Command{
		Use:     "tracingpolicy",
		Aliases: []string{"tp"},
		Short:   "Manage tracing policies",
	}

	tpAddCmd := &cobra.Command{
		Use:   "add <yaml_file>",
		Short: "add a new sensor based on a tracing policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := common.NewConnectedClient()
			defer c.Close()

			yamlb, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read yaml file %s: %w", args[0], err)
			}

			_, err = c.Client.AddTracingPolicy(c.Ctx, &tetragon.AddTracingPolicyRequest{
				Yaml: string(yamlb),
			})
			if err != nil {
				return fmt.Errorf("failed to add tracing policy: %w", err)
			}
			cmd.Printf("tracing policy %q added\n", args[0])

			return nil
		},
	}

	tpDelCmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "delete a tracing policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := common.NewConnectedClient()
			defer c.Close()

			_, err := c.Client.DeleteTracingPolicy(c.Ctx, &tetragon.DeleteTracingPolicyRequest{
				Name: args[0],
			})
			if err != nil {
				return fmt.Errorf("failed to delete tracing policy: %w", err)
			}
			cmd.Printf("tracing policy %q deleted\n", args[0])

			return nil
		},
	}

	tpEnableCmd := &cobra.Command{
		Use:   "enable <name>",
		Short: "enable a tracing policy",
		Long:  "Enable a disabled tracing policy. Use disable to re-disable the tracing policy.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := common.NewConnectedClient()
			defer c.Close()

			_, err := c.Client.EnableTracingPolicy(c.Ctx, &tetragon.EnableTracingPolicyRequest{
				Name: args[0],
			})
			if err != nil {
				return fmt.Errorf("failed to enable tracing policy: %w", err)
			}
			cmd.Printf("tracing policy %q enabled\n", args[0])

			return nil
		},
	}

	tpDisableCmd := &cobra.Command{
		Use:   "disable <name>",
		Short: "disable a tracing policy",
		Long:  "Disable an enabled tracing policy. Use enable to re-enable the tracing policy.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := common.NewConnectedClient()
			defer c.Close()

			_, err := c.Client.DisableTracingPolicy(c.Ctx, &tetragon.DisableTracingPolicyRequest{
				Name: args[0],
			})

			if err != nil {
				return fmt.Errorf("failed to disable tracing policy: %w", err)
			}

			cmd.Printf("tracing policy %q disabled\n", args[0])

			return nil
		},
	}

	var tpListOutputFlag string
	tpListCmd := &cobra.Command{
		Use:   "list",
		Short: "list tracing policies",
		Args:  cobra.ExactArgs(0),
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if tpListOutputFlag != "json" && tpListOutputFlag != "text" {
				return fmt.Errorf("invalid value for %q flag: %s", common.KeyOutput, tpListOutputFlag)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			c := common.NewConnectedClient()
			defer c.Close()

			res, err := c.Client.ListTracingPolicies(c.Ctx, &tetragon.ListTracingPoliciesRequest{})
			if err != nil || res == nil {
				return fmt.Errorf("failed to list tracing policies: %w", err)
			}

			if tpListOutputFlag == "json" {
				b, err := res.MarshalJSON()
				if err != nil {
					return fmt.Errorf("failed to generate json: %w", err)
				}
				cmd.Println(string(b))
			} else {
				for _, pol := range res.Policies {
					namespace := pol.Namespace
					if namespace == "" {
						namespace = "(global)"
					}

					sensors := strings.Join(pol.Sensors, ",")

					// From v0.11 and before, enabled, filterID and error were
					// bundled in a string. To have a retro-compatible tetra
					// command, we scan the string. If the scan fails, it means
					// something else might be in Info and we print it.
					//
					// we can drop the following block (and comment) when we
					// feel tetra should support only version after v0.11
					if pol.Info != "" {
						var parsedEnabled bool
						var parsedFilterID uint64
						var parsedError string
						var parsedName string
						str := strings.NewReader(pol.Info)
						_, err := fmt.Fscanf(str, "%253s enabled:%t filterID:%d error:%512s", &parsedName, &parsedEnabled, &parsedFilterID, &parsedError)
						if err == nil {
							pol.Enabled = parsedEnabled
							pol.FilterId = parsedFilterID
							pol.Error = parsedError
							pol.Info = ""
						}
					}

					cmd.Printf("[%d] %s enabled:%t filterID:%d namespace:%s sensors:%s\n", pol.Id, pol.Name, pol.Enabled, pol.FilterId, namespace, sensors)
					if pol.Info != "" {
						cmd.Printf("\tinfo: %s\n", pol.Info)
					}
					if pol.Error != "" && pol.Error != "<nil>" {
						cmd.Printf("\terror: %s\n", pol.Error)
					}
				}
			}

			return nil
		},
	}
	tpListFlags := tpListCmd.Flags()
	tpListFlags.StringVarP(&tpListOutputFlag, common.KeyOutput, "o", "text", "Output format. text or json")

	tpCmd.AddCommand(
		tpAddCmd,
		tpDelCmd,
		tpEnableCmd,
		tpDisableCmd,
		tpListCmd,
		generate.New(),
	)

	return tpCmd
}
