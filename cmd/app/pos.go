package cmd

import (
	"math"
	"ws/dtn-satellite-sdn/position"

	"github.com/spf13/cobra"
)

var (
	tle       string
	fixedNum  int
	maxNum int

	posCmd = &cobra.Command{
		Use:   "pos",
		Short: "Start position computing module.",
		RunE: func(cmd *cobra.Command, args []string) error {
			position.RunPositionModule(tle, fixedNum, maxNum)
			return nil
		},
	}
)

func init() {
	posCmd.Flags().StringVarP(&tle, "tle", "t", "", "The TLE file's path")
	posCmd.Flags().IntVarP(&fixedNum, "num", "n", 10, "The number of fixed network node.")
	posCmd.Flags().IntVar(&maxNum, "max", math.MaxInt32, "The max number of satellites")

	posCmd.MarkFlagRequired("tle")
	posCmd.MarkFlagRequired("num")

	rootCmd.AddCommand(posCmd)
}
