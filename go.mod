module github.com/ardanlabs/service

go 1.13

require (
	contrib.go.opencensus.io/exporter/zipkin v0.1.1
	github.com/GuiaBolso/darwin v0.0.0-20170210191649-86919dfcf808
	github.com/ardanlabs/conf v1.1.0
	github.com/cznic/ql v1.2.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dimfeld/httptreemux/v5 v5.0.2
	github.com/go-playground/locales v0.12.1
	github.com/go-playground/universal-translator v0.16.0
	github.com/google/go-cmp v0.3.1
	github.com/google/uuid v1.1.1
	github.com/jmoiron/sqlx v1.2.0
	github.com/leodido/go-urn v1.1.0 // indirect
	github.com/lib/pq v1.2.0
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.4.0 // indirect
	go.opencensus.io v0.22.1
	golang.org/x/crypto v0.0.0-20190911031432-227b76d455e7
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.29.1
)

replace gopkg.in/DATA-DOG/go-sqlmock.v1 => gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0
