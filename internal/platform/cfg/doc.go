// Package cfg provides configuration options that are loaded from the environment.
// Configuration is then stored in memory and can be retrieved by its proper
// type.
//
// To initialize the configuration system from your environment, call Init:
//
//		cfg.Init(cfg.EnvProvider{Namespace: "configKey"})
//
// To retrieve values from configuration:
//
//  	proc, err := cfg.String("proc_id")
//  	port, err := cfg.Int("port")
//  	ms, err := cfg.Time("stamp")
//
// Use the Must set of function to retrieve a single value but these calls
// will panic if the key does not exist:
//
//  	proc := cfg.MustString("proc_id")
//  	port := cfg.MustInt("port")
//  	ms := cfg.MustTime("stamp")
package cfg
