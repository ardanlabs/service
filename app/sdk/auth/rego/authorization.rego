package ardan.rego

import rego.v1

role_user := "USER"

role_admin := "ADMIN"

role_all := {role_admin, role_user}

default rule_any := false

rule_any if {
	claim_roles := {role | some role in input.Roles}
	input_roles := role_all & claim_roles
	count(input_roles) > 0
}

default rule_admin_only := false

rule_admin_only if {
	claim_roles := {role | some role in input.Roles}
	input_admin := {role_admin} & claim_roles
	count(input_admin) > 0
}

default rule_user_only := false

rule_user_only if {
	claim_roles := {role | some role in input.Roles}
	input_user := {role_user} & claim_roles
	count(input_user) > 0
}

default rule_admin_or_subject := false

rule_admin_or_subject if {
	claim_roles := {role | some role in input.Roles}
	input_admin := {role_admin} & claim_roles
	count(input_admin) > 0
} else if {
	claim_roles := {role | some role in input.Roles}
	input_user := {role_user} & claim_roles
	count(input_user) > 0
	input.UserID == input.Subject
}
