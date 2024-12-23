package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// readLine lit une ligne depuis l'entrée standard
func readLine(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// strToInt convertit une string en int (simplement)
func strToInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}

// runCommand exécute une commande système et redirige stdout/stderr
func runCommand(name string, args ...string) error {
	fmt.Printf("=> Commande : %s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// dockerLogin exécute un `docker login <registry>` interactif (mode mot de passe)
func dockerLogin(registry string) error {
	user := readLine("Nom d'utilisateur : ")
	pass := readLine("Mot de passe/Token : ")
	cmd := exec.Command("docker", "login", registry, "--username", user, "--password-stdin")
	cmd.Stdin = strings.NewReader(pass)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CheckDockerInstalled vérifie que la commande docker (et docker compose) sont dispo
func CheckDockerInstalled() error {
	// Vérifier "docker version"
	_, err := exec.LookPath("docker")
	if err != nil {
		return errors.New("docker n'est pas installé ou introuvable dans le PATH")
	}

	// Vérifier "docker compose" (sous-commande V2)
	// Essayons d'appeler "docker compose version"
	out, err := exec.Command("docker", "compose", "version").CombinedOutput()
	if err != nil {
		return fmt.Errorf("La commande 'docker compose' a échoué : %v", err)
	}

	if !strings.Contains(string(out), "Docker Compose") {
		return errors.New("'docker compose' semble indisponible ou incompatible")
	}

	return nil
}
