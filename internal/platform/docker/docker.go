package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

// Container contains the information about the conainer.
type Container struct {
	ID   string
	Port string
}

// StartMongo runs a mongo container to execute commands.
func StartMongo() (*Container, error) {
	cmd := exec.Command("docker", "run", "-P", "-d", "mongo")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("starting container: %v", err)
	}

	id := out.String()[:3]
	log.Println("DB ContainerID:", id)

	cmd.Wait()

	cmd = exec.Command("docker", "inspect", string(id))
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("inspect container: %v", err)
	}

	cmd.Wait()

	var doc []map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		return nil, fmt.Errorf("decoding json: %v", err)
	}

	c := Container{
		ID:   id,
		Port: doc[0]["NetworkSettings"].(map[string]interface{})["Ports"].(map[string]interface{})["27017/tcp"].([]interface{})[0].(map[string]interface{})["HostPort"].(string),
	}

	log.Println("DB Port:", c.Port)

	return &c, nil
}

// StopMongo stops and removes the specified container.
func StopMongo(c *Container) error {
	if err := exec.Command("docker", "stop", c.ID).Run(); err != nil {
		return err
	}
	log.Println("Stopped:", c.ID)

	if err := exec.Command("docker", "rm", c.ID, "-v").Run(); err != nil {
		return err
	}
	log.Println("Removed:", c.ID)

	return nil
}
