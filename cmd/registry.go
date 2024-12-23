package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xpdemon/ac-deploy/config"
)

// Add a Docker registry
var AddRegistryCmd = &cobra.Command{
	Use:   "add-registry",
	Short: "Add a Docker registry to the list",
	Run: func(cmd *cobra.Command, args []string) {
		registry := readLine("URL/Host of the registry (e.g., docker.io/myuser): ")
		config.Cfg.DockerRegistries = append(config.Cfg.DockerRegistries, registry)
		// Save
		err := config.SaveConfig()
		if err != nil {
			fmt.Printf("Error during save: %v\n", err)
		} else {
			fmt.Printf("Registry '%s' added.\n", registry)
		}
	},
}

// Log in to a Docker registry
var LoginRegistryCmd = &cobra.Command{
	Use:   "login-registry",
	Short: "Log in to an existing Docker registry",
	Run: func(cmd *cobra.Command, args []string) {
		if len(config.Cfg.DockerRegistries) == 0 {
			fmt.Println("No registries are registered. Use `xpdemon-deploy add-registry`.")
			return
		}

		// Display available registries
		fmt.Println("Available registries:")
		for i, r := range config.Cfg.DockerRegistries {
			fmt.Printf("  [%d] %s\n", i, r)
		}

		idx := readLine("Choose the index of the registry to log in to: ")
		selectedIndex := strToInt(idx)
		if selectedIndex < 0 || selectedIndex >= len(config.Cfg.DockerRegistries) {
			fmt.Println("Invalid index.")
			return
		}

		registry := config.Cfg.DockerRegistries[selectedIndex]
		err := dockerLogin(registry)
		if err != nil {
			fmt.Printf("Error logging in to '%s': %v\n", registry, err)
			return
		}
		fmt.Printf("Logged in to registry: %s\n", registry)
	},
}
