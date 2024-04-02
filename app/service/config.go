package service

import (
	"time"

	"github.com/ardanlabs/conf/v3"
)

// WebConf provides the set http related configuration.
type WebConf struct {
	ReadTimeout        time.Duration `conf:"default:5s"`
	WriteTimeout       time.Duration `conf:"default:10s"`
	IdleTimeout        time.Duration `conf:"default:120s"`
	ShutdownTimeout    time.Duration `conf:"default:20s"`
	APIHost            string        `conf:"default:0.0.0.0:3000"`
	DebugHost          string        `conf:"default:0.0.0.0:4000"`
	CORSAllowedOrigins []string      `conf:"default:*"`
}

// AuthConf provides the set auth related configuration.
type AuthConf struct {
	KeysFolder string `conf:"default:zarf/keys/"`
	ActiveKID  string `conf:"default:54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"`
	Issuer     string `conf:"default:service project"`
}

// DBConf provides the set database related configuration.
type DBConf struct {
	User         string `conf:"default:postgres"`
	Password     string `conf:"default:postgres,mask"`
	HostPort     string `conf:"default:database-service.sales-system.svc.cluster.local"`
	Name         string `conf:"default:postgres"`
	MaxIdleConns int    `conf:"default:2"`
	MaxOpenConns int    `conf:"default:0"`
	DisableTLS   bool   `conf:"default:true"`
}

// TempoConf provides the set tempo related configuration.
type TempoConf struct {
	ReporterURI string  `conf:"default:tempo.sales-system.svc.cluster.local:4317"`
	ServiceName string  `conf:"default:sales-api"`
	Probability float64 `conf:"default:0.05"`
	// Shouldn't use a high Probability value in non-developer systems.
	// 0.05 should be enough for most systems. Some might want to have
	// this even lower.
}

// Config is the set of configuration needed to start the system.
type Config struct {
	conf.Version
	Web   WebConf
	Auth  AuthConf
	DB    DBConf
	Tempo TempoConf
}
