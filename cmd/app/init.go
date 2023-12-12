package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"ws/dtn-satellite-sdn/sdn"
)

var (
	url      string
	node     int
	interval int
	is_test  bool
	version  string

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Init a Satellite Network emulation environment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch version {
			case "v1":
				if err := sdn.RunSatelliteSDN(url, node, interval); err != nil {
					return fmt.Errorf("init emulation environment failed: %v", err)
				}
			case "v2":
				if is_test {
					if err := sdn.RunSDNServerTest(url, node, interval); err != nil {
						return fmt.Errorf("init test emulation environment failed: %v", err)
					}

				} else {
					if err := sdn.RunSDNServer(url, node, interval); err != nil {
						return fmt.Errorf("init emulation environment failed: %v", err)
					}
				}
			}

			return nil
		},
	}
)

func init() {
	initCmd.Flags().StringVarP(&url, "url", "u", "", "v1: TLE file's path to read from / v2: The address of Position Calculation Module")
	initCmd.Flags().IntVarP(&node, "node", "n", 3, "Expected node num")
	initCmd.Flags().IntVarP(&interval, "interval", "i", -1, "Assign update interval for Satellite SDN Controller (-1 means 'no update')")
	initCmd.Flags().BoolVar(&is_test, "test", false, "Open the test mode")
	initCmd.Flags().StringVarP(&version, "version", "v", "", "SDN server's version")

	initCmd.MarkFlagRequired("url")
	initCmd.MarkFlagRequired("node")
	initCmd.MarkFlagRequired("version")

	rootCmd.AddCommand(initCmd)
}
