package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xpdemon/ac-deploy/config"
)

// Ajouter une registry
var AddRegistryCmd = &cobra.Command{
	Use:   "add-registry",
	Short: "Ajoute une registry Docker à la liste",
	Run: func(cmd *cobra.Command, args []string) {
		registry := readLine("URL/Host de la registry (ex: docker.io/monuser) : ")
		config.Cfg.DockerRegistries = append(config.Cfg.DockerRegistries, registry)
		// Sauvegarder
		err := config.SaveConfig()
		if err != nil {
			fmt.Printf("Erreur lors de la sauvegarde : %v\n", err)
		} else {
			fmt.Printf("Registry '%s' ajoutée.\n", registry)
		}
	},
}

// Se logguer à la registry
var LoginRegistryCmd = &cobra.Command{
	Use:   "login-registry",
	Short: "Se connecter à une registry existante",
	Run: func(cmd *cobra.Command, args []string) {
		if len(config.Cfg.DockerRegistries) == 0 {
			fmt.Println("Aucune registry n'est enregistrée. Utilisez `xpdemon-deploy add-registry`.")
			return
		}

		// Affiche les registries disponibles
		fmt.Println("Registries disponibles :")
		for i, r := range config.Cfg.DockerRegistries {
			fmt.Printf("  [%d] %s\n", i, r)
		}

		idx := readLine("Choisissez l'index de la registry à laquelle se connecter : ")
		selectedIndex := strToInt(idx)
		if selectedIndex < 0 || selectedIndex >= len(config.Cfg.DockerRegistries) {
			fmt.Println("Index invalide.")
			return
		}

		registry := config.Cfg.DockerRegistries[selectedIndex]
		err := dockerLogin(registry)
		if err != nil {
			fmt.Printf("Erreur lors du login à '%s' : %v\n", registry, err)
			return
		}
		fmt.Printf("Connecté à la registry : %s\n", registry)
	},
}
