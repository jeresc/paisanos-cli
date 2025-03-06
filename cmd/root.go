package cmd

import (
	"fmt"
	"os"
)

// Execute runs the Cobra command.
func Execute() {
	if err := WelcomeCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := SetupCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
