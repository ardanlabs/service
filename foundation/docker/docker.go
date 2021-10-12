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

	cmd = exec.Command("docker", "inspect", id)
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("could not inspect container %s: %w", id, err)
	}

	var doc []map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		return nil, fmt.Errorf("could not decode json: %w", err)
	}

	ip, randPort := extractIPPort(doc, port)

	c := Container{
		ID:   id,
		Host: net.JoinHostPort(ip, randPort),
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

func extractIPPort(doc []map[string]interface{}, port string) (string, string) {
	nw, exists := doc[0]["NetworkSettings"]
	if !exists {
		return "", ""
	}
	ports, exists := nw.(map[string]interface{})["Ports"]
	if !exists {
		return "", ""
	}
	tcp, exists := ports.(map[string]interface{})[port+"/tcp"]
	if !exists {
		return "", ""
	}
	list, exists := tcp.([]interface{})
	if !exists {
		return "", ""
	}

	var hostIP string
	var hostPort string
	for _, l := range list {
		data, exists := l.(map[string]interface{})
		if !exists {
			return "", ""
		}
		hostIP = data["HostIp"].(string)
		if hostIP != "::" {
			hostPort = data["HostPort"].(string)
		}
	}

	return hostIP, hostPort
}
