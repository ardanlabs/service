package ardan.rego

default ruleAny = false
default ruleAdminOnly = false
default ruleUserOnly = false
default ruleAdminOrSubject = false

roleUser := "USER"
roleAdmin := "ADMIN"
roleAll := {roleAdmin, roleUser}

ruleAny {
	claim_roles := {role | role := input.Roles[_]}
	input_roles := roleAll & claim_roles
	count(input_roles) > 0
}

ruleAdminOnly {
	claim_roles := {role | role := input.Roles[_]}
	input_admin := {roleAdmin} & claim_roles
	count(input_admin) > 0
}

ruleUserOnly {
	claim_roles := {role | role := input.Roles[_]}
	input_user := {roleUser} & claim_roles
	count(input_user) > 0
}

ruleAdminOrSubject {
	claim_roles := {role | role := input.Roles[_]}
	input_admin := {roleAdmin} & claim_roles
    count(input_admin) > 0
} else {
    claim_roles := {role | role := input.Roles[_]}
	input_user := {roleUser} & claim_roles
	count(input_user) > 0
	input.UserID == input.Subject
}
