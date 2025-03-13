package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "paisanos-cli",
	Short: "A program for quick macOS environment setup made by paisanos.io",
	Long:  `paisanos-cli is a CLI tool for setting up your macOS environment. It is designed to be easy to use and quick to set up.`,
}

// Execute runs the Cobra command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
