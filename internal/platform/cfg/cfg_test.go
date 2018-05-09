package cfg_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/ardanlabs/service/internal/platform/cfg"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

// TestExists validates the ability to load configuration values
// using the OS-level environment variables and read them back.
func TestExists(t *testing.T) {
	t.Log("Given the need to read environment variables.")
	{
		uStr := "postgres://root:root@127.0.0.1:8080/postgres?sslmode=disable"
		dStr := 2 * time.Minute

		cfg.Init(cfg.MapProvider{
			Map: map[string]string{
				"PROC_ID": "322",
				"SOCKET":  "./tmp/sockets.po",
				"PORT":    "4034",
				"FLAG":    "on",
				"DSN":     uStr,
				"TIMEOUT": dStr.String(),
			},
		})

		t.Log("\tWhen given a namespace key to search for that exists.")
		{
			proc, err := cfg.Int("PROC_ID")

			if err != nil {
				t.Errorf("\t\t%s Should not return error when valid key %q", failed, "PROC_ID")
			} else {
				t.Logf("\t\t%s Should not return error when valid key %q", success, "PROC_ID")

				if proc != 322 {
					t.Errorf("\t\t%s Should have key %q with value %d", failed, "PROC_ID", 322)
				} else {
					t.Logf("\t\t%s Should have key %q with value %d", success, "PROC_ID", 322)
				}
			}

			shouldNotPanic(t, "PROC_ID", func() {
				cfg.MustInt("PROC_ID")
			})

			socket, err := cfg.String("SOCKET")

			if err != nil {
				t.Errorf("\t\t%s Should not return error when valid key %q", failed, "SOCKET")
			} else {
				t.Logf("\t\t%s Should not return error when valid key %q", success, "SOCKET")

				if socket != "./tmp/sockets.po" {
					t.Errorf("\t\t%s Should have key %q with value %q", failed, "SOCKET", "./tmp/sockets.po")
				} else {
					t.Logf("\t\t%s Should have key %q with value %q", success, "SOCKET", "./tmp/sockets.po")
				}
			}

			shouldNotPanic(t, "SOCKET", func() {
				cfg.MustString("SOCKET")
			})

			port, err := cfg.Int("PORT")

			if err != nil {
				t.Errorf("\t\t%s Should not return error when valid key %q", failed, "PORT")
			} else {
				t.Logf("\t\t%s Should not return error when valid key %q", success, "PORT")

				if port != 4034 {
					t.Errorf("\t\t%s Should have key %q with value %d", failed, "PORT", 4034)
				} else {
					t.Logf("\t\t%s Should have key %q with value %d", success, "PORT", 4034)
				}
			}

			shouldNotPanic(t, "PORT", func() {
				cfg.MustInt("PORT")
			})

			flag, err := cfg.Bool("FLAG")

			if err != nil {
				t.Errorf("\t\t%s Should not return error when valid key %q", failed, "FLAG")
			} else {
				t.Logf("\t\t%s Should not return error when valid key %q", success, "FLAG")

				if !flag {
					t.Errorf("\t\t%s Should have key %q with value %v", failed, "FLAG", true)
				} else {
					t.Logf("\t\t%s Should have key %q with value %v", success, "FLAG", true)
				}
			}

			shouldNotPanic(t, "FLAG", func() {
				cfg.MustBool("FLAG")
			})

			u, err := cfg.URL("DSN")

			if err != nil {
				t.Errorf("\t\t%s Should not return error when valid key %q", failed, "DSN")
			} else {
				t.Logf("\t\t%s Should not return error when valid key %q", success, "DSN")

				if u.String() != uStr {
					t.Errorf("\t\t%s Should have key %q with value %v", failed, "DSN", uStr)
				} else {
					t.Logf("\t\t%s Should have key %q with value %v", success, "DSN", uStr)
				}
			}

			shouldNotPanic(t, "DSN", func() {
				cfg.MustURL("DSN")
			})

			d, err := cfg.Duration("TIMEOUT")

			if err != nil {
				t.Errorf("\t\t%s Should not return error when valid key %q", failed, "TIMEOUT")
			} else {
				t.Logf("\t\t%s Should not return error when valid key %q", success, "TIMEOUT")

				if d != dStr {
					t.Errorf("\t\t%s Should have key %q with value %v", failed, "TIMEOUT", dStr)
				} else {
					t.Logf("\t\t%s Should have key %q with value %v", success, "TIMEOUT", dStr)
				}
			}

			shouldNotPanic(t, "TIMEOUT", func() {
				cfg.MustDuration("TIMEOUT")
			})
		}
	}
}

