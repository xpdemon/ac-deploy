package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xpdemon/ac-deploy/compose"
	"github.com/xpdemon/ac-deploy/config"
)

var RunFlowCmd = &cobra.Command{
	Use:   "run-flow",
	Short: "Exécute le flux complet : choisir contextes, builder, pousser, déployer",
	Run: func(cmd *cobra.Command, args []string) {
		// 1) Vérifier s'il y a des contextes
		if len(config.Cfg.DockerContexts) == 0 {
			fmt.Println("Aucun contexte Docker n'est enregistré. Utilisez `xpdemon-deploy add-context`.")
			return
		}

		// 2) Afficher la liste des contextes disponibles
		fmt.Println("Liste des contextes Docker :")
		for i, c := range config.Cfg.DockerContexts {
			fmt.Printf("  [%d] %s (host=%s)\n", i, c.Name, c.Host)
		}

		// 3) Sélection du contexte pour BUILDER
		buildIdxInput := readLine("Choisir l'index du contexte pour BUILDER : ")
		buildIdx := strToInt(buildIdxInput)
		if buildIdx < 0 || buildIdx >= len(config.Cfg.DockerContexts) {
			fmt.Println("Index invalide pour le build context.")
			return
		}
		buildContext := config.Cfg.DockerContexts[buildIdx]

		// 4) Sélection du contexte pour DÉPLOYER
		deployIdxInput := readLine("Choisir l'index du contexte pour DÉPLOYER (no-build) : ")
		deployIdx := strToInt(deployIdxInput)
		if deployIdx < 0 || deployIdx >= len(config.Cfg.DockerContexts) {
			fmt.Println("Index invalide pour le deploy context.")
			return
		}
		deployContext := config.Cfg.DockerContexts[deployIdx]

		// 5) Sélection de la registry (facultatif)
		var registry string
		if len(config.Cfg.DockerRegistries) > 0 {
			fmt.Println("Registries disponibles :")
			for i, r := range config.Cfg.DockerRegistries {
				fmt.Printf("  [%d] %s\n", i, r)
			}
			regIdxInput := readLine("Choisir l'index de la registry pour push (ou ENTER pour ignorer) : ")
			if regIdxInput != "" {
				regIdx := strToInt(regIdxInput)
				if regIdx >= 0 && regIdx < len(config.Cfg.DockerRegistries) {
					registry = config.Cfg.DockerRegistries[regIdx]
				}
			}
		}

		// 6) Chemin du docker-compose.yml
		composeFile := readLine("Chemin vers votre docker-compose.yml : ")
		if composeFile == "" {
			fmt.Println("Chemin du docker-compose.yml non spécifié, annulation.")
			return
		}

		// 7) Parser le docker-compose.yml pour détecter les images
		images, err := compose.ParseComposeFile(composeFile)
		if err != nil {
			fmt.Printf("Erreur lors du parsing du docker-compose : %v\n", err)
			return
		}
		fmt.Println("Images détectées dans ce docker-compose :")
		for _, img := range images {
			fmt.Printf("  - %s\n", img)
		}

		// 7.a) Étape optionnelle : Prune avant le build
		pruneChoice := readLine("Voulez-vous faire du ménage (prune) sur le contexte de build ? (o/n) : ")
		if strings.ToLower(pruneChoice) == "o" {
			// On peut découper en deux questions si on veut un contrôle précis :
			pruneImagesChoice := readLine("   > Supprimer les images Docker inutilisées (docker images prune -a) ? (o/n) : ")
			if strings.ToLower(pruneImagesChoice) == "o" {
				fmt.Println("==> Exécution de docker image prune...")
				err = runCommand(
					"docker",
					"--context", buildContext.Name,
					"image", "prune",
					"-a",
					"-f", // pour forcer sans demander de confirmation
				)
				if err != nil {
					fmt.Printf("Erreur lors du prune des images : %v\n", err)
					// On peut décider de continuer ou de quitter
				}
			}

			pruneBuilderChoice := readLine("   > Supprimer le cache builder Docker (docker builder prune) ? (o/n) : ")
			if strings.ToLower(pruneBuilderChoice) == "o" {
				fmt.Println("==> Exécution de docker builder prune...")
				err = runCommand(
					"docker",
					"--context", buildContext.Name,
					"builder", "prune",
					"-f",
				)
				if err != nil {
					fmt.Printf("Erreur lors du prune du builder : %v\n", err)
					// On peut décider de continuer ou de quitter
				}
			}
		}

		// 8) Build
		fmt.Println("==> Build des images...")
		err = runCommand(
			"docker",
			"--context", buildContext.Name,
			"compose",
			"-f", composeFile,
			"build",
		)
		if err != nil {
			fmt.Printf("Erreur lors du build : %v\n", err)
			return
		}

		// 9) Push
		pushChoice := readLine("Voulez-vous pousser les images sur la registry sélectionnée ? (o/n) : ")
		if strings.ToLower(pushChoice) == "o" && registry != "" {
			fmt.Println("==> Push des images...")
			err = runCommand(
				"docker",
				"--context", buildContext.Name,
				"compose",
				"-f", composeFile,
				"push",
			)
			if err != nil {
				fmt.Printf("Erreur lors du push : %v\n", err)
				return
			}
		}

		// 10) Déployer
		deployChoice := readLine("Voulez-vous déployer les images en no-build ? (o/n) : ")
		if strings.ToLower(deployChoice) == "o" {
			fmt.Println("==> Déploiement en mode no-build...")
			err = runCommand(
				"docker",
				"--context", deployContext.Name,
				"compose",
				"-f", composeFile,
				"up",
				"-d",
				"--no-build",
			)
			if err != nil {
				fmt.Printf("Erreur lors du déploiement : %v\n", err)
				return
			}
			fmt.Println("Déploiement effectué avec succès !")
		} else {
			fmt.Println("Déploiement annulé.")
		}
	},
}
