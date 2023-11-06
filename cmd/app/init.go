package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"ws/dtn-satellite-sdn/sdn"
	"ws/dtn-satellite-sdn/sdn/util"
)
  

var (
	url string
	node int
	interval int
	is_test bool
	version string

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Init a Satellite Network emulation environment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			startTimer := time.Now()

			switch version {
			case "v1":
				if err := sdn.RunSatelliteSDN(url, node, interval); err != nil {
					return fmt.Errorf("Init emulation environment failed: %v\n", err)
				}
			case "v2":
				if err := sdn.RunSDNServer(url, node, interval); err != nil {
					return fmt.Errorf("Init emulation environment failed: %v\n", err)
				}
			}

			// fmt.Printf("%v %v %v %v\n", tle, node, interval, is_test)

			if is_test {
				test_result, err := util.InitEnvTimeCounter(startTimer)
				if err != nil {
					return fmt.Errorf("Count time failed: %v\n", err)
				}
				fmt.Printf("Init Emulation Env Lasts For: %vs\n", test_result)
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