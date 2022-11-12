package auth

import (
	_ "embed"
)

// These are the expected values for Claims.Roles.
const (
	RoleAdmin = "ADMIN"
	RoleUser  = "USER"
)

// Package name of our rego code.
const (
	opaPackage string = "ardan.rego"
)

// Core OPA policies.
var (
	//go:embed rego/authentication.rego
	opaAuthentication string

	//go:embed rego/authorization.rego
	opaAuthorization string
)
