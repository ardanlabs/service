package cfg

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// EnvProvider provides configuration from the environment. All keys will be
// made uppercase.
type EnvProvider struct {
	Namespace string
}

// Provide implements the Provider interface.
func (ep EnvProvider) Provide() (map[string]string, error) {

	// Store the config in this empty map.
	config := map[string]string{}

	// Get the lists of available environment variables.
	envs := os.Environ()
	if len(envs) == 0 {
		return nil, errors.New("No environment variables found")
	}

	// Create the uppercase version to meet the standard {NAMESPACE_} format.
	uspace := fmt.Sprintf("%s_", strings.ToUpper(ep.Namespace))

	// Loop and match each variable using the uppercase namespace.
	for _, val := range envs {
		if !strings.HasPrefix(val, uspace) {
			continue
		}

		idx := strings.Index(val, "=")
		config[strings.ToUpper(strings.TrimPrefix(val[0:idx], uspace))] = val[idx+1:]
	}

	// Did we find any keys for this namespace?
	if len(config) == 0 {
		return nil, fmt.Errorf("Namespace %q was not found", ep.Namespace)
	}

	return config, nil
}
