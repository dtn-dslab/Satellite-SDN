package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sdnctl [COMMANDS]",
	Short: "A generator for Satellite SDN Applications",
	Long: `Sdnctl is a CLI interface for Satellite SDN applications.`,
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
