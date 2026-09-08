package handlers

import "testing"

func TestInspectInputString(t *testing.T) {
	isMalicious, pType := InspectInputString("SELECT * FROM users WHERE 1=1")
	if !isMalicious || pType != "SQLi" {
		t.Errorf("expected SQLi detection")
	}

	isMalicious, pType = InspectInputString("<script>alert('xss')</script>")
	if !isMalicious || pType != "XSS" {
		t.Errorf("expected XSS detection")
	}

	isMalicious, _ = InspectInputString("normal_param_value")
	if isMalicious {
		t.Errorf("expected clean string not to be flagged")
	}
}
