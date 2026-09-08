package ironclad.authz_test

import rego.v1
import data.ironclad.authz.allow

test_low_risk_allowed if {
	allow with input as {"risk_score": 0.3}
}

test_high_risk_blocked if {
	not allow with input as {"risk_score": 0.9, "environment": "production", "user": {"role": "developer"}}
}
