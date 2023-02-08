package main

import (
	"github.com/SDkie/rollinghash/pkg/delta"
	"github.com/spf13/cobra"
)

var deltaCmd = &cobra.Command{
	Use:   "delta",
	Short: "Generate delta for given file",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		delta.GenerateDelta(args[0], args[1], args[2], args[3])
	},
}

func init() {
	deltaCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		cmd.Println("Usage: rollinghash delta <original_file> <signature_file> <updated_file> <delta_file>")
		return nil
	})
	rootCmd.AddCommand(deltaCmd)
}
