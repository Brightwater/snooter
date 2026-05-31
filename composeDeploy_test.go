package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// ==============================================================================================
// SNOOTER INTEGRATION TESTS
//
// Run standard sandbox tests (safe, isolated):           go test -v ./...
// Run LIVE production stack tests (modifies real apps):  RUN_REAL_STACKS=true go test -v ./...
// ==============================================================================================

// IntegrationTestConfig allows toggling behaviors for integration tests.
type IntegrationTestConfig struct {
	CleanupContainers bool
}

var testConfig = IntegrationTestConfig{
	CleanupContainers: true, // Set to false to leave containers running for manual inspection
}

func TestRunExternalCompose(t *testing.T) {
	// Read the external test config from the testdata folder
	composeContent, err := os.ReadFile("testdata/test_compose.yaml")
	if err != nil {
		t.Fatalf("Failed to read external test config: %v", err)
	}

	// 1. Create a safe temporary directory that Go will automatically clean up
	tmpDir := t.TempDir()

	// Register cleanup to tear down the docker compose stack
	t.Cleanup(func() {
		if testConfig.CleanupContainers {
			downCmd := exec.Command("docker", "compose", "down")
			downCmd.Dir = tmpDir
			out, err := RunCommandCaptureOutput(downCmd)
			if err != nil {
				t.Logf("Failed to clean up external compose stack: %v\nOutput: %s", err, out)
			}
		}
	})

	// 2. Write the external test config into the temp directory as docker-compose.yml
	err = os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), composeContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write test compose file: %v", err)
	}

	// 3. Manually construct a Snooter deployment object
	dep := DockerComposeDeployment{
		Name:        "test-hello-world",
		Type:        "DockerComposeDeployment",
		ComposePath: tmpDir,
		ComposeType: "external",
	}

	// 4. Execute the actual deployment logic
	if err := dep.Run(); err != nil {
		t.Errorf("Expected deployment to succeed, but got error: %v", err)
	}
}

func TestRunInternalCompose_GoDrive(t *testing.T) {
	// Read the mock compose file for the local godrive stack
	composeContent, err := os.ReadFile("testdata/godrive_compose.yaml")
	if err != nil {
		t.Fatalf("Failed to read godrive test config: %v", err)
	}

	tmpDir := t.TempDir()

	// Register cleanup to tear down the docker compose stack
	t.Cleanup(func() {
		if testConfig.CleanupContainers {
			downCmd := exec.Command("docker", "compose", "down")
			downCmd.Dir = tmpDir
			out, err := RunCommandCaptureOutput(downCmd)
			if err != nil {
				t.Logf("Failed to clean up internal compose stack: %v\nOutput: %s", err, out)
			}
		}
	})

	err = os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), composeContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write test compose file: %v", err)
	}

	// Construct the internal deployment object for GoDrive
	dep := DockerComposeDeployment{
		Name:        "godrive",
		Type:        "DockerComposeDeployment",
		ComposePath: tmpDir,
		ComposeType: "internal",
		GitConfig: &GitConfig{
			Pull: true,
			Path: "https://github.com/Brightwater/goDrive", // The real repo to clone and build
		},
	}

	if err := dep.Run(); err != nil {
		t.Errorf("Expected internal godrive deployment to succeed, but got error: %v", err)
	}
}

func TestRealConfigDeployments(t *testing.T) {
	// Safety toggle: Only run this if the environment variable is explicitly set
	if os.Getenv("RUN_REAL_STACKS") != "true" {
		t.Skip("Skipping real stack tests. Run with RUN_REAL_STACKS=true to execute.")
	}

	data, err := os.ReadFile("snooter.yaml")
	if err != nil {
		t.Fatalf("Failed to read snooter.yaml: %v", err)
	}

	var config SnooterConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse snooter.yaml: %v", err)
	}

	for _, dw := range config.Deployments {
		dep := dw.Deployment
		t.Logf("=== Executing Real Deployment: %s (%s) ===", dep.GetName(), dep.GetType())

		// Execute the actual deployment in its real directory
		if err := dep.Run(); err != nil {
			t.Errorf("Failed to run real deployment %s: %v", dep.GetName(), err)
		}
	}
}
