package conf_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ardanlabs/service/internal/platform/conf"
	"github.com/google/go-cmp/cmp"
)

const (
	success = "\u2713"
	failed  = "\u2717"
)

type ip struct {
	Name string `conf:"default:localhost,env:IP_NAME_VAR"`
	IP   string `conf:"default:127.0.0.0"`
}
type Embed struct {
	Name     string        `conf:"default:bill"`
	Duration time.Duration `conf:"default:1s,flag:e-dur,short:d"`
}
type config struct {
	AnInt   int    `conf:"default:9"`
	AString string `conf:"default:B,short:s"`
	Bool    bool
	Skip    string `conf:"-"`
	IP      ip
	Embed
}

// =============================================================================

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		args []string
		want config
	}{
		{
			"default",
			nil,
			nil,
			config{9, "B", false, "", ip{"localhost", "127.0.0.0"}, Embed{"bill", time.Second}},
		},
		{
			"env",
			map[string]string{"TEST_AN_INT": "1", "TEST_A_STRING": "s", "TEST_BOOL": "TRUE", "TEST_SKIP": "SKIP", "TEST_IP_NAME_VAR": "local", "TEST_NAME": "andy", "TEST_DURATION": "1m"},
			nil,
			config{1, "s", true, "", ip{"local", "127.0.0.0"}, Embed{"andy", time.Minute}},
		},
		{
			"flag",
			nil,
			[]string{"--an-int", "1", "-s", "s", "--bool", "--skip", "skip", "--ip-name", "local", "--name", "andy", "--e-dur", "1m"},
			config{1, "s", true, "", ip{"local", "127.0.0.0"}, Embed{"andy", time.Minute}},
		},
		{
			"multi",
			map[string]string{"TEST_A_STRING": "s", "TEST_BOOL": "TRUE", "TEST_IP_NAME_VAR": "local", "TEST_NAME": "andy", "TEST_DURATION": "1m"},
			[]string{"--an-int", "2", "--bool", "--skip", "skip", "--name", "jack", "-d", "1ms"},
			config{2, "s", true, "", ip{"local", "127.0.0.0"}, Embed{"jack", time.Millisecond}},
		},
	}

	t.Log("Given the need to parse basic configuration.")
	{
		for i, tt := range tests {
			t.Logf("\tTest: %d\tWhen checking with arguments %v", i, tt.args)
			{
				os.Clearenv()
				for k, v := range tt.envs {
					os.Setenv(k, v)
				}

				f := func(t *testing.T) {
					var cfg config
					if err := conf.Parse(tt.args, "TEST", &cfg); err != nil {
						t.Fatalf("\t%s\tShould be able to Parse arguments : %s.", failed, err)
					}
					t.Logf("\t%s\tShould be able to Parse arguments.", success)

					if diff := cmp.Diff(tt.want, cfg); diff != "" {
						t.Fatalf("\t%s\tShould have properly initialized struct value\n%s", failed, diff)
					}
					t.Logf("\t%s\tShould have properly initialized struct value.", success)
				}

				t.Run(tt.name, f)
			}
		}
	}
}

func TestParse_Args(t *testing.T) {
	t.Log("Given the need to capture remaining command line arguments after flags.")
	{
		type configArgs struct {
			Port int
			Args conf.Args
		}

		args := []string{"--port", "9000", "migrate", "seed"}

		want := configArgs{
			Port: 9000,
			Args: conf.Args{"migrate", "seed"},
		}

		var cfg configArgs
		if err := conf.Parse(args, "TEST", &cfg); err != nil {
			t.Fatalf("\t%s\tShould be able to Parse arguments : %s.", failed, err)
		}
		t.Logf("\t%s\tShould be able to Parse arguments.", success)

		if diff := cmp.Diff(want, cfg); diff != "" {
			t.Fatalf("\t%s\tShould have properly initialized struct value\n%s", failed, diff)
		}
		t.Logf("\t%s\tShould have properly initialized struct value.", success)
	}
}

