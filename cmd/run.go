package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xpdemon/ac-deploy/compose"
	"github.com/xpdemon/ac-deploy/config"
	"gopkg.in/yaml.v3"
)

var RunFlowCmd = &cobra.Command{
	Use:   "run-flow",
	Short: "Execute the complete flow: choose contexts, build, push, deploy",
	Run: func(cmd *cobra.Command, args []string) {
		// 1) Check if there are contexts
		if len(config.Cfg.DockerContexts) == 0 {
			fmt.Println("No Docker contexts are registered. Use `xpdemon-deploy add-context`.")
			return
		}

		// 2) Display the list of available contexts
		fmt.Println("List of Docker contexts:")
		for i, c := range config.Cfg.DockerContexts {
			fmt.Printf("  [%d] %s (host=%s)\n", i, c.Name, c.Host)
		}

		// 3) Select the context for BUILDER
		buildIdxInput := readLine("Choose the index of the context for BUILDER: ")
		buildIdx := strToInt(buildIdxInput)
		if buildIdx < 0 || buildIdx >= len(config.Cfg.DockerContexts) {
			fmt.Println("Invalid index for build context.")
			return
		}
		buildContext := config.Cfg.DockerContexts[buildIdx]

		// 4) Select the context for DEPLOY
		deployIdxInput := readLine("Choose the index of the context for DEPLOY (no-build): ")
		deployIdx := strToInt(deployIdxInput)
		if deployIdx < 0 || deployIdx >= len(config.Cfg.DockerContexts) {
			fmt.Println("Invalid index for deploy context.")
			return
		}
		deployContext := config.Cfg.DockerContexts[deployIdx]

		// 5) Select the registry (optional)
		var registry string
		if len(config.Cfg.DockerRegistries) > 0 {
			fmt.Println("Available registries:")
			for i, r := range config.Cfg.DockerRegistries {
				fmt.Printf("  [%d] %s\n", i, r)
			}
			regIdxInput := readLine("Choose the index of the registry to push to (or press ENTER to skip): ")
			if regIdxInput != "" {
				regIdx := strToInt(regIdxInput)
				if regIdx >= 0 && regIdx < len(config.Cfg.DockerRegistries) {
					registry = config.Cfg.DockerRegistries[regIdx]
				}
			}
		}

		// 6) Path to docker-compose.yml
		composeFile := readLine("Path to your docker-compose.yml: ")
		if composeFile == "" {
			fmt.Println("Path to docker-compose.yml not specified, cancellation.")
			return
		}

		// 7) Parse docker-compose.yml to detect images
		images, err := compose.ParseComposeFile(composeFile)
		if err != nil {
			fmt.Printf("Error parsing docker-compose: %v\n", err)
			return
		}
		fmt.Println("Images detected in this docker-compose:")
		for _, img := range images {
			fmt.Printf("  - %s\n", img)
		}

		// === 7.a) Ask the user for a tag ===
		tagChoice := readLine("Enter the tag to apply (replace only if tag=latest). Leave empty to not change: ")
		if tagChoice != "" {
			if err := validateTag(tagChoice); err != nil {
				fmt.Printf("Invalid tag: %v\n", err)
				return
			}
		}

		// === 7.b) Offer a prefix (registry + user) if the user wants
		prefixChoice := ""
		if registry != "" {
			// Offer to add a prefix like "registry.com/myuser"
			prefixChoice = readLine("Do you want to prefix the images with the selected registry (e.g., my-registry.com/user)? (Press ENTER to skip): ")
		} else {
			prefixChoice = readLine("Do you want to prefix the images (e.g., my-registry.com/user)? (Press ENTER to skip): ")
		}

		var newComposePath string
		if tagChoice != "" || prefixChoice != "" {
			// 7.c) Generate a new compose if tag or prefix is requested
			newComposePath, err = generateTaggedCompose(composeFile, tagChoice, prefixChoice)
			if err != nil {
				fmt.Printf("Error generating the modified docker-compose file: %v\n", err)
				return
			}
			fmt.Printf("New docker-compose created: %s\n", newComposePath)
		} else {
			// No tag or prefix => use the original composeFile
			newComposePath = composeFile
		}

		// 7.d) Optional step: Prune before build
		pruneChoice := readLine("Do you want to prune the build context? (y/n): ")
		if strings.ToLower(pruneChoice) == "y" {
			// Can split into two questions for precise control:
			pruneImagesChoice := readLine("   > Remove unused Docker images (docker image prune -a)? (y/n): ")
			if strings.ToLower(pruneImagesChoice) == "y" {
				fmt.Println("==> Executing docker image prune...")
				err = runCommand(
					"docker",
					"--context", buildContext.Name,
					"image", "prune",
					"-a",
					"-f", // to force without asking for confirmation
				)
				if err != nil {
					fmt.Printf("Error pruning images: %v\n", err)
				}
			}

			pruneBuilderChoice := readLine("   > Remove Docker builder cache (docker builder prune)? (y/n): ")
			if strings.ToLower(pruneBuilderChoice) == "y" {
				fmt.Println("==> Executing docker builder prune...")
				err = runCommand(
					"docker",
					"--context", buildContext.Name,
					"builder", "prune",
					"-f",
				)
				if err != nil {
					fmt.Printf("Error pruning builder: %v\n", err)
				}
			}
		}

		// 8) Build
		fmt.Println("==> Building images...")
		err = runCommand(
			"docker",
			"--context", buildContext.Name,
			"compose",
			"-f", newComposePath,
			"build",
		)
		if err != nil {
			fmt.Printf("Error during build: %v\n", err)
			cleanupFile(newComposePath) // Optional cleanup
			return
		}

		// 9) Push
		pushChoice := readLine("Do you want to push the images to the selected registry? (y/n): ")
		if strings.ToLower(pushChoice) == "y" && registry != "" {
			fmt.Println("==> Pushing images...")
			err = runCommand(
				"docker",
				"--context", buildContext.Name,
				"compose",
				"-f", newComposePath,
				"push",
			)
			if err != nil {
				fmt.Printf("Error during push: %v\n", err)
				cleanupFile(newComposePath) // Optional cleanup
				return
			}
		}

		// 10) Deploy
		deployChoice := readLine("Do you want to deploy the images in no-build mode? (y/n): ")
		if strings.ToLower(deployChoice) == "y" {
			fmt.Println("==> Deploying in no-build mode...")
			err = runCommand(
				"docker",
				"--context", deployContext.Name,
				"compose",
				"-f", newComposePath,
				"up",
				"-d",
				"--no-build",
			)
			if err != nil {
				fmt.Printf("Error during deployment: %v\n", err)
			} else {
				fmt.Println("Deployment completed successfully!")
			}
		} else {
			fmt.Println("Deployment canceled.")
		}

		// Optional cleanup of the file
		cleanupChoice := readLine("Do you want to delete the temporary docker-compose file? (y/n): ")
		if strings.ToLower(cleanupChoice) == "y" {
			cleanupFile(newComposePath)
		}
	},
}

