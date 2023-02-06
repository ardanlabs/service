package ardan.rego

default auth = false

auth {
	jwt_valid
}

jwt_valid := valid {
	[valid, header, payload] := verify_jwt
}

verify_jwt := io.jwt.decode_verify(input.Token, {
        "cert": input.Key,
        "iss": input.ISS,
	}
)
