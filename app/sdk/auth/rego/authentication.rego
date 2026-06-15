package ardan.rego

default auth := false

auth if {
	[valid, _, _] := io.jwt.decode_verify(input.Token, {
		"cert": input.Key,
		"iss": input.ISS,
	})
	valid == true
}
