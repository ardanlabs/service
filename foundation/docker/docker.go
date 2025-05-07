// Package docker provides support for starting and stopping docker containers
// for running tests.
package docker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"time"
)

// Container tracks information about the docker container started for tests.
type Container struct {
	Name     string
	HostPort string
}

// StartContainer starts the specified container for running tests.
func StartContainer(image string, name string, port string, dockerArgs []string, appArgs []string) (Container, error) {

	// When this code is used in tests, each test could be running in it's own
	// process, so there is no way to serialize the call. The idea is to wait
	// for the container to exist if the code fails to start it.
	for i := range 2 {
		c, err := startContainer(image, name, port, dockerArgs, appArgs)
		if err != nil {
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
			continue
		}

		return c, nil
	}

	return startContainer(image, name, port, dockerArgs, appArgs)
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

// =============================================================================

func startContainer(image string, name string, port string, dockerArgs []string, appArgs []string) (Container, error) {
	if c, err := exists(name, port); err == nil {
		return c, nil
	}

	// Just in case there is a container with the same name.
	exec.Command("docker", "rm", name, "-v").Run()

	arg := []string{"run", "-P", "-d", "--name", name}
	arg = append(arg, dockerArgs...)
	arg = append(arg, image)
	arg = append(arg, appArgs...)

	var out bytes.Buffer
	cmd := exec.Command("docker", arg...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return Container{}, fmt.Errorf("could not start container %s: %w: %s", image, err, arg)
	}

	id := out.String()[:12]
	hostIP, hostPort, err := extractIPPort(id, port)
	if err != nil {
		StopContainer(id)
		return Container{}, fmt.Errorf("could not extract ip/port: %w", err)
	}

	c := Container{
		Name:     name,
		HostPort: net.JoinHostPort(hostIP, hostPort),
	}

	return c, nil
}

func exists(name string, port string) (Container, error) {
	hostIP, hostPort, err := extractIPPort(name, port)
	if err != nil {
		return Container{}, errors.New("container not running")
	}

	c := Container{
		Name:     name,
		HostPort: net.JoinHostPort(hostIP, hostPort),
	}

	return c, nil
}

func extractIPPort(name string, port string) (hostIP string, hostPort string, err error) {

	// When IPv6 is turned on with Docker.
	// Got  [{"HostIp":"0.0.0.0","HostPort":"49190"}{"HostIp":"::","HostPort":"49190"}]
	// Need [{"HostIp":"0.0.0.0","HostPort":"49190"},{"HostIp":"::","HostPort":"49190"}]
	// '[{{range $i,$v := (index .NetworkSettings.Ports "5432/tcp")}}{{if $i}},{{end}}{{json $v}}{{end}}]'

	tmpl := fmt.Sprintf("[{{range $i,$v := (index .NetworkSettings.Ports \"%s/tcp\")}}{{if $i}},{{end}}{{json $v}}{{end}}]", port)

	var out bytes.Buffer
	cmd := exec.Command("docker", "inspect", "-f", tmpl, name)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("could not inspect container %s: %w", name, err)
	}

	var docs []struct {
		HostIP   string `json:"HostIp"`
		HostPort string `json:"HostPort"`
	}
	if err := json.Unmarshal(out.Bytes(), &docs); err != nil {
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
