

# cfg
`import "github.com/ardanlabs/kit/cfg"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)
* [Examples](#pkg-examples)

## <a name="pkg-overview">Overview</a>
Package cfg provides configuration options that are loaded from the environment.
Configuration is then stored in memory and can be retrieved by its proper
type.

To initialize the configuration system from your environment, call Init:


	cfg.Init(cfg.EnvProvider{Namespace: "configKey"})

To retrieve values from configuration:


	proc, err := cfg.String("proc_id")
	port, err := cfg.Int("port")
	ms, err := cfg.Time("stamp")

Use the Must set of function to retrieve a single value but these calls
will panic if the key does not exist:


	proc := cfg.MustString("proc_id")
	port := cfg.MustInt("port")
	ms := cfg.MustTime("stamp")




## <a name="pkg-index">Index</a>
* [func Bool(key string) (bool, error)](#Bool)
* [func Duration(key string) (time.Duration, error)](#Duration)
* [func Init(p Provider) error](#Init)
* [func Int(key string) (int, error)](#Int)
* [func Log() string](#Log)
* [func MustBool(key string) bool](#MustBool)
* [func MustDuration(key string) time.Duration](#MustDuration)
* [func MustInt(key string) int](#MustInt)
* [func MustString(key string) string](#MustString)
* [func MustTime(key string) time.Time](#MustTime)
* [func MustURL(key string) *url.URL](#MustURL)
* [func SetBool(key string, value bool)](#SetBool)
* [func SetDuration(key string, value time.Duration)](#SetDuration)
* [func SetInt(key string, value int)](#SetInt)
* [func SetString(key string, value string)](#SetString)
* [func SetTime(key string, value time.Time)](#SetTime)
* [func SetURL(key string, value *url.URL)](#SetURL)
* [func String(key string) (string, error)](#String)
* [func Time(key string) (time.Time, error)](#Time)
* [func URL(key string) (*url.URL, error)](#URL)
* [type Config](#Config)
  * [func New(p Provider) (*Config, error)](#New)
  * [func (c *Config) Bool(key string) (bool, error)](#Config.Bool)
  * [func (c *Config) Duration(key string) (time.Duration, error)](#Config.Duration)
  * [func (c *Config) Int(key string) (int, error)](#Config.Int)
  * [func (c *Config) Log() string](#Config.Log)
  * [func (c *Config) MustBool(key string) bool](#Config.MustBool)
  * [func (c *Config) MustDuration(key string) time.Duration](#Config.MustDuration)
  * [func (c *Config) MustInt(key string) int](#Config.MustInt)
  * [func (c *Config) MustString(key string) string](#Config.MustString)
  * [func (c *Config) MustTime(key string) time.Time](#Config.MustTime)
  * [func (c *Config) MustURL(key string) *url.URL](#Config.MustURL)
  * [func (c *Config) SetBool(key string, value bool)](#Config.SetBool)
  * [func (c *Config) SetDuration(key string, value time.Duration)](#Config.SetDuration)
  * [func (c *Config) SetInt(key string, value int)](#Config.SetInt)
  * [func (c *Config) SetString(key string, value string)](#Config.SetString)
  * [func (c *Config) SetTime(key string, value time.Time)](#Config.SetTime)
  * [func (c *Config) SetURL(key string, value *url.URL)](#Config.SetURL)
  * [func (c *Config) String(key string) (string, error)](#Config.String)
  * [func (c *Config) Time(key string) (time.Time, error)](#Config.Time)
  * [func (c *Config) URL(key string) (*url.URL, error)](#Config.URL)
* [type EnvProvider](#EnvProvider)
  * [func (ep EnvProvider) Provide() (map[string]string, error)](#EnvProvider.Provide)
* [type FileProvider](#FileProvider)
  * [func (fp FileProvider) Provide() (map[string]string, error)](#FileProvider.Provide)
* [type MapProvider](#MapProvider)
  * [func (mp MapProvider) Provide() (map[string]string, error)](#MapProvider.Provide)
* [type Provider](#Provider)

#### <a name="pkg-examples">Examples</a>
* [New](#example_New)

#### <a name="pkg-files">Package files</a>
[cfg.go](/src/github.com/ardanlabs/kit/cfg/cfg.go) [cfg_default.go](/src/github.com/ardanlabs/kit/cfg/cfg_default.go) [doc.go](/src/github.com/ardanlabs/kit/cfg/doc.go) [env_provider.go](/src/github.com/ardanlabs/kit/cfg/env_provider.go) [file_provider.go](/src/github.com/ardanlabs/kit/cfg/file_provider.go) [map_provider.go](/src/github.com/ardanlabs/kit/cfg/map_provider.go) 





## <a name="Bool">func</a> [Bool](/src/target/cfg_default.go?s=2730:2765#L86)
``` go
func Bool(key string) (bool, error)
```
Bool calls the default Config and returns the bool value of a given key as a
bool. It will return an error if the key was not found or the value can't be
converted to a bool.



## <a name="Duration">func</a> [Duration](/src/target/cfg_default.go?s=3988:4036#L124)
``` go
func Duration(key string) (time.Duration, error)
```
Duration calls the default Config and returns the value of the given key as a
duration. It will return an error if the key was not found or the value can't be
converted to a Duration.



## <a name="Init">func</a> [Init](/src/target/cfg_default.go?s=339:366#L5)
``` go
func Init(p Provider) error
```
Init populates the package's default Config and should be called only once.
A Provider must be supplied which will return a map of key/value pairs to be
loaded.



## <a name="Int">func</a> [Int](/src/target/cfg_default.go?s=1479:1512#L48)
``` go
func Int(key string) (int, error)
```
Int calls the Default config and returns the value of the given key as an
int. It will return an error if the key was not found or the value
can't be converted to an int.



## <a name="Log">func</a> [Log](/src/target/cfg_default.go?s=696:713#L23)
``` go
func Log() string
```
Log returns a string to help with logging the package's default Config. It
excludes any values whose key contains the string "PASS".



## <a name="MustBool">func</a> [MustBool](/src/target/cfg_default.go?s=2969:2999#L93)
``` go
func MustBool(key string) bool
```
MustBool calls the default Config and returns the bool value of a given key
as a bool. It will panic if the key was not found or the value can't be
converted to a bool.



## <a name="MustDuration">func</a> [MustDuration](/src/target/cfg_default.go?s=4261:4304#L131)
``` go
func MustDuration(key string) time.Duration
```
MustDuration calls the default Config and returns the value of the given
key as a MustDuration. It will panic if the key was not found or the value
can't be converted to a MustDuration.



## <a name="MustInt">func</a> [MustInt](/src/target/cfg_default.go?s=1711:1739#L55)
``` go
func MustInt(key string) int
```
MustInt calls the default Config and returns the value of the given key as
an int. It will panic if the key was not found or the value can't be
converted to an int.



## <a name="MustString">func</a> [MustString](/src/target/cfg_default.go?s=1077:1111#L35)
``` go
func MustString(key string) string
```
MustString calls the default Config and returns the value of the given key
as a string, else it will panic if the key was not found.



## <a name="MustTime">func</a> [MustTime](/src/target/cfg_default.go?s=2331:2366#L74)
``` go
func MustTime(key string) time.Time
```
MustTime calls the default Config ang returns the value of the given key as
a Time. It will panic if the key was not found or the value can't be
converted to a Time.



## <a name="MustURL">func</a> [MustURL](/src/target/cfg_default.go?s=3587:3620#L112)
``` go
func MustURL(key string) *url.URL
```
MustURL calls the default Config and returns the value of the given key as a
URL. It will panic if the key was not found or the value can't be converted
to a URL.



## <a name="SetBool">func</a> [SetBool](/src/target/cfg_default.go?s=3109:3145#L98)
``` go
func SetBool(key string, value bool)
```
SetBool adds or modifies the default Config for the specified key and value.



## <a name="SetDuration">func</a> [SetDuration](/src/target/cfg_default.go?s=4422:4471#L136)
``` go
func SetDuration(key string, value time.Duration)
```
SetDuration adds or modifies the default Config for the specified key and value.



## <a name="SetInt">func</a> [SetInt](/src/target/cfg_default.go?s=1847:1881#L60)
``` go
func SetInt(key string, value int)
```
SetInt adds or modifies the default Config for the specified key and value.



## <a name="SetString">func</a> [SetString](/src/target/cfg_default.go?s=1228:1268#L41)
``` go
func SetString(key string, value string)
```
SetString adds or modifies the default Config for the specified key and
value.



## <a name="SetTime">func</a> [SetTime](/src/target/cfg_default.go?s=2476:2517#L79)
``` go
func SetTime(key string, value time.Time)
```
SetTime adds or modifies the default Config for the specified key and value.



## <a name="SetURL">func</a> [SetURL](/src/target/cfg_default.go?s=3728:3767#L117)
``` go
func SetURL(key string, value *url.URL)
```
SetURL adds or modifies the default Config for the specified key and value.



## <a name="String">func</a> [String](/src/target/cfg_default.go?s=871:910#L29)
``` go
func String(key string) (string, error)
```
String calls the default Config and returns the value of the given key as a
string. It will return an error if key was not found.



## <a name="Time">func</a> [Time](/src/target/cfg_default.go?s=2090:2130#L67)
``` go
func Time(key string) (time.Time, error)
```
Time calls the default Config and returns the value of the given key as a
Time. It will return an error if the key was not found or the value can't be
converted to a Time.



## <a name="URL">func</a> [URL](/src/target/cfg_default.go?s=3352:3390#L105)
``` go
func URL(key string) (*url.URL, error)
```
URL calls the default Config and returns the value of the given key as a
URL. It will return an error if the key was not found or the value can't be
converted to a URL.




## <a name="Config">type</a> [Config](/src/target/cfg.go?s=193:254#L5)
``` go
type Config struct {
    // contains filtered or unexported fields
}
```
Config is a goroutine safe configuration store, with a map of values
set from a config Provider.







### <a name="New">func</a> [New](/src/target/cfg.go?s=606:643#L18)
``` go
func New(p Provider) (*Config, error)
```
New populates a new Config from a Provider. It will return an error if there
was any problem reading from the Provider.





### <a name="Config.Bool">func</a> (\*Config) [Bool](/src/target/cfg.go?s=4202:4249#L179)
``` go
func (c *Config) Bool(key string) (bool, error)
```
Bool returns the bool value of a given key as a bool. It will return an
error if the key was not found or the value can't be converted to a bool.




### <a name="Config.Duration">func</a> (\*Config) [Duration](/src/target/cfg.go?s=6557:6617#L290)
``` go
func (c *Config) Duration(key string) (time.Duration, error)
```
Duration returns the value of the given key as a Duration. It will return an
error if the key was not found or the value can't be converted to a Duration.




### <a name="Config.Int">func</a> (\*Config) [Int](/src/target/cfg.go?s=2049:2094#L85)
``` go
func (c *Config) Int(key string) (int, error)
```
Int returns the value of the given key as an int. It will return an error if
the key was not found or the value can't be converted to an int.




### <a name="Config.Log">func</a> (\*Config) [Log](/src/target/cfg.go?s=876:905#L31)
``` go
func (c *Config) Log() string
```
Log returns a string to help with logging your configuration. It excludes
any values whose key contains the string "PASS".




### <a name="Config.MustBool">func</a> (\*Config) [MustBool](/src/target/cfg.go?s=4749:4791#L204)
``` go
func (c *Config) MustBool(key string) bool
```
MustBool returns the bool value of a given key as a bool. It will panic if
the key was not found or the value can't be converted to a bool.




### <a name="Config.MustDuration">func</a> (\*Config) [MustDuration](/src/target/cfg.go?s=7010:7065#L309)
``` go
func (c *Config) MustDuration(key string) time.Duration
```
MustDuration returns the value of the given key as a Duration. It will panic
if the key was not found or the value can't be converted into a Duration.




### <a name="Config.MustInt">func</a> (\*Config) [MustInt](/src/target/cfg.go?s=2453:2493#L104)
``` go
func (c *Config) MustInt(key string) int
```
MustInt returns the value of the given key as an int. It will panic if the
key was not found or the value can't be converted to an int.




### <a name="Config.MustString">func</a> (\*Config) [MustString](/src/target/cfg.go?s=1514:1560#L61)
``` go
func (c *Config) MustString(key string) string
```
MustString returns the value of the given key as a string. It will panic if
the key was not found.




### <a name="Config.MustTime">func</a> (\*Config) [MustTime](/src/target/cfg.go?s=3527:3574#L151)
``` go
func (c *Config) MustTime(key string) time.Time
```
MustTime returns the value of the given key as a Time. It will panic if the
key was not found or the value can't be converted to a Time.




### <a name="Config.MustURL">func</a> (\*Config) [MustURL](/src/target/cfg.go?s=5910:5955#L262)
``` go
func (c *Config) MustURL(key string) *url.URL
```
MustURL returns the value of the given key as a URL. It will panic if the
key was not found or the value can't be converted to a URL.




### <a name="Config.SetBool">func</a> (\*Config) [SetBool](/src/target/cfg.go?s=5208:5256#L228)
``` go
func (c *Config) SetBool(key string, value bool)
```
SetBool adds or modifies the configuration for the specified key and value.




### <a name="Config.SetDuration">func</a> (\*Config) [SetDuration](/src/target/cfg.go?s=7415:7476#L328)
``` go
func (c *Config) SetDuration(key string, value time.Duration)
```
SetDuration adds or modifies the configuration for a given duration at a
specific key.




### <a name="Config.SetInt">func</a> (\*Config) [SetInt](/src/target/cfg.go?s=2823:2869#L122)
``` go
func (c *Config) SetInt(key string, value int)
```
SetInt adds or modifies the configuration for the specified key and value.




### <a name="Config.SetString">func</a> (\*Config) [SetString](/src/target/cfg.go?s=1790:1842#L75)
``` go
func (c *Config) SetString(key string, value string)
```
SetString adds or modifies the configuration for the specified key and
value.




### <a name="Config.SetTime">func</a> (\*Config) [SetTime](/src/target/cfg.go?s=3916:3969#L169)
``` go
func (c *Config) SetTime(key string, value time.Time)
```
SetTime adds or modifies the configuration for the specified key and value.




### <a name="Config.SetURL">func</a> (\*Config) [SetURL](/src/target/cfg.go?s=6277:6328#L280)
``` go
func (c *Config) SetURL(key string, value *url.URL)
```
SetURL adds or modifies the configuration for the specified key and value.




### <a name="Config.String">func</a> (\*Config) [String](/src/target/cfg.go?s=1206:1257#L47)
``` go
func (c *Config) String(key string) (string, error)
```
String returns the value of the given key as a string. It will return an
error if key was not found.




### <a name="Config.Time">func</a> (\*Config) [Time](/src/target/cfg.go?s=3091:3143#L132)
``` go
func (c *Config) Time(key string) (time.Time, error)
```
Time returns the value of the given key as a Time. It will return an error
if the key was not found or the value can't be converted to a Time.




### <a name="Config.URL">func</a> (\*Config) [URL](/src/target/cfg.go?s=5506:5556#L243)
``` go
func (c *Config) URL(key string) (*url.URL, error)
```
URL returns the value of the given key as a URL. It will return an error if
the key was not found or the value can't be converted to a URL.




## <a name="EnvProvider">type</a> [EnvProvider](/src/target/env_provider.go?s=155:200#L2)
``` go
type EnvProvider struct {
    Namespace string
}
```
EnvProvider provides configuration from the environment. All keys will be
made uppercase.










### <a name="EnvProvider.Provide">func</a> (EnvProvider) [Provide](/src/target/env_provider.go?s=248:306#L7)
``` go
func (ep EnvProvider) Provide() (map[string]string, error)
```
Provide implements the Provider interface.




## <a name="FileProvider">type</a> [FileProvider](/src/target/file_provider.go?s=150:195#L1)
``` go
type FileProvider struct {
    Filename string
}
```
FileProvider describes a file based loader which loads the configuration
from a file listed.










### <a name="FileProvider.Provide">func</a> (FileProvider) [Provide](/src/target/file_provider.go?s=243:302#L6)
``` go
func (fp FileProvider) Provide() (map[string]string, error)
```
Provide implements the Provider interface.




## <a name="MapProvider">type</a> [MapProvider](/src/target/map_provider.go?s=118:168#L1)
``` go
type MapProvider struct {
    Map map[string]string
}
```
MapProvider provides a simple implementation of the Provider whereby it just
returns a stored map.










### <a name="MapProvider.Provide">func</a> (MapProvider) [Provide](/src/target/map_provider.go?s=216:274#L1)
``` go
func (mp MapProvider) Provide() (map[string]string, error)
```
Provide implements the Provider interface.




## <a name="Provider">type</a> [Provider](/src/target/cfg.go?s=413:478#L12)
``` go
type Provider interface {
    Provide() (map[string]string, error)
}
```
Provider is implemented by the user to provide the configuration as a map.
There are currently two Providers implemented, EnvProvider and MapProvider.














- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
