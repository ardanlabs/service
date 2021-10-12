package tests

import (
	"fmt"
	"testing"

	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/foundation/docker"
)

var dbc = dbtest.DBContainer{
	Image: "postgres:14-alpine",
	Port:  "5432",
	Args:  []string{"-e", "POSTGRES_PASSWORD=postgres"},
}

var c *docker.Container

func TestMain(m *testing.M) {
	var err error
	c, err = dbtest.StartDB(dbc)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbtest.StopDB(c)

	m.Run()
}
