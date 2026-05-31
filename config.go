package main

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// SnooterConfig represents the root of the snooter.yaml file.
type SnooterConfig struct {
	Deployments map[string]DeploymentWrapper `yaml:"deployments"`
}

// DeploymentWrapper is a helper struct that handles the dynamic two-pass unmarshaling.
type DeploymentWrapper struct {
	Deployment Deployment
}

// UnmarshalYAML intercepts the parsing for the wrapper to safely determine the interface type.
func (w *DeploymentWrapper) UnmarshalYAML(value *yaml.Node) error {
	// 1. Peek at the "type" field
	var peek struct {
		Type string `yaml:"type"`
	}
	if err := value.Decode(&peek); err != nil {
		return err
	}

	// 2. Switch on the type and decode into the strict concrete struct
	switch peek.Type {
		case "DockerComposeDeployment":
			var dep DockerComposeDeployment
			if err := value.Decode(&dep); err != nil {
				return err
			}
			w.Deployment = dep
		default:
			return fmt.Errorf("unknown deployment type: %s", peek.Type)
	}
	return nil
}