func TestErrors(t *testing.T) {
	t.Log("Given the need to validate errors that can occur with Parse.")
	{
		t.Logf("\tTest: %d\tWhen passing bad values to Parse.", 0)
		{
			f := func(t *testing.T) {
				var cfg struct {
					TestInt    int
					TestString string
					TestBool   bool
				}
				err := conf.Parse(nil, "TEST", cfg)
				if err == nil {
					t.Fatalf("\t%s\tShould NOT be able to accept a value by value.", failed)
				}
				t.Logf("\t%s\tShould NOT be able to accept a value by value : %s", success, err)
			}
			t.Run("not-by-ref", f)

			f = func(t *testing.T) {
				var cfg []string
				err := conf.Parse(nil, "TEST", &cfg)
				if err == nil {
					t.Fatalf("\t%s\tShould NOT be able to pass anything but a struct value.", failed)
				}
				t.Logf("\t%s\tShould NOT be able to pass anything but a struct value : %s", success, err)
			}
			t.Run("not-struct-value", f)
		}

		t.Logf("\tTest: %d\tWhen bad tags to Parse.", 1)
		{
			f := func(t *testing.T) {
				var cfg struct {
					TestInt    int `conf:"default:"`
					TestString string
					TestBool   bool
				}
				err := conf.Parse(nil, "TEST", &cfg)
				if err == nil {
					t.Fatalf("\t%s\tShould NOT be able to accept tag missing value.", failed)
				}
				t.Logf("\t%s\tShould NOT be able to accept tag missing value : %s", success, err)
			}
			t.Run("tag-missing-value", f)

			f = func(t *testing.T) {
				var cfg struct {
					TestInt    int `conf:"short:ab"`
					TestString string
					TestBool   bool
				}
				err := conf.Parse(nil, "TEST", &cfg)
				if err == nil {
					t.Fatalf("\t%s\tShould NOT be able to accept invalid short tag.", failed)
				}
				t.Logf("\t%s\tShould NOT be able to accept invalid short tag : %s", success, err)
			}
			t.Run("tag-bad-short", f)
		}

		t.Logf("\tTest: %d\tWhen required values are missing.", 2)
		{
			f := func(t *testing.T) {
				var cfg struct {
					TestInt    int `conf:"required, default:1"`
					TestString string
					TestBool   bool
				}
				err := conf.Parse(nil, "TEST", &cfg)
				if err == nil {
					t.Fatalf("\t%s\tShould fail for missing required value.", failed)
				}
				t.Logf("\t%s\tShould fail for missing required value : %s", success, err)
			}
			t.Run("required-missing-value", f)
		}

		t.Logf("\tTest: %d\tWhen struct has no fields.", 2)
		{
			f := func(t *testing.T) {
				var cfg struct {
					testInt    int `conf:"required, default:1"`
					testString string
					testBool   bool
				}
				err := conf.Parse(nil, "TEST", &cfg)
				if err == nil {
					t.Fatalf("\t%s\tShould fail for struct with no exported fields.", failed)
				}
				t.Logf("\t%s\tShould fail for struct with no exported fields : %s", success, err)
			}
			t.Run("struct-missing-fields", f)
		}
	}
}

func TestUsage(t *testing.T) {
	tt := struct {
		name string
		envs map[string]string
	}{
		name: "one-example",
		envs: map[string]string{"TEST_AN_INT": "1", "TEST_A_STRING": "s", "TEST_BOOL": "TRUE", "TEST_SKIP": "SKIP", "TEST_IP_NAME_VAR": "local", "TEST_NAME": "andy", "TEST_DURATION": "1m"},
	}

	t.Log("Given the need validate usage output.")
	{
		t.Logf("\tTest: %d\tWhen using a basic struct.", 0)
		{
			os.Clearenv()
			for k, v := range tt.envs {
				os.Setenv(k, v)
			}

			var cfg config
			if err := conf.Parse(nil, "TEST", &cfg); err != nil {
				fmt.Print(err)
				return
			}

			got, err := conf.Usage("TEST", &cfg)
			if err != nil {
				fmt.Print(err)
				return
			}

			got = strings.TrimRight(got, " \n")
			want := `Usage: conf.test [options] [arguments]

OPTIONS
  --an-int/$TEST_AN_INT         <int>       (default: 9)
  --a-string/-s/$TEST_A_STRING  <string>    (default: B)
  --bool/$TEST_BOOL             <bool>      
  --ip-name/$TEST_IP_NAME_VAR   <string>    (default: localhost)
  --ip-ip/$TEST_IP_IP           <string>    (default: 127.0.0.0)
  --name/$TEST_NAME             <string>    (default: bill)
  --e-dur/-d/$TEST_DURATION     <duration>  (default: 1s)
  --help/-h                     
  display this help message`

			gotS := strings.Split(got, "\n")
			wantS := strings.Split(want, "\n")
			if diff := cmp.Diff(gotS, wantS); diff != "" {
				t.Errorf("\t%s\tShould match the output byte for byte. See diff:", failed)
				t.Log(diff)
			}
			t.Logf("\t%s\tShould match byte for byte the output.", success)
		}

		t.Logf("\tTest: %d\tWhen using a struct with arguments.", 1)
		{
			var cfg struct {
				Port int
				Args conf.Args
			}

			got, err := conf.Usage("TEST", &cfg)
			if err != nil {
				fmt.Print(err)
				return
			}

			got = strings.TrimRight(got, " \n")
			want := `Usage: conf.test [options] [arguments]

OPTIONS
  --port/$TEST_PORT  <int>
  --help/-h              
  display this help message`

			gotS := strings.Split(got, "\n")
			wantS := strings.Split(want, "\n")
			if diff := cmp.Diff(gotS, wantS); diff != "" {
				t.Errorf("\t%s\tShould match the output byte for byte. See diff:", failed)
				t.Log(diff)
			}
			t.Logf("\t%s\tShould match byte for byte the output.", success)
		}
	}
}

func ExampleString() {
	tt := struct {
		name string
		envs map[string]string
	}{
		name: "one-example",
		envs: map[string]string{"TEST_AN_INT": "1", "TEST_S": "s", "TEST_BOOL": "TRUE", "TEST_SKIP": "SKIP", "TEST_IP_NAME": "local", "TEST_NAME": "andy", "TEST_DURATION": "1m"},
	}

	os.Clearenv()
	for k, v := range tt.envs {
		os.Setenv(k, v)
	}

	var cfg config
	if err := conf.Parse(nil, "TEST", &cfg); err != nil {
		fmt.Print(err)
		return
	}

	out, err := conf.String(&cfg)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Print(out)

	// Output:
	// --an-int=1
	// --a-string/-s=B
	// --bool=true
	// --ip-name=localhost
	// --ip-ip=127.0.0.0
	// --name=andy
	// --e-dur/-d=1m0s
}
