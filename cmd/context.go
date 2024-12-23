package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xpdemon/ac-deploy/config"
)

// ListContextsCmd : lists contexts present in the config AND on the machine
var ListContextsCmd = &cobra.Command{
	Use:   "list-contexts",
	Short: "List Docker contexts (in config + on the machine)",
	Run: func(cmd *cobra.Command, args []string) {
		// Retrieve contexts registered in the config
		cfgContexts := config.Cfg.DockerContexts

		// Retrieve contexts actually present on the machine
		localContexts, err := getLocalDockerContexts()
		if err != nil {
			fmt.Printf("Unable to retrieve the list of local contexts: %v\n", err)
			return
		}

		// Display
		fmt.Println("=== Docker Contexts in the Application Config ===")
		if len(cfgContexts) == 0 {
			fmt.Println("No Docker contexts are registered in the config.")
		} else {
			for i, ctx := range cfgContexts {
				fmt.Printf("[%d] %s (host=%s)\n", i, ctx.Name, ctx.Host)
				if ctx.Description != "" {
					fmt.Printf("    Description: %s\n", ctx.Description)
				}
			}
		}

		fmt.Println("\n=== Docker Contexts Detected on the Machine ===")
		if len(localContexts) == 0 {
			fmt.Println("No local Docker contexts detected (docker context ls is empty).")
		} else {
			for i, ctx := range localContexts {
				fmt.Printf("[%d] %s (host=%s)\n", i, ctx.Name, ctx.Host)
				if ctx.Description != "" {
					fmt.Printf("    Description: %s\n", ctx.Description)
				}
			}
		}
	},
}

// AddContextCmd : allows creating a new Docker context or registering an existing local context
var AddContextCmd = &cobra.Command{
	Use:   "add-context",
	Short: "Create a new Docker context or register an existing local context in the configuration",
	Run: func(cmd *cobra.Command, args []string) {
		// Choose mode
		choice := readLine("Do you want to (C)reate a new context or (R)egister an existing context? (C/R): ")
		switch strings.ToLower(choice) {
		case "c":
			createNewContext()
		case "r":
			registerExistingContext()
		default:
			fmt.Println("Invalid choice, cancellation.")
			return
		}
	},
}

// createNewContext contains the logic to create a real Docker context (docker context create).
func createNewContext() {
	// 1. Ask for the context name
	contextName := readLine("Name of the new Docker context: ")
	if contextName == "" {
		fmt.Println("Context name is required, cancellation.")
		return
	}

	// 2. Ask for an optional description
	desc := readLine("Description (optional): ")

	// 3. Ask for the Docker Host (e.g., ssh://user@server or tcp://192.168.1.10:2376)
	host := readLine("Docker Host (e.g., ssh://user@host, tcp://X.X.X.X:2376): ")
	if host == "" {
		fmt.Println("Docker host is required, cancellation.")
		return
	}

	// 4. Build the docker context create command
	dockerArgs := []string{
		"context", "create",
		contextName,
		"--docker", fmt.Sprintf("host=%s", host),
	}
	// Add the description if provided
	if desc != "" {
		dockerArgs = append(dockerArgs, "--description", desc)
	}

	fmt.Println("==> Creating context via docker CLI...")
	err := runCommand("docker", dockerArgs...)
	if err != nil {
		fmt.Printf("Failed to create Docker context: %v\n", err)
		return
	}
	fmt.Printf("Docker context '%s' created successfully.\n", contextName)

	// === Connection Test for the Context ===
	fmt.Println("==> Testing connection for the new context...")
	if err := testDockerContext(contextName, host); err != nil {
		fmt.Printf("Connection test failed: %v\n", err)
		choice := readLine("Do you want to remove this context (y/n)? : ")
		if strings.ToLower(choice) == "y" {
			_ = removeDockerContext(contextName) // ignoring error
			fmt.Printf("Context '%s' has been removed.\n", contextName)
		} else {
			fmt.Println("Context is kept, but it may be non-functional.")
		}
		return // stop here without saving to config
	}

	// 5. Store this context in the config for future use (only if test is OK)
	newCtx := config.DockerContext{
		Name:        contextName,
		Description: desc,
		Host:        host,
	}

	config.Cfg.DockerContexts = append(config.Cfg.DockerContexts, newCtx)

	// 6. Save the config
	err = config.SaveConfig()
	if err != nil {
		fmt.Printf("Error while saving: %v\n", err)
		return
	}

	fmt.Printf("Context '%s' added and saved in the configuration.\n", contextName)
}

// registerExistingContext goes through the list of local Docker contexts and allows the user to register one.
func registerExistingContext() {
	localContexts, err := getLocalDockerContexts()
	if err != nil {
		fmt.Printf("Error: unable to list existing contexts: %v\n", err)
		return
	}
	if len(localContexts) == 0 {
		fmt.Println("No local contexts are available for registration.")
		return
	}

	fmt.Println("Available local contexts:")
	for i, ctx := range localContexts {
		fmt.Printf("[%d] %s (host=%s)\n", i, ctx.Name, ctx.Host)
	}

	idxInput := readLine("Select the index of the context to register in the config: ")
	idx := strToInt(idxInput)
	if idx < 0 || idx >= len(localContexts) {
		fmt.Println("Invalid index.")
		return
	}

	selectedCtx := localContexts[idx]

	// Check if it's already in the config
	for _, c := range config.Cfg.DockerContexts {
		if c.Name == selectedCtx.Name {
			fmt.Printf("Context '%s' is already present in the configuration.\n", selectedCtx.Name)
			return
		}
	}

	// Add to config
	config.Cfg.DockerContexts = append(config.Cfg.DockerContexts, selectedCtx)
	err = config.SaveConfig()
	if err != nil {
		fmt.Printf("Error while saving the config: %v\n", err)
		return
	}
	fmt.Printf("Context '%s' registered in the configuration.\n", selectedCtx.Name)
}

