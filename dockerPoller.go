package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// DockerPoller encapsulates the Docker daemon client.
type DockerPoller struct {
	cli *client.Client
}

// NewDockerPoller initializes a new Docker client connecting via the local socket.
func NewDockerPoller() (*DockerPoller, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize docker client")
	}
	return &DockerPoller{cli: cli}, nil
}

// getComposeProjectName determines the project name by checking (in order):
// 1. A top-level `name:` key in the docker-compose.yml file.
// 2. A `COMPOSE_PROJECT_NAME` variable in a `.env` file in the compose directory.
// 3. The base name of the compose file's parent directory.
func getComposeProjectName(composePath string) (string, error) {
	// 1. Check for top-level 'name:' in docker-compose.yml
	composeFilePath := filepath.Join(composePath, "docker-compose.yml")
	if _, err := os.Stat(composeFilePath); err == nil {
		data, err := os.ReadFile(composeFilePath)
		if err != nil {
			return "", errors.Wrap(err, "failed to read docker-compose.yml")
		}
		var compose struct {
			Name string `yaml:"name"`
		}
		// This unmarshal might fail if the compose file is old, that's okay.
		if err := yaml.Unmarshal(data, &compose); err == nil && compose.Name != "" {
			return compose.Name, nil
		}
	}

	// 2. Check for COMPOSE_PROJECT_NAME in .env file
	envFilePath := filepath.Join(composePath, ".env")
	if _, err := os.Stat(envFilePath); err == nil {
		envMap, err := godotenv.Read(envFilePath)
		if err == nil {
			if projectName, ok := envMap["COMPOSE_PROJECT_NAME"]; ok && projectName != "" {
				return projectName, nil
			}
		}
	}

	// 3. Fallback to directory name
	return filepath.Base(composePath), nil
}

// GetActiveImageVersions checks the Docker daemon for all running containers of a Compose project.
// It returns a map of service names to their image tags (versions).
func (dp *DockerPoller) GetActiveImageVersions(ctx context.Context, dep DockerComposeDeployment) (map[string]string, error) {
	// 1. Intelligently determine the project name.
	projectName, err := getComposeProjectName(dep.ComposePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to determine compose project name")
	}

	// 2. Find all running containers for this project.
	f := filters.NewArgs()
	f.Add("label", "com.docker.compose.project="+projectName)
	containers, err := dp.cli.ContainerList(ctx, container.ListOptions{Filters: f})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query docker for deployment: %s", dep.GetName())
	}

	versions := make(map[string]string)
	if len(containers) == 0 {
		return versions, nil // Likely stopped, building, or not deployed yet
	}

	// 3. Extract versions for all containers in the stack.
	for _, c := range containers {
		serviceName := c.Labels["com.docker.compose.service"]
		if serviceName == "" {
			// Fallback if the compose service label is somehow missing
			serviceName = strings.TrimPrefix(c.Names[0], "/")
		}
		versions[serviceName] = extractTag(c.Image)
	}

	return versions, nil
}

// extractTag safely extracts the version tag from a full Docker image string.
func extractTag(imageName string) string {
	imageName = strings.Split(imageName, "@")[0]
	i := strings.LastIndex(imageName, ":")
	if i == -1 || strings.Contains(imageName[i+1:], "/") {
		return "latest"
	}
	return imageName[i+1:]
}

// Close cleanly shuts down the Docker client connection.
func (dp *DockerPoller) Close() error {
	return dp.cli.Close()
}
