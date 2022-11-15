package ardan.rego

default allowAny = false
default allowOnlyUser = false
default allowOnlyAdmin = false

roleUser := "USER"
roleAdmin := "ADMIN"
roleAll := {roleAdmin, roleUser}

allowAny {
	roles_from_claims := {role | role := input.Roles[_]}
	input_role_is_in_claim := roleAll & roles_from_claims
	count(input_role_is_in_claim) > 0
}

allowOnlyUser {
	roles_from_claims := {role | role := input.Roles[_]}
	input_role_is_in_claim := {roleUser} & roles_from_claims
	count(input_role_is_in_claim) > 0
}

allowOnlyAdmin {
	roles_from_claims := {role | role := input.Roles[_]}
	input_role_is_in_claim := {roleAdmin} & roles_from_claims
	count(input_role_is_in_claim) > 0
}
