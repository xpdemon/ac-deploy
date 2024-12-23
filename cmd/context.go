package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xpdemon/ac-deploy/config"
)

// ListContextsCmd : commande pour lister les contextes Docker existants
var ListContextsCmd = &cobra.Command{
	Use:   "list-contexts",
	Short: "Liste les contextes Docker existants dans la configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if len(config.Cfg.DockerContexts) == 0 {
			fmt.Println("Aucun contexte Docker n'est enregistré.")
			return
		}

		fmt.Println("Contextes Docker enregistrés :")
		for i, ctx := range config.Cfg.DockerContexts {
			// ctx est de type config.DockerContext
			// Affichez ici les informations pertinentes
			// (par ex. Name, Host, Description, etc.)
			fmt.Printf("[%d] %s (host=%s)\n", i, ctx.Name, ctx.Host)
			if ctx.Description != "" {
				fmt.Printf("    Description: %s\n", ctx.Description)
			}
		}
	},
}

// AddContextCmd : commande pour créer un vrai contexte Docker
var AddContextCmd = &cobra.Command{
	Use:   "add-context",
	Short: "Crée un vrai contexte Docker (docker context create) avec des informations avancées",
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Demander le nom du contexte
		contextName := readLine("Nom du nouveau contexte Docker : ")
		if contextName == "" {
			fmt.Println("Le nom de contexte est obligatoire, annulation.")
			return
		}

		// 2. Demander une description éventuelle
		desc := readLine("Description (optionnel) : ")

		// 3. Demander le Docker Host (ex: ssh://user@server ou tcp://192.168.1.10:2376)
		host := readLine("Docker Host (ex: ssh://user@host, tcp://X.X.X.X:2376) : ")
		if host == "" {
			fmt.Println("Le host Docker est obligatoire, annulation.")
			return
		}

		// 4. (Optionnel) Demander si TLS est requis, etc.
		// tlsChoice := readLine("Voulez-vous activer TLS? (o/n) : ")
		// tlsVerify := (strings.ToLower(tlsChoice) == "o")
		// certPath := ""
		// if tlsVerify {
		//     certPath = readLine("Chemin vers le dossier de certificats (ex: /home/user/.docker): ")
		// }

		// 5. Construire la commande docker context create
		// Par exemple:
		// docker context create monContexte
		//     --description "Ma description"
		//     --docker "host=ssh://user@host"
		//     --docker "tlsverify=true"
		//     --docker "cacert=...certs/ca.pem"
		// etc.
		dockerArgs := []string{
			"context", "create",
			contextName,
			"--docker", fmt.Sprintf("host=%s", host),
		}
		// Ajouter la description si fournie
		if desc != "" {
			dockerArgs = append(dockerArgs, "--description", desc)
		}

		fmt.Println("==> Création du contexte via docker CLI...")
		err := runCommand("docker", dockerArgs...)
		if err != nil {
			fmt.Printf("Échec de création du contexte Docker : %v\n", err)
			return
		}
		fmt.Printf("Contexte Docker '%s' créé avec succès.\n", contextName)

		// 6. Stocker ce contexte dans la config pour usage ultérieur
		newCtx := config.DockerContext{
			Name:        contextName,
			Description: desc,
			Host:        host,
			// TLSVerify:   tlsVerify,
			// CertPath:    certPath,
		}
		config.Cfg.DockerContexts = append(config.Cfg.DockerContexts, newCtx)

		// 7. Sauvegarder la config
		err = config.SaveConfig()
		if err != nil {
			fmt.Printf("Erreur lors de la sauvegarde dans ~/.myapp/config.json : %v\n", err)
			return
		}

		fmt.Printf("Contexte '%s' ajouté et sauvegardé dans la configuration.\n", contextName)
	},
}
