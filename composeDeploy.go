package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/phuslu/log"
	"github.com/pkg/errors"
)

func RunCompose(d DockerComposeDeployment) error {
	if d.ComposeType == "external" {
		return runExternalCompose(d)
	}

	if d.ComposeType == "internal" {
		return runInternalCompose(d)
	}

	return fmt.Errorf("Invalid compose type")
}

func runExternalCompose(d DockerComposeDeployment) error {

	preMsg := fmt.Sprintf("**Deployment output log:**\n*%s %s external*\n", d.Name, d.Type)

	pullCmd := exec.Command("docker", "compose", "pull")
	pullCmd.Dir = d.ComposePath

	outA, err := RunCommandCaptureOutput(pullCmd)
	if err != nil {
		log.Error().Err(err).Msg(preMsg + "\n$> docker compose pull FAILURE!!!!\n" + outA)
		return err
	}

	buildCmd := exec.Command("docker", "compose", "up", "-d", "--force-recreate")
	buildCmd.Dir = d.ComposePath

	outB, err := RunCommandCaptureOutput(buildCmd)
	if err != nil {
		log.Error().Err(err).Msg(preMsg + "\n$> docker compose pull\n" + outA + "\n\n$> docker compose up -d --force-recreate FAILURE!!!!\n" + outB)
		return err
	}

	log.Info().Msg(preMsg + "\n$> docker compose pull\n" + outA + "\n\n$> docker compose up -d --force-recreate\n" + outB)

	return nil
}

func runInternalCompose(d DockerComposeDeployment) error {

	preMsg := fmt.Sprintf("**Deployment output log:**\n*%s %s internal*\n", d.Name, d.Type)
	tmpRepoPath := filepath.Join(os.TempDir(), "snooter-"+d.Name)

	pullCmd := exec.Command("git", "clone", d.GitConfig.Path, tmpRepoPath)
	outA, err := RunCommandCaptureOutput(pullCmd)
	if err != nil {
		log.Error().Err(err).Msg(preMsg + "\n$> git clone FAILED!!\n" + outA)
		return err
	}

	buildCmd := exec.Command("docker", "build", "-t", d.Name, tmpRepoPath)
	outB, err := RunCommandCaptureOutput(buildCmd)
	if err != nil {
		cleanUpGit(tmpRepoPath)
		log.Error().Err(err).Msg(preMsg + "\n$> docker build FAILED!!\n" + outB)
		return err
	}

	composeCmd := exec.Command("docker", "compose", "up", "-d", "--force-recreate")
	composeCmd.Dir = d.ComposePath
	outC, err := RunCommandCaptureOutput(composeCmd)
	if err != nil {
		cleanUpGit(tmpRepoPath)
		log.Error().Err(err).Msg(preMsg + "\n$> docker compose up -d FAILED!!\n" + outC)
		return errors.Wrap(err, "Failed to docker compose")
	}

	cleanUpGit(tmpRepoPath)

	log.Info().Msg(preMsg + "\n$> git clone " + tmpRepoPath + "\n" + outA + "\n\n$> docker build -t " + d.Name + "\n" + outB + "\n\n$> docker compose up -d --force-recreate\n" + outC)

	return nil
}

func cleanUpGit(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to delete dir: %s", path)
	}
}
