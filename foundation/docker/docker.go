// Package docker provides support for starting and stopping docker containers
// for running tests.
package docker

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/go-json-experiment/json"
)

// Container tracks information about the docker container started for tests.
type Container struct {
	Name     string
	HostPort string
}

// StartContainer starts the specified container for running tests.
func StartContainer(image string, name string, port string, dockerArgs []string, appArgs []string) (*Container, error) {
	if name == "" || port == "" {
		return nil, errors.New("image name and port is required")
	}

	if c, err := exists(name, port); err == nil {
		return c, nil
	}

	arg := []string{"run", "-P", "-d", "--name", name}
	arg = append(arg, dockerArgs...)
	arg = append(arg, image)
	arg = append(arg, appArgs...)

	var out bytes.Buffer
	cmd := exec.Command("docker", arg...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {

		// CircleCI might fail here, try again without naming the container.
		arg = []string{"run", "-P", "-d"}
		arg = append(arg, dockerArgs...)
		arg = append(arg, image)
		arg = append(arg, appArgs...)

		out.Reset()
		cmd := exec.Command("docker", arg...)
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("could not start container %s: %w", image, err)
		}
	}

	id := out.String()[:12]
	hostIP, hostPort, err := extractIPPort(id, port)
	if err != nil {
		StopContainer(id)
		return nil, fmt.Errorf("could not extract ip/port: %w", err)
	}

	c := Container{
		Name:     name,
		HostPort: net.JoinHostPort(hostIP, hostPort),
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

func exists(name string, port string) (*Container, error) {
	hostIP, hostPort, err := extractIPPort(name, port)
	if err != nil {
		return nil, errors.New("container not running")
	}

	c := Container{
		Name:     name,
		HostPort: net.JoinHostPort(hostIP, hostPort),
	}

	return &c, nil
}

func extractIPPort(name string, port string) (hostIP string, hostPort string, err error) {
	tmpl := fmt.Sprintf("[{{range $k,$v := (index .NetworkSettings.Ports \"%s/tcp\")}}{{json $v}}{{end}}]", port)

	var out bytes.Buffer
	cmd := exec.Command("docker", "inspect", "-f", tmpl, name)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("could not inspect container %s: %w", name, err)
	}

	// When IPv6 is turned on with Docker.
	// Got  [{"HostIp":"0.0.0.0","HostPort":"49190"}{"HostIp":"::","HostPort":"49190"}]
	// Need [{"HostIp":"0.0.0.0","HostPort":"49190"},{"HostIp":"::","HostPort":"49190"}]
	data := strings.ReplaceAll(out.String(), "}{", "},{")

	var docs []struct {
		HostIP   string `json:"HostIp"`
		HostPort string `json:"HostPort"`
	}
	if err := json.Unmarshal([]byte(data), &docs); err != nil {
		return "", "", fmt.Errorf("could not decode json: %w", err)
	}

	for _, doc := range docs {
		if doc.HostIP != "::" {
			// Podman keeps HostIP empty instead of using 0.0.0.0.
			// - https://github.com/containers/podman/issues/17780
			if doc.HostIP == "" {
				return "localhost", doc.HostPort, nil
			}

			return doc.HostIP, doc.HostPort, nil
		}
	}

	return "", "", fmt.Errorf("could not locate ip/port")
}
