// Package docker provides support for starting and stopping docker containers
// for running tests.
package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

// Container tracks information about the docker container started for tests.
type Container struct {
	ID   string
	Host string // IP:Port
}

// StartContainer starts the specified container for running tests.
func StartContainer(image string, port string, args ...string) (*Container, error) {
	arg := []string{"run", "-P", "-d"}
	arg = append(arg, args...)
	arg = append(arg, image)

	cmd := exec.Command("docker", arg...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("could not start container %s: %w", image, err)
	}

	id := out.String()[:12]
	hostIP, hostPort, err := extractIPPort(id, port)
	if err != nil {
		StopContainer(id)
		return nil, fmt.Errorf("could not extract ip/port: %w", err)
	}

	c := Container{
		ID:   id,
		Host: net.JoinHostPort(hostIP, hostPort),
	}

	return &c, nil
}

// StopContainer stops and removes the specified container.
func StopContainer(id string) error {
	if err := exec.Command("docker", "stop", id).Run(); err != nil {
		return fmt.Errorf("could not stop container: %w", err)
	}

	if err := exec.Command("docker", "rm", id, "-v").Run(); err != nil {
		return fmt.Errorf("could not remove container: %w", err)
	}

	return nil
}

// DumpContainerLogs outputs logs from the running docker container.
func DumpContainerLogs(id string) []byte {
	out, err := exec.Command("docker", "logs", id).CombinedOutput()
	if err != nil {
		return nil
	}
	return out
}

func extractIPPort(id string, port string) (hostIP string, hostPort string, err error) {
	tmpl := fmt.Sprintf("[{{range $k,$v := (index .NetworkSettings.Ports \"%s/tcp\")}}{{json $v}}{{end}}]", port)

	cmd := exec.Command("docker", "inspect", "-f", tmpl, id)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("could not inspect container %s: %w", id, err)
	}

	// When IPv6 is turned on with Docker.
	// Got  [{"HostIp":"0.0.0.0","HostPort":"49190"}{"HostIp":"::","HostPort":"49190"}]
	// Need [{"HostIp":"0.0.0.0","HostPort":"49190"},{"HostIp":"::","HostPort":"49190"}]
	data := strings.ReplaceAll(out.String(), "}{", "},{")

	var docs []struct {
		HostIP   string
		HostPort string
	}
	if err := json.Unmarshal([]byte(data), &docs); err != nil {
		return "", "", fmt.Errorf("could not decode json: %w", err)
	}

	for _, doc := range docs {
		if doc.HostIP != "::" {
			return doc.HostIP, doc.HostPort, nil
		}
	}

	return "", "", fmt.Errorf("could not locate ip/port")
}
