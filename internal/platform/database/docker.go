package database

import (
	"bytes"
	"encoding/json"
	"os/exec"

	"github.com/pkg/errors"
)

// Container tracks information about a docker container started for tests.
type Container struct {
	ID   string
	Host string // IP:Port
}

// StartContainer runs a postgres container to execute commands.
func StartContainer() (*Container, error) {
	cmd := exec.Command("docker", "run", "-P", "-d", "postgres:11.1-alpine")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrap(err, "starting docker")
	}

	id := out.String()[:12]

	cmd = exec.Command("docker", "inspect", id)
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "could not inspect container %s", id)
	}

	var doc []struct {
		NetworkSettings struct {
			Ports struct {
				TCP5432 []struct {
					HostIP   string `json:"HostIp"`
					HostPort string `json:"HostPort"`
				} `json:"5432/tcp"`
			} `json:"Ports"`
		} `json:"NetworkSettings"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		return nil, errors.Wrap(err, "could not decode json")
	}

	network := doc[0].NetworkSettings.Ports.TCP5432[0]

	c := Container{
		ID:   id,
		Host: network.HostIP + ":" + network.HostPort,
	}

	return &c, nil
}

// StopContainer stops and removes the specified container.
func StopContainer(c *Container) error {
	if err := exec.Command("docker", "stop", c.ID).Run(); err != nil {
		return errors.Wrapf(err, "could not stop container: %s", c.ID)
	}

	if err := exec.Command("docker", "rm", c.ID, "-v").Run(); err != nil {
		return errors.Wrapf(err, "could not remove container: %s", c.ID)
	}

	return nil
}

// DumpContainerLogs runs "docker logs" against the container and send it to t.Log
func DumpContainerLogs(c *Container) (string, error) {
	out, err := exec.Command("docker", "logs", c.ID).CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "could not log container: %s", c.ID)
	}
	return string(out), nil
}
