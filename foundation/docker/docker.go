// Package docker provides support for starting and stopping docker containers
// for running tests.
package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"testing"
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
	hostIp, hostPort, err := extractIPPort(id, port)
	if err != nil {
		return nil, fmt.Errorf("could not extract ip/port: %w", err)
	}

	c := Container{
		ID:   id,
		Host: net.JoinHostPort(hostIp, hostPort),
	}

	fmt.Printf("Image:       %s\n", image)
	fmt.Printf("ContainerID: %s\n", c.ID)
	fmt.Printf("Host:        %s\n", c.Host)

	return &c, nil
}

// StopContainer stops and removes the specified container.
func StopContainer(id string) error {
	if err := exec.Command("docker", "stop", id).Run(); err != nil {
		return fmt.Errorf("could not stop container: %w", err)
	}
	fmt.Println("Stopped:", id)

	if err := exec.Command("docker", "rm", id, "-v").Run(); err != nil {
		return fmt.Errorf("could not remove container: %w", err)
	}
	fmt.Println("Removed:", id)

	return nil
}

// DumpContainerLogs outputs logs from the running docker container.
func DumpContainerLogs(t *testing.T, id string) {
	out, err := exec.Command("docker", "logs", id).CombinedOutput()
	if err != nil {
		t.Fatalf("could not log container: %v", err)
	}
	t.Logf("Logs for %s\n%s:", id, out)
}

func extractIPPort(id string, port string) (hostIP string, hostPort string, err error) {
	cmd := exec.Command("docker", "inspect", "-f", "{{json .NetworkSettings.Ports}}", id)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("could not inspect container %s: %w", id, err)
	}

	s := out.Bytes()
	fmt.Println(string(s))

	// {"5432/tcp":[{"HostIp":"0.0.0.0","HostPort":"55000"}],"9080/tcp":null}
	var doc map[string]interface{}
	if err := json.Unmarshal(s, &doc); err != nil {
		return "", "", fmt.Errorf("could not decode json: %w", err)
	}

	// [{"HostIp":"0.0.0.0","HostPort":"55000"}]
	tcp, exists := doc[port+"/tcp"]
	if !exists {
		return "", "", fmt.Errorf("could not find %q", port+"/tcp")
	}

	// [{"HostIp":"0.0.0.0","HostPort":"55000"}]
	list, exists := tcp.([]interface{})
	if !exists {
		return "", "", fmt.Errorf("could not find host list information")
	}

	// {"HostIp":"0.0.0.0","HostPort":"55000"}
	data, exists := list[0].(map[string]interface{})
	if !exists {
		return "", "", fmt.Errorf("could not find host information")
	}

	hostIP = data["HostIp"].(string)
	hostPort = data["HostPort"].(string)

	return hostIP, hostPort, nil
}
