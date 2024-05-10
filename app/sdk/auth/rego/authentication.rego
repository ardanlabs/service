package ardan.rego

import rego.v1

default auth := false

auth if {
	[valid, _, _] := verify_jwt
	valid = true
}

verify_jwt := io.jwt.decode_verify(input.Token, {
	"cert": input.Key,
	"iss": input.ISS,
})
