package flag

import (
	"encoding/json"
	"testing"
	"time"
)

const (
	success = "\u2713"
	failed  = "\u2717"
)

// TestProcessNoArgs validates when no arguments are passed to the Process API.
func TestProcessNoArgs(t *testing.T) {
	var cfg struct {
		Web struct {
			APIHost     string        `default:"0.0.0.0:3000" flag:"a" flagdesc:"The ip:port for the api endpoint."`
			BatchSize   int           `default:"1000" flagdesc:"Represets number of items to move."`
			ReadTimeout time.Duration `default:"5s"`
		}
		DialTimeout time.Duration `default:"5s"`
		Host        string        `default:"mongo:27017/gotraining" flag:"h"`
	}

	t.Log("Given the need to validate was handle no arguments.")
	{
		t.Log("\tWhen there are no OS arguments.")
		{
			if err := Process(&cfg); err != nil {
				t.Fatalf("\t%s\tShould be able to call Process with no arguments : %s.", failed, err)
			}
			t.Logf("\t%s\tShould be able to call Process with no arguments.", success)
		}
	}
}

// TestParse validates the ability to reflect and parse out the argument
// metadata from the provided struct value.
func TestParse(t *testing.T) {
	var cfg struct {
		Web struct {
			APIHost     string        `default:"0.0.0.0:3000" flag:"a" flagdesc:"The ip:port for the api endpoint."`
			BatchSize   int           `default:"1000" flagdesc:"Represets number of items to move."`
			ReadTimeout time.Duration `default:"5s"`
		}
		DialTimeout time.Duration `default:"5s"`
		Host        string        `default:"mongo:27017/gotraining" flag:"h"`
		Insecure    bool          `flag:"i"`
	}
	parseOutput := `[{"Short":"a","Long":"web_apihost","Default":"0.0.0.0:3000","Type":"string","Desc":"The ip:port for the api endpoint."},{"Short":"","Long":"web_batchsize","Default":"1000","Type":"int","Desc":"Represets number of items to move."},{"Short":"","Long":"web_readtimeout","Default":"5s","Type":"Duration","Desc":""},{"Short":"","Long":"dialtimeout","Default":"5s","Type":"Duration","Desc":""},{"Short":"h","Long":"host","Default":"mongo:27017/gotraining","Type":"string","Desc":""},{"Short":"i","Long":"insecure","Default":"","Type":"bool","Desc":""}]`

	t.Log("Given the need to validate we can parse a struct value.")
	{
		t.Log("\tWhen parsing the test config.")
		{
			args, err := parse("", &cfg)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to parse arguments without error : %s.", failed, err)
			}
			t.Logf("\t%s\tShould be able to parse arguments without error.", success)

			d, _ := json.Marshal(args)
			if string(d) != parseOutput {
				t.Log("\t\tGot :", string(d))
				t.Log("\t\tWant:", parseOutput)
				t.Fatalf("\t%s\tShould get back the expected arguments.", failed)
			}
			t.Logf("\t%s\tShould get back the expected arguments.", success)
		}
	}
}

// TestApply validates the ability to apply overrides to a struct value
// based on provided flag arguments.
func TestApply(t *testing.T) {
	var cfg struct {
		Web struct {
			APIHost     string        `default:"0.0.0.0:3000" flag:"a" flagdesc:"The ip:port for the api endpoint."`
			BatchSize   int           `default:"1000" flagdesc:"Represets number of items to move."`
			ReadTimeout time.Duration `default:"5s"`
		}
		DialTimeout time.Duration `default:"5s"`
		Host        string        `default:"mongo:27017/gotraining" flag:"h"`
		Insecure    bool          `flag:"i"`
	}
	osArgs := []string{"./crud", "-i", "-a", "0.0.1.1:5000", "--web_batchsize", "300", "--dialtimeout", "10s"}
	expected := `{"Web":{"APIHost":"0.0.1.1:5000","BatchSize":300,"ReadTimeout":0},"DialTimeout":10000000000,"Host":"","Insecure":true}`

	t.Log("Given the need to validate we can apply overrides a struct value.")
	{
		t.Log("\tWhen parsing the test config.")
		{
			args, err := parse("", &cfg)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to parse arguments without error : %s.", failed, err)
			}
			t.Logf("\t%s\tShould be able to parse arguments without error.", success)

			if err := apply(osArgs, args); err != nil {
				t.Fatalf("\t%s\tShould be able to apply arguments without error : %s.", failed, err)
			}
			t.Logf("\t%s\tShould be able to apply arguments without error.", success)

			d, _ := json.Marshal(&cfg)
			if string(d) != expected {
				t.Log("\t\tGot :", string(d))
				t.Log("\t\tWant:", expected)
				t.Fatalf("\t%s\tShould get back the expected struct value.", failed)
			}
			t.Logf("\t%s\tShould get back the expected struct value.", success)
		}
	}
}

// TestApplyBad validates the ability to handle bad arguments on the command line.
func TestApplyBad(t *testing.T) {
	var cfg struct {
		Web struct {
			APIHost     string        `default:"0.0.0.0:3000" flag:"a" flagdesc:"The ip:port for the api endpoint."`
			BatchSize   int           `default:"1000" flagdesc:"Represets number of items to move."`
			ReadTimeout time.Duration `default:"5s"`
		}
		DialTimeout time.Duration `default:"5s"`
		Host        string        `default:"mongo:27017/gotraining" flag:"h"`
		Insecure    bool
	}

	tests := []struct {
		osArg []string
	}{
		{[]string{"testapp", "-help"}},
		{[]string{"testapp", "-bad", "value"}},
		{[]string{"testapp", "-insecure", "value"}},
	}

	t.Log("Given the need to validate we can parse a struct value with bad OS arguments.")
	{
		for i, tt := range tests {
			t.Logf("\tTest: %d\tWhen checking %v", i, tt.osArg)
			{
				args, err := parse("", &cfg)
				if err != nil {
					t.Fatalf("\t%s\tShould be able to parse arguments without error : %s.", failed, err)
				}
				t.Logf("\t%s\tShould be able to parse arguments without error.", success)

				if err := apply(tt.osArg, args); err != nil {
					t.Logf("\t%s\tShould not be able to apply arguments.", success)
				} else {
					t.Errorf("\t%s\tShould not be able to apply arguments.", failed)
				}
			}
		}
	}
}

// TestDisplay provides a test for displaying the command line arguments.
func TestDisplay(t *testing.T) {
	var cfg struct {
		Web struct {
			APIHost     string        `default:"0.0.0.0:3000" flag:"a" flagdesc:"The ip:port for the api endpoint."`
			BatchSize   int           `default:"1000" flagdesc:"Represets number of items to move."`
			ReadTimeout time.Duration `default:"5s"`
		}
		DialTimeout time.Duration `default:"5s"`
		Host        string        `default:"mongo:27017/gotraining" flag:"h"`
		Insecure    bool          `flag:"i"`
	}

	want := `
Useage of TestApp
-a --web_apihost string  <0.0.0.0:3000> : The ip:port for the api endpoint.
--web_batchsize int  <1000> : Represets number of items to move.
--web_readtimeout Duration  <5s>
--dialtimeout Duration  <5s>
-h --host string  <mongo:27017/gotraining>
-i --insecure bool
`

	got := display("TestApp", &cfg)
	if got != want {
		t.Log("\t\tGot :", []byte(got))
		t.Log("\t\tWant:", []byte(want))
		t.Fatalf("\t%s\tShould get back the expected help output.", failed)
	}
	t.Logf("\t%s\tShould get back the expected help output.", success)
}
