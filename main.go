package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/xpdemon/ac-deploy/cmd"
	"github.com/xpdemon/ac-deploy/config"
)

func main() {
	// Load the configuration
	err := config.LoadConfig()
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		if err != nil {
			return
		}
		// Continue with an empty config or exit, depending on your logic
	}

	// Docker verification
	err = cmd.CheckDockerInstalled()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Docker verification failed: %v\n", err)
		os.Exit(1)
	}

	// Root command
	rootCmd := &cobra.Command{
		Use:   "xpdemon-deploy",
		Short: "Docker Deployment CLI (advanced example)",
		Long:  "An example Go CLI with Cobra to build and deploy Docker Compose images across different contexts and registries.",
	}

	// Add sub-commands
	rootCmd.AddCommand(
		cmd.AddContextCmd,
		cmd.ListContextsCmd, // <-- Added here
		cmd.AddRegistryCmd,
		cmd.LoginRegistryCmd,
		cmd.RunFlowCmd,
	)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