// generateTaggedCompose reads the original docker-compose as a map[string]interface{}.
// 1. Read the entire YAML content.
// 2. ONLY modify the 'image' key for each service (based on prefix / tagChoice).
// 3. Rewrite a new complete docker-compose, keeping all fields intact.
func generateTaggedCompose(originalPath, tagChoice, prefix string) (string, error) {
	// 1) Read the YAML file as binary
	data, err := os.ReadFile(originalPath)
	if err != nil {
		return "", err
	}

	// 2) Parse into a general map
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return "", err
	}

	// 3) Retrieve the "services" section (if it exists)
	servicesRaw, ok := raw["services"].(map[string]interface{})
	if !ok {
		// No services => nothing to tag
		// Can either return the original path or create a new unchanged file
		return "", fmt.Errorf("the docker-compose file does not contain a 'services' key")
	}

	// 4) Iterate through the services
	for svcName, svcVal := range servicesRaw {
		// Each service is potentially a map[string]interface{}
		svcMap, ok := svcVal.(map[string]interface{})
		if !ok {
			// Malformed service?
			continue
		}

		// Check if 'image' exists
		imageVal, ok := svcMap["image"].(string)
		if !ok {
			// No image => do not modify
			continue
		}

		// Split to extract the repo part and the tag part
		// Example: "xpdemon/ac-wotlk-authserver:test"
		// parts[0] = "xpdemon/ac-wotlk-authserver", parts[1] = "test"
		parts := strings.SplitN(imageVal, ":", 2)
		repoPart := parts[0]
		var oldTag string
		if len(parts) == 2 {
			oldTag = parts[1]
		}

		// 4.a) Add a prefix, if requested
		//    If prefix = "my-registry.com/myuser" and repoPart = "mysql", => "my-registry.com/myuser/mysql"
		//    Otherwise, if the image already contains "/", it's up to you how to concatenate.
		if prefix != "" {
			repoPart = fmt.Sprintf("%s/%s", prefix, repoPart)
		}

		// 4.b) Replace the tag ONLY if it was "latest" (and if tagChoice != "")
		//     - oldTag == "latest" => replace with tagChoice
		//     - oldTag != "latest" => leave as is
		//     - no tag => do nothing
		if oldTag == "latest" && tagChoice != "" {
			oldTag = tagChoice
		}

		// 4.c) Reconstruct the image
		var newImage string
		if oldTag == "" {
			// No tag => keep just "repoPart"
			newImage = repoPart
		} else {
			newImage = fmt.Sprintf("%s:%s", repoPart, oldTag)
		}

		// 4.d) Update in the map
		svcMap["image"] = newImage
		servicesRaw[svcName] = svcMap
	}

	// 5) Reinstate the modified services section into raw
	raw["services"] = servicesRaw

	// 6) Marshal back to YAML in a new file
	newYaml, err := yaml.Marshal(raw)
	if err != nil {
		return "", err
	}

	// 7) Construct a temporary or suffixed file path
	dir := filepath.Dir(originalPath)
	base := filepath.Base(originalPath)
	ext := filepath.Ext(base) // .yml or .yaml
	nameOnly := strings.TrimSuffix(base, ext)
	newFileName := fmt.Sprintf("%s-tagged%s", nameOnly, ext) // e.g., docker-compose-tagged.yml
	newPath := filepath.Join(dir, newFileName)

	// 8) Write the new content
	if err := os.WriteFile(newPath, newYaml, 0644); err != nil {
		return "", err
	}

	return newPath, nil
}

// validateTag checks the format of the tag (allowed characters)
func validateTag(tag string) error {
	// Ex. Docker tag must be lowercase alphanum + . _ -
	// This is a mini-check, you can adjust it
	matched, _ := regexp.MatchString(`^[a-z0-9._-]+$`, tag)
	if !matched {
		return errors.New("the tag must contain only [a-z0-9._-]")
	}
	return nil
}

// cleanupFile deletes a file if it exists
func cleanupFile(path string) {
	if path == "" {
		return
	}
	if path == "-" {
		return
	}
	if _, err := os.Stat(path); err == nil {
		err = os.Remove(path)
		if err == nil {
			fmt.Printf("Temporary file deleted: %s\n", path)
		} else {
			fmt.Printf("Unable to delete file %s: %v\n", path, err)
		}
	}
}
