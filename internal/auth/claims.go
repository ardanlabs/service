package auth

import jwt "github.com/dgrijalva/jwt-go"

const (
	RoleAdmin = "ADMIN"
	RoleTODO  = "TODO:MORE_ROLES"
)

type Claims struct {
	Roles []string `json:"roles"`
	jwt.StandardClaims
}
