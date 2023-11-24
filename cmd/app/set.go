package cmd

import (
	"fmt"
	"ws/dtn-satellite-sdn/flow"

	"github.com/spf13/cobra"
)

var (
	resource  string
	bandwidth string
	createCmd = &cobra.Command{
		Use:   "set",
		Short: "Set the flows in the sdn.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if resource == "server" {
				if err := flow.StartServer(); err != nil {
					return fmt.Errorf("Set flows failed: %v\n", err)
				}
			} else if resource == "client" {
				if err := flow.StartClient(bandwidth); err != nil {
					return fmt.Errorf("Set flows failed: %v\n", err)
				}
			}
			return nil
		},
	}
)

func init() {
	createCmd.Flags().StringVarP(&resource, "resource", "r", "flow", "the flow resource for sdn to create,such as client and server")
	createCmd.Flags().StringVarP(&bandwidth, "bandwidth", "b", "100M", "the bandwidth of each flow pair")

	rootCmd.AddCommand(createCmd)
}
