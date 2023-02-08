package main

import (
	"github.com/SDkie/rollinghash/pkg/signature"
	"github.com/spf13/cobra"
)

var signatureCmd = &cobra.Command{
	Use:   "signature",
	Short: "Generate signature for given file",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		signature.GenerateSignature(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(signatureCmd)
}
