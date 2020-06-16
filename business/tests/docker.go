package tests

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"testing"
)

// DBContainer tracks information about the DB docker container started for tests.
type DBContainer struct {
	ID     string
	DBHost string // IP:Port
}

func startDBContainer(t *testing.T, image string) *DBContainer {
	cmd := exec.Command("docker", "run", "-P", "-d", image)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not start container %s: %v", image, err)
	}

	id := out.String()[:12]

	cmd = exec.Command("docker", "inspect", id)
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not inspect container %s: %v", id, err)
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
		t.Fatalf("could not decode json: %v", err)
	}

	dbHost := doc[0].NetworkSettings.Ports.TCP5432[0]

	c := DBContainer{
		ID:     id,
		DBHost: dbHost.HostIP + ":" + dbHost.HostPort,
	}

	t.Logf("Image:          %s", image)
	t.Logf("DB ContainerID: %s", c.ID)
	t.Logf("DB Host:        %s", c.DBHost)

	return &c
}

func stopContainer(t *testing.T, id string) {
	if err := exec.Command("docker", "stop", id).Run(); err != nil {
		t.Fatalf("could not stop container: %v", err)
	}
	t.Log("Stopped:", id)

	if err := exec.Command("docker", "rm", id, "-v").Run(); err != nil {
		t.Fatalf("could not remove container: %v", err)
	}
	t.Log("Removed:", id)
}

func dumpContainerLogs(t *testing.T, id string) {
	out, err := exec.Command("docker", "logs", id).CombinedOutput()
	if err != nil {
		t.Fatalf("could not log container: %v", err)
	}
	t.Logf("Logs for %s\n%s:", id, out)
}
