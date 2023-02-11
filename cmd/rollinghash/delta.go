package main

import (
	"github.com/SDkie/rollinghash/pkg/delta"
	"github.com/spf13/cobra"
)

func getDeltaCmd() *cobra.Command {
	deltaCmd := &cobra.Command{
		Use:   "delta",
		Short: "Generate delta between original and updated file",
		Args:  cobra.ExactArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			delta.GenerateDelta(args[0], args[1], args[2], args[3])
		},
	}

	deltaCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		cmd.Println("Usage: rollinghash delta <original_file> <signature_file> <updated_file> <delta_file>")
		return nil
	})

	return deltaCmd
}
