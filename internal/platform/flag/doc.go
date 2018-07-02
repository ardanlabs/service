/*
Package flag is compatible with the GNU extensions to the POSIX recommendations
for command-line options. See
http://www.gnu.org/software/libc/manual/html_node/Argument-Syntax.html

There is no hard bindings for this package. This package takes a struct
value and parses it for flags. It supports three tags to customize the
flag options.

	flag     - Denotes a shorthand option
	flagdesc - Provides a description for the help
	default  - Provides the default value for the help

The field name and any parent struct name will be used for the long form of
the command name.

As an example, this config struct:

	var cfg struct {
		Web struct {
			APIHost     string        `default:"0.0.0.0:3000" flag:"a" flagdesc:"The ip:port for the api endpoint."`
			BatchSize   int           `default:"1000" flagdesc:"Represets number of items to move."`
			ReadTimeout time.Duration `default:"5s"`
		}
		DialTimeout time.Duration `default:"5s"`
		Host        string        `default:"mongo:27017/gotraining" flag:"h"`
	}

Would produce the following flag output:

	-a --web_apihost string  <0.0.0.0:3000> : The ip:port for the api endpoint.
	--web_batchsize int  <1000> : Represets number of items to move.
	--web_readtimeout Duration  <5s>
	--dialtimeout Duration  <5s>
	-h --host string  <mongo:27017/gotraining>

The command line flag syntax assumes a regular or shorthand version based on the
type of dash used.
	Regular versions
	--flag=x
	--flag x

	Shorthand versions
	-f=x
	-f x
*/
package flag