// TestNotExists validates the ability to load configuration values
// using the OS-level environment variables and panic when something
// is missing.
func TestNotExists(t *testing.T) {
	t.Log("Given the need to panic when environment variables are missing.")
	{
		cfg.Init(cfg.MapProvider{
			Map: map[string]string{},
		})

		t.Log("\tWhen given a namspace key to search for that does NOT exist.")
		{
			shouldPanic(t, "STAMP", func() {
				cfg.MustTime("STAMP")
			})

			if _, err := cfg.Time("STAMP"); err == nil {
				t.Errorf("\t\t%s Should have error when giving unknown key %q.", failed, "STAMP")
			} else {
				t.Logf("\t\t%s Should have error when giving unknown key %q.", success, "STAMP")
			}

			shouldPanic(t, "PID", func() {
				cfg.MustInt("PID")
			})

			if _, err := cfg.Int("PID"); err == nil {
				t.Errorf("\t\t%s Should have error when giving unknown key %q.", failed, "PID")
			} else {
				t.Logf("\t\t%s Should have error when giving unknown key %q.", success, "PID")
			}

			shouldPanic(t, "DEST", func() {
				cfg.MustString("DEST")
			})

			if _, err := cfg.String("DEST"); err == nil {
				t.Errorf("\t\t%s Should have error when giving unknown key %q.", failed, "DEST")
			} else {
				t.Logf("\t\t%s Should have error when giving unknown key %q.", success, "DEST")
			}

			shouldPanic(t, "ACTIVE", func() {
				cfg.MustBool("ACTIVE")
			})

			if _, err := cfg.Bool("ACTIVE"); err == nil {
				t.Errorf("\t\t%s Should have error when giving unknown key %q.", failed, "ACTIVE")
			} else {
				t.Logf("\t\t%s Should have error when giving unknown key %q.", success, "ACTIVE")
			}

			shouldPanic(t, "SOCKET_DSN", func() {
				cfg.MustURL("SOCKET_DSN")
			})

			if _, err := cfg.URL("SOCKET_DSN"); err == nil {
				t.Errorf("\t\t%s Should have error when giving unknown key %q.", failed, "SOCKET_DSN")
			} else {
				t.Logf("\t\t%s Should have error when giving unknown key %q.", success, "SOCKET_DSN")
			}

			shouldPanic(t, "TIMEOUT", func() {
				cfg.MustDuration("TIMEOUT")
			})

			if _, err := cfg.Duration("TIMEOUT"); err == nil {
				t.Errorf("\t\t%s Should have error when giving unknown key %q.", failed, "TIMEOUT")
			} else {
				t.Logf("\t\t%s Should have error when giving unknown key %q.", success, "TIMEOUT")
			}
		}
	}
}

// shouldPanic receives a ctx string and a function to run, if the function
// panics, it is considered a success else a failure.
func shouldPanic(t *testing.T, ctx string, fx func()) {
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("\t\t%s Should paniced when giving unknown key %q.", failed, ctx)
		} else {
			t.Logf("\t\t%s Should paniced when giving unknown key %q.", success, ctx)
		}
	}()

	fx()
}

// shouldNotPanic receives a ctx string and a function to run, if the function
// panics, it is considered a failure else a success.
func shouldNotPanic(t *testing.T, ctx string, fx func()) {
	defer func() {
		if err := recover(); err == nil {
			t.Logf("\t\t%s Should not have paniced when giving known key %q.", success, ctx)
		} else {
			t.Errorf("\t\t%s Should not have paniced when giving known key %q.", failed, ctx)
		}
	}()

	fx()
}

