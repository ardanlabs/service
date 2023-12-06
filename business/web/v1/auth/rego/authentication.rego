package ardan.rego

import future.keywords.if
import future.keywords.in

default auth := false

auth if {
	[valid, _, _] := verify_jwt
	valid = true
}

verify_jwt := io.jwt.decode_verify(input.Token, {
	"cert": input.Key,
	"iss": input.ISS,
})
