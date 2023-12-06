package ardan.rego

default auth = false

auth := valid {
	[valid, _, _] := verify_jwt
}

verify_jwt := io.jwt.decode_verify(input.Token, {
        "cert": input.Key,
        "iss": input.ISS,
	}
)