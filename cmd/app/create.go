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
		Use:   "create",
		Short: "Create the flows in the sdn.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if resource == "flow" {
				if err := flow.CreateFlows(bandwidth); err != nil {
					return fmt.Errorf("Create flows failed: %v\n", err)
				}
			}
			return nil
		},
	}
)

func init() {
	createCmd.Flags().StringVarP(&resource, "resource", "r", "flow", "the flow resource for sdn to create")
	createCmd.Flags().StringVarP(&bandwidth, "bandwidth", "b", "100M", "the bandwidth of each flow pair")

	rootCmd.AddCommand(createCmd)
}
