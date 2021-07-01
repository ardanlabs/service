package config

import "go.uber.org/zap"

func RunApi(log *zap.SugaredLogger, build string) error {
	// =========================================================================
	// Configuration
	cfg, err := Configuration(log, build)
	if err != nil {
		return err
	}

	// =========================================================================
	// App Starting
	initExpvar(log, cfg)

	// =========================================================================
	// Initialize authentication support
	auth, err := initAuth(log, cfg)
	if err != nil {
		return err
	}

	// =========================================================================
	// Start Database
	db, err := initDatabase(log, cfg)
	if err != nil {
		return err
	}

	// =========================================================================
	// Start Tracing Support
	err = initTracing(log, cfg)
	if err != nil {
		return err
	}
	// =========================================================================
	// Start Debug Service
	initDebugService(log, cfg, db)

	// =========================================================================
	// Start API Service
	api, shutdown, serverErrors := initAPIService(log, cfg, auth, db)

	// =========================================================================
	// Shutdown
	err = shut(log, cfg, api, shutdown, serverErrors)
	if err != nil {
		return err
	}

	return nil
}
