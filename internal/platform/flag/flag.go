package main

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	_ "github.com/spf13/pflag"
)

var cfg struct {
	Web struct {
		APIHost         string        `default:"0.0.0.0:3000" envconfig:"API_HOST" flag:"a" flagdesc:"api host ip:port"`
		DebugHost       string        `default:"0.0.0.0:4000" envconfig:"DEBUG_HOST" flagdesc:"debug host ip:port"`
		ReadTimeout     time.Duration `default:"5s" envconfig:"READ_TIMEOUT"`
		WriteTimeout    time.Duration `default:"5s" envconfig:"WRITE_TIMEOUT"`
		ShutdownTimeout time.Duration `default:"5s" envconfig:"SHUTDOWN_TIMEOUT"`
	}
	DB struct {
		DialTimeout time.Duration `default:"5s" envconfig:"DIAL_TIMEOUT"`
		Host        string        `default:"mongo:27017/gotraining" envconfig:"HOST"`
	}
	Trace struct {
		Host         string        `default:"http://tracer:3002/v1/publish" envconfig:"HOST"`
		BatchSize    int           `default:"1000" envconfig:"batch_size" envconfig:"BATCH_SIZE"`
		SendInterval time.Duration `default:"15s" envconfig:"send_interval" envconfig:"SEND_INTERVAL"`
		SendTimeout  time.Duration `default:"500ms" envconfig:"send_timeout" envconfig:"SEND_TIMEOUT"`
	}
}

func main() {
	if err := Apply(nil, "", &cfg); err != nil {
		fmt.Println(err)
	}
}

// Apply will take a collection of command line arguments and
// a user defined struct and apply changes to the struct's value.
func Apply(args []string, fieldName string, v interface{}) error {
	rawValue := reflect.ValueOf(v)
	if fieldName != "" {
		fieldName = strings.ToLower(fieldName) + "_"
	}

	var val reflect.Value
	switch rawValue.Kind() {
	case reflect.Ptr:
		val = rawValue.Elem()
		if val.Kind() != reflect.Struct {
			return fmt.Errorf("incompatible type `%v` looking for a pointer", val.Kind())
		}
	case reflect.Struct:
		var ok bool
		if val, ok = v.(reflect.Value); !ok {
			return fmt.Errorf("internal recurse error")
		}
	default:
		return fmt.Errorf("incompatible type `%v`", rawValue.Kind())
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		if field.Type.Kind() == reflect.Struct {
			if err := Apply(args, fieldName+field.Name, val.Field(i)); err != nil {
				return err
			}
			continue
		}

		tag := field.Tag
		flagShort := tag.Get("flag")
		flagLong := fieldName + strings.ToLower(field.Name)
		flagdef := tag.Get("default")
		flagDesc := tag.Get("flagdesc")

		fmt.Println("flagShort:", flagShort, " flagLong:", flagLong, " default:", flagdef, " desc:", flagDesc)
	}

	return nil
}
