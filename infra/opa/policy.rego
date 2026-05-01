package ironclad.authz

import rego.v1

default allow = false

# Allow if risk score is low (< 0.6)
allow if {
	input.risk_score < 0.6
}

# Allow if user is an admin regardless of risk
allow if {
	input.user.role == "admin"
}

# Allow if intent is a "hotfix" and risk is moderate (< 0.8)
allow if {
	input.intent == "hotfix"
	input.risk_score < 0.8
}
