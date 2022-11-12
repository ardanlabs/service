package ardan.rego

default allow = false

allow {
	# The roles provided in the argument.
	input_roles := {inputRole | inputRole := input.InputRoles[_]}

	# The roles from the user's claims
	roles_from_claims := {role | role := input.Roles[_]}

	# The intersection between the input roles, and roles in the claim
	input_role_is_in_claim := input_roles & roles_from_claims

	# The number of roles within the users' claims, if its more than 0 then the condition is true.
	count(input_role_is_in_claim) > 0
}
