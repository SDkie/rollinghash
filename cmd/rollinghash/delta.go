package main

import (
	"github.com/SDkie/rollinghash/pkg/common"
	"github.com/spf13/cobra"
)

var deltaCmd = &cobra.Command{
	Use:   "delta",
	Short: "Generate delta for given file",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		common.GenerateDelta(args[0], args[1], args[2], args[3])
	},
}

func init() {
	rootCmd.AddCommand(deltaCmd)
}
