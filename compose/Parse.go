package compose

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ComposeFile struct {
	Services map[string]struct {
		Image string `yaml:"image"`
	} `yaml:"services"`
}

// ParseComposeFile reads a docker-compose.yml and returns the list of images
func ParseComposeFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c ComposeFile
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	var images []string
	for _, svc := range c.Services {
		if svc.Image != "" {
			images = append(images, svc.Image)
		}
	}
	return images, nil
}
