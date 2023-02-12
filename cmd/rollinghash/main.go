package main

import (
	"log"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "rollinghash",
		Short: "rollinghash is a CLI tool to calculate signature and delta for files using rolling hash algorithm",
	}
	rootCmd.AddCommand(getSignatureCmd(), getDeltaCmd())

	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("error from cmd execution: %s", err)
	}
}
