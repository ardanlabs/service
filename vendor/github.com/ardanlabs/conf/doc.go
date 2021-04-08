/*
Package conf provides support for using environmental variables and command
line arguments for configuration.

It is compatible with the GNU extensions to the POSIX recommendations
for command-line options. See
http://www.gnu.org/software/libc/manual/html_node/Argument-Syntax.html

There are no hard bindings for this package. This package takes a struct
value and parses it for both the environment and flags. It supports several tags
to customize the flag options.

	default  - Provides the default value for the help
	env      - Allows for overriding the default variable name.
	flag     - Allows for overriding the default flag name.
	short    - Denotes a shorthand option for the flag.
	noprint  - Denotes to not include the field in any display string.
	mask     - Includes the field in any display string but masks out the value.
	required - Denotes a value must be provided.
	help     - Provides a description for the help.

The field name and any parent struct name will be used for the long form of
the command name unless the name is overridden.

As an example, this config struct:

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

Would produce the following usage output:

Usage: conf.test [options] [arguments]

OPTIONS
  --an-int/$CRUD_AN_INT         <int>       (default: 9)
  --a-string/-s/$CRUD_A_STRING  <string>    (default: B)
  --bool/$CRUD_BOOL             <bool>
  --ip-name/$CRUD_IP_NAME_VAR   <string>    (default: localhost)
  --ip-ip/$CRUD_IP_IP           <string>    (default: 127.0.0.0)
  --name/$CRUD_NAME             <string>    (default: bill)
  --e-dur/-d/$CRUD_DURATION     <duration>  (default: 1s)
  --help/-h
  display this help message
  --version/-v
  display version information

The API is a single call to Parse

	// Parse(args []string, namespace string, cfgStruct interface{}, sources ...Sourcer) error

	if err := conf.Parse(os.Args, "CRUD", &cfg); err != nil {
		log.Fatalf("main : Parsing Config : %v", err)
	}

Additionally, if the config struct has a field of the slice type conf.Args
then it will be populated with any remaining arguments from the command line
after flags have been processed.

For example a program with a config struct like this:

	var cfg struct {
		Port int
		Args conf.Args
	}

If that program is executed from the command line like this:

	$ my-program --port=9000 serve http

Then the cfg.Args field will contain the string values ["serve", "http"].
The Args type has a method Num for convenient access to these arguments
such as this:

	arg0 := cfg.Args.Num(0) // "serve"
	arg1 := cfg.Args.Num(1) // "http"
	arg2 := cfg.Args.Num(2) // "" empty string: not enough arguments

You can add a version with a description by adding the Version type to
your config type

	type ConfExplicit struct {
		Version conf.Version
		Address string
	}

	type ConfImplicit struct {
		conf.Version
		Address string
	}

Then you can set these values at run time for display.

	cfg := struct {
		Version conf.Version
	}{
		Version: conf.Version{
			SVN:  "v1.0.0",
			Desc: "Service Description",
		},
	}

	if err := conf.Parse(os.Args[1:], "APP", &cfg); err != nil {
		if err == conf.ErrVersionWanted {
			version, err := conf.VersionString("APP", &cfg)
			if err != nil {
				return err
			}
			fmt.Println(version)
			return nil
		}
		fmt.Println("parsing config", err)
	}
*/
package conf
