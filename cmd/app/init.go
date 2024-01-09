package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"ws/dtn-satellite-sdn/sdn"
)

var (
	url      string
	node     int
	interval int
	is_test  bool
	is_debug bool

	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Init a Satellite Network emulation environment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
			if is_debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
			if is_test {
				if err := sdn.RunSDNServerTest(url, node, interval); err != nil {
					return fmt.Errorf("init test emulation environment failed: %v", err)
				}
			} else {
				if err := sdn.RunSDNServer(url, node, interval); err != nil {
					return fmt.Errorf("init emulation environment failed: %v", err)
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
	initCmd.Flags().BoolVar(&is_debug, "debug", false, "Open the debug mode")

	initCmd.MarkFlagRequired("url")
	initCmd.MarkFlagRequired("node")

	rootCmd.AddCommand(initCmd)
}
