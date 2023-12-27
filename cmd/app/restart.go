package cmd

import (
	"ws/dtn-satellite-sdn/sdn"

	"github.com/spf13/cobra"
)

var (
	script	string

	restartCmd = &cobra.Command{
		Use:   "restart",
		Short: "Start restart server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return sdn.RunRestartTestServer(script)
		},
	}
)

func init() {
	restartCmd.Flags().StringVarP(&script, "script", "s", "", "specify start script path.")

	restartCmd.MarkFlagRequired("script")

	rootCmd.AddCommand(restartCmd)
}