package main

import (
	"github.com/SDkie/rollinghash/pkg/common"
	"github.com/spf13/cobra"
)

var signatureCmd = &cobra.Command{
	Use:   "signature",
	Short: "Generate signature for given file",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		common.GenerateSignature(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(signatureCmd)
}
