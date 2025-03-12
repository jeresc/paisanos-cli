package cmd

import (
	"fmt"
	"os"
)

// Execute runs the Cobra command.
func Execute() {
	if err := SetupCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