// TestSets validates the ability to manually set configuration values.
func TestSets(t *testing.T) {
	t.Log("Given the need to manually set configuration values.")
	{
		cfg.Init(cfg.MapProvider{
			Map: map[string]string{},
		})

		t.Log("\tWhen setting values.")
		{
			key := "key1"
			strVal := "bill"
			cfg.SetString(key, strVal)

			retStrVal, err := cfg.String(key)
			if err != nil {
				t.Errorf("\t\t%s Should find a value for the specified key %q.", failed, key)
			} else {
				t.Logf("\t\t%s Should find a value for the specified key %q.", success, key)
			}
			if strVal != retStrVal {
				t.Log(strVal)
				t.Log(retStrVal)
				t.Errorf("\t\t%s Should return the string value %q that was set.", failed, strVal)
			} else {
				t.Logf("\t\t%s Should return the string value %q that was set.", success, strVal)
			}

			key = "key2"
			intVal := 223
			cfg.SetInt(key, intVal)

			retIntVal, err := cfg.Int(key)
			if err != nil {
				t.Errorf("\t\t%s Should find a value for the specified key %q.", failed, key)
			} else {
				t.Logf("\t\t%s Should find a value for the specified key %q.", success, key)
			}
			if intVal != retIntVal {
				t.Log(intVal)
				t.Log(retIntVal)
				t.Errorf("\t\t%s Should return the int value %d that was set.", failed, intVal)
			} else {
				t.Logf("\t\t%s Should return the int value %d that was set.", success, intVal)
			}

			key = "key3"
			timeVal, _ := time.Parse(time.UnixDate, "Mon Oct 27 20:18:15 EST 2016")
			cfg.SetTime(key, timeVal)

			retTimeVal, err := cfg.Time(key)
			if err != nil {
				t.Errorf("\t\t%s Should find a value for the specified key %q.", failed, key)
			} else {
				t.Logf("\t\t%s Should find a value for the specified key %q.", success, key)
			}
			if timeVal.UnixNano() != retTimeVal.UnixNano() {
				t.Log(timeVal)
				t.Log(retTimeVal)
				t.Errorf("\t\t%s Should return the time value %q that was set.", failed, timeVal)
			} else {
				t.Logf("\t\t%s Should return the time value %q that was set.", success, timeVal)
			}

			key = "key4"
			boolVal := true
			cfg.SetBool(key, boolVal)

			retBoolVal, err := cfg.Bool(key)
			if err != nil {
				t.Errorf("\t\t%s Should find a value for the specified key %q.", failed, key)
			} else {
				t.Logf("\t\t%s Should find a value for the specified key %q.", success, key)
			}
			if boolVal != retBoolVal {
				t.Log(boolVal)
				t.Log(retBoolVal)
				t.Errorf("\t\t%s Should return the bool value \"%v\" that was set.", failed, boolVal)
			} else {
				t.Logf("\t\t%s Should return the bool value \"%v\" that was set.", success, boolVal)
			}

			key = "key5"
			urlVal, _ := url.Parse("postgres://root:root@127.0.0.1:8080/postgres?sslmode=disable")
			cfg.SetURL(key, urlVal)

			retURLVal, err := cfg.URL(key)
			if err != nil {
				t.Errorf("\t\t%s Should find a value for the specified key %q.", failed, key)
			} else {
				t.Logf("\t\t%s Should find a value for the specified key %q.", success, key)
			}
			if urlVal.String() != retURLVal.String() {
				t.Log(urlVal)
				t.Log(retURLVal)
				t.Errorf("\t\t%s Should return the url value \"%v\" that was set.", failed, urlVal)
			} else {
				t.Logf("\t\t%s Should return the url value \"%v\" that was set.", success, urlVal)
			}

			key = "key6"
			dVal := 2 * time.Minute
			cfg.SetDuration(key, dVal)

			retDurVal, err := cfg.Duration(key)
			if err != nil {
				t.Errorf("\t\t%s Should find a value for the specified key %q.", failed, key)
			} else {
				t.Logf("\t\t%s Should find a value for the specified key %q.", success, key)
			}
			if dVal != retDurVal {
				t.Log(dVal)
				t.Log(retDurVal)
				t.Errorf("\t\t%s Should return the bool value \"%v\" that was set.", failed, dVal)
			} else {
				t.Logf("\t\t%s Should return the bool value \"%v\" that was set.", success, dVal)
			}
		}
	}
}

