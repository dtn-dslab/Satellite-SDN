package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"ws/dtn-satellite-sdn/sdn"
	"ws/dtn-satellite-sdn/sdn/util"
)
  

var (
	tle string
	node int
	interval int
	is_test bool

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Init a Satellite Network emulation environment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			startTimer := time.Now()

			if err := sdn.RunSatelliteSDN(tle, node, interval); err != nil {
				return fmt.Errorf("Init emulation environment failed: %v\n", err)
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
	initCmd.Flags().StringVarP(&tle, "tle", "t", "", "TLE file's path to read from")
	initCmd.Flags().IntVarP(&node, "node", "n", 3, "Expected node num")
	initCmd.Flags().IntVarP(&interval, "interval", "i", -1, "Assign update interval for Satellite SDN Controller (-1 means 'no update')")
	initCmd.Flags().BoolVar(&is_test, "test", false, "Open the test mode")

	initCmd.MarkFlagRequired("tle")
	initCmd.MarkFlagRequired("node")

	rootCmd.AddCommand(initCmd)
}