package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/xpdemon/ac-deploy/cmd"
	"github.com/xpdemon/ac-deploy/config"
)

func main() {
	// Charger la configuration
	err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erreur lors du chargement de la config : %v\n", err)
		// Continuer avec une config vide ou quitter, selon votre logique
	}

	// Vérification Docker
	err = cmd.CheckDockerInstalled()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Vérification Docker échouée : %v\n", err)
		os.Exit(1)
	}

	// Commande racine
	rootCmd := &cobra.Command{
		Use:   "xpdemon-deploy",
		Short: "CLI de déploiement Docker (exemple avancé)",
		Long:  "Un exemple de CLI Go avec Cobra pour builder et déployer des images Docker Compose sur différents contextes et registries.",
	}

	// Ajout des sous-commandes
	rootCmd.AddCommand(
		cmd.AddContextCmd,
		cmd.ListContextsCmd, // <-- On ajoute ici
		cmd.AddRegistryCmd,
		cmd.LoginRegistryCmd,
		cmd.RunFlowCmd,
	)

	// Exécution
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
