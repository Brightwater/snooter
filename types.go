package main

// Deployment interface for different deployment configs
type Deployment interface {
	GetName() string
	GetType() string
	Run() error
}

// rawDeployment used to peek at the type field before full unmarshaling
type RawDeployment struct {
	Type string `yaml:"type"`
}

// DockerComposeDeployment represents a deployment using a local Docker Compose project.
type DockerComposeDeployment struct {
	Name           string     `yaml:"name"`
	Type           string     `yaml:"type"`
	ValidationURL  string     `yaml:"validationUrl,omitempty"`
	ComposePath    string     `yaml:"composePath"`
	DockerFilePath string     `yaml:"dockerFilePath,omitempty"`
	ComposeType    string     `yaml:"composeType"`         // external or internal
	GitConfig      *GitConfig `yaml:"gitConfig,omitempty"` // required for internal
}

func (d DockerComposeDeployment) GetName() string {
	return d.Name
}
func (d DockerComposeDeployment) GetType() string {
	return d.Type
}
func (d DockerComposeDeployment) Run() error {
	return RunCompose(d)
}

// GitConfig holds optional git-related settings.
type GitConfig struct {
	Pull bool   `yaml:"pull"`
	Path string `yaml:"path"`
}