// testDockerContext executes a simple command to verify that the context is functional
// and in case of error, attempts to add the SSH key (if Host key verification failed or Permission denied).
func testDockerContext(contextName, dockerHost string) error {
	// Execute "docker --context <name> info"
	err := runCommand("docker", "--context", contextName, "info")
	if err == nil {
		return nil // OK
	}

	// Convert the error to string for analysis
	errStr := err.Error()

	// Check for "Host key verification failed" or "Permission denied"
	if strings.Contains(errStr, "Host key verification failed") || strings.Contains(errStr, "Permission denied") {
		fmt.Println("==> SSH connection failed.")

		// Parse the dockerHost URL, e.g., "ssh://user@192.168.1.16"
		user, hostAddr, parseErr := parseSSHURL(dockerHost)
		if parseErr != nil {
			return fmt.Errorf("Unable to parse SSH URL for key addition: %v", parseErr)
		}

		// Offer the user to add the SSH key automatically in 'accept-new' mode
		choice := readLine("Do you want to automatically add the SSH key (mode 'accept-new')? (y/n): ")
		if strings.ToLower(choice) == "y" {
			// Prompt for SSH password
			password := readPassword("Enter SSH password: ")

			if password == "" {
				return errors.New("SSH password not provided")
			}

			// Attempt to add the SSH key via sshpass
			if sshErr := addSSHKeyWithPassword(user, hostAddr, password); sshErr != nil {
				return fmt.Errorf("Unable to add SSH key: %v", sshErr)
			}

			// Retry the Docker command after adding the key
			fmt.Println("SSH key added, retrying Docker connection...")
			if reErr := runCommand("docker", "--context", contextName, "info"); reErr != nil {
				return fmt.Errorf("Docker connection still fails after adding the key: %v\n, "+
					"please consult https://gist.github.com/dnaprawa/d3cfd6e444891c84846e099157fd51ef to add your public key \n"+
					"on the remote machine", reErr)
			}
			// If it passes, it's good
			return nil
		} else {
			return errors.New("SSH key missing, aborting")
		}
	}

	// Otherwise, it's another error
	return err
}

// removeDockerContext removes a Docker context
func removeDockerContext(contextName string) error {
	// -f to force removal (just in case)
	return runCommand("docker", "context", "rm", "-f", contextName)
}

// addSSHKeyWithPassword executes an SSH command with sshpass to add the SSH key
func addSSHKeyWithPassword(user, hostAddr, password string) error {
	fmt.Printf("==> Adding SSH key via 'sshpass' for %s@%s\n", user, hostAddr)

	// Check if sshpass is installed
	if !isSSHpassInstalled() {
		return errors.New("sshpass is not installed. Please install it to continue")
	}

	// Build the sshpass command
	// Force only password authentication
	// Limit authentication methods to avoid "Too many authentication failures"
	// Use PreferredAuthentications=password and PubkeyAuthentication=no
	// sshpass -p "password" ssh -o StrictHostKeyChecking=accept-new -o PreferredAuthentications=password -o PubkeyAuthentication=no user@host "exit"

	cmdArgs := []string{
		"-p", password,
		"ssh",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "PreferredAuthentications=password",
		"-o", "PubkeyAuthentication=no",
		fmt.Sprintf("%s@%s", user, hostAddr),
		"exit",
	}

	cmd := exec.Command("sshpass", cmdArgs...)

	// Redirect stdout/stderr for feedback
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("sshpass command failed: %v", err)
	}

	return nil
}

// isSSHpassInstalled checks if sshpass is available in the PATH
func isSSHpassInstalled() bool {
	_, err := exec.LookPath("sshpass")
	return err == nil
}

// parseSSHURL extracts the user and host from a URL like "ssh://user@host" or returns an error if not manageable
func parseSSHURL(sshURL string) (string, string, error) {
	// Example sshURL: "ssh://root@192.168.1.16"
	// 1. Remove the "ssh://" prefix
	if !strings.HasPrefix(sshURL, "ssh://") {
		return "", "", fmt.Errorf("URL does not start with ssh:// : %s", sshURL)
	}
	cleaned := strings.TrimPrefix(sshURL, "ssh://")

	// 2. Expecting "user@host"
	//    Using a simple regexp => ^([^@]+)@(.+)$
	re := regexp.MustCompile(`^([^@]+)@(.+)$`)
	matches := re.FindStringSubmatch(cleaned)
	if len(matches) < 3 {
		return "", "", fmt.Errorf("'user@host' format invalid in '%s'", cleaned)
	}
	user := matches[1]
	hostAddr := matches[2]

	return user, hostAddr, nil
}

// readPassword reads a password from standard input without displaying it
func readPassword(prompt string) string {
	fmt.Print(prompt)
	// Disable character echoing
	// Works on Unix systems (Linux, macOS)
	cmd := exec.Command("stty", "-echo")
	cmd.Stdin = os.Stdin
	cmd.Run()

	// Read the password
	reader := bufio.NewReader(os.Stdin)
	password, _ := reader.ReadString('\n')

	// Re-enable character echoing
	cmd = exec.Command("stty", "echo")
	cmd.Stdin = os.Stdin
	cmd.Run()

	fmt.Println() // New line after password input
	return strings.TrimSpace(password)
}
