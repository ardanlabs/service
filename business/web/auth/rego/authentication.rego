package ardan.rego

default auth = false

# This function decodes and verifies the JWT, it also makes sure that it hasn't expired etc. 
auth {
	jwt_valid
}

jwt_valid := valid {
	[valid, header, payload] := decoded_jwt
}

decoded_jwt := [valid, header, payload] {
	[valid, header, payload] := io.jwt.decode_verify(input.Token, {
		"cert": input.Key,
		"alg": "RS256",
        "iss": "service project",
	})
}
