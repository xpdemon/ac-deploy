package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/xpdemon/ac-deploy/config"
	"os"
	"os/exec"
	"strings"
)

// getLocalDockerContexts returns the list of Docker contexts
// actually present on the machine (via docker CLI).
func getLocalDockerContexts() ([]config.DockerContext, error) {
	// Format for listing: choose a "parsable" format.
	// For example: name|description|dockerEndpoint
	formatString := "{{.Name}}|{{.Description}}|{{.DockerEndpoint}}"

	cmd := exec.Command("docker", "context", "ls", "--format", formatString)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Error executing docker context ls: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []config.DockerContext
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			// Ignore any malformed lines
			continue
		}
		ctx := config.DockerContext{
			Name:        parts[0],
			Description: parts[1],
			Host:        parts[2],
		}
		result = append(result, ctx)
	}

	return result, nil
}

// readLine reads a line from standard input
func readLine(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// strToInt converts a string to an int (simply)
func strToInt(s string) int {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return 0
	}
	return i
}

// runCommand executes a system command and redirects stdout/stderr
func runCommand(name string, args ...string) error {
	fmt.Printf("=> Command: %s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// dockerLogin executes an interactive `docker login <registry>` (password mode)
func dockerLogin(registry string) error {
	user := readLine("Username: ")
	pass := readLine("Password/Token: ")
	cmd := exec.Command("docker", "login", registry, "--username", user, "--password-stdin")
	cmd.Stdin = strings.NewReader(pass)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CheckDockerInstalled verifies that the docker command (and docker compose) are available
func CheckDockerInstalled() error {
	// Check "docker version"
	_, err := exec.LookPath("docker")
	if err != nil {
		return errors.New("docker is not installed or not found in PATH")
	}

	// Check "docker compose" (V2 subcommand)
	// Let's try calling "docker compose version"
	out, err := exec.Command("docker", "compose", "version").CombinedOutput()
	if err != nil {
		return fmt.Errorf("The 'docker compose' command failed: %v", err)
	}

	if !strings.Contains(string(out), "Docker Compose") {
		return errors.New("'docker compose' seems unavailable or incompatible")
	}

	return nil
}