// TestNew validates the ability to create new instances of Config.
func TestNew(t *testing.T) {
	t.Log("Given the need to create a new instance of Config.")
	{
		t.Log("\tWhen instantiating configs")
		{
			uStr := "postgres://root:root@127.0.0.1:8080/postgres?sslmode=disable"
			dStr := 2 * time.Minute
			c, err := cfg.New(cfg.MapProvider{
				Map: map[string]string{
					"PROC_ID": "322",
					"SOCKET":  "./tmp/sockets.po",
					"PORT":    "4034",
					"FLAG":    "on",
					"DSN":     uStr,
					"TIMEOUT": dStr.String(),
				},
			})
			if err != nil {
				t.Fatalf("\t\t%s Should not return an error.", failed)
			} else {
				t.Logf("\t\t%s Should not return an error.", success)
			}

			proc, err := c.Int("PROC_ID")

			if err != nil {
				t.Errorf("\t\t%s Should not return error when valid key %q", failed, "PROC_ID")
			} else {
				t.Logf("\t\t%s Should not return error when valid key %q", success, "PROC_ID")

				if proc != 322 {
					t.Errorf("\t\t%s Should have key %q with value %d", failed, "PROC_ID", 322)
				} else {
					t.Logf("\t\t%s Should have key %q with value %d", success, "PROC_ID", 322)
				}
			}

			socket, err := c.String("SOCKET")

			if err != nil {
				t.Errorf("\t\t%s Should not return error when valid key %q", failed, "SOCKET")
			} else {
				t.Logf("\t\t%s Should not return error when valid key %q", success, "SOCKET")

				if socket != "./tmp/sockets.po" {
					t.Errorf("\t\t%s Should have key %q with value %q", failed, "SOCKET", "./tmp/sockets.po")
				} else {
					t.Logf("\t\t%s Should have key %q with value %q", success, "SOCKET", "./tmp/sockets.po")
				}
			}

			port, err := c.Int("PORT")

			if err != nil {
				t.Errorf("\t\t%s Should not return error when valid key %q", failed, "PORT")
			} else {
				t.Logf("\t\t%s Should not return error when valid key %q", success, "PORT")

				if port != 4034 {
					t.Errorf("\t\t%s Should have key %q with value %d", failed, "PORT", 4034)
				} else {
					t.Logf("\t\t%s Should have key %q with value %d", success, "PORT", 4034)
				}
			}

			flag, err := c.Bool("FLAG")

			if err != nil {
				t.Errorf("\t\t%s Should not return error when valid key %q", failed, "FLAG")
			} else {
				t.Logf("\t\t%s Should not return error when valid key %q", success, "FLAG")

				if !flag {
					t.Errorf("\t\t%s Should have key %q with value %v", failed, "FLAG", true)
				} else {
					t.Logf("\t\t%s Should have key %q with value %v", success, "FLAG", true)
				}
			}

			u, err := c.URL("DSN")

			if err != nil {
				t.Errorf("\t\t%s Should not return error when valid key %q", failed, "DSN")
			} else {
				t.Logf("\t\t%s Should not return error when valid key %q", success, "DSN")

				if u.String() != uStr {
					t.Errorf("\t\t%s Should have key %q with value %v", failed, "DSN", uStr)
				} else {
					t.Logf("\t\t%s Should have key %q with value %v", success, "DSN", uStr)
				}
			}

			d, err := c.Duration("TIMEOUT")

			if err != nil {
				t.Errorf("\t\t%s Should not return error when valid key %q", failed, "TIMEOUT")
			} else {
				t.Logf("\t\t%s Should not return error when valid key %q", success, "TIMEOUT")

				if d != dStr {
					t.Errorf("\t\t%s Should have key %q with value %v", failed, "TIMEOUT", dStr)
				} else {
					t.Logf("\t\t%s Should have key %q with value %v", success, "TIMEOUT", dStr)
				}
			}
		}
	}
}
