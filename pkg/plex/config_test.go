package plex

import (
	"testing"
)

func TestTriggerIsMatch(t *testing.T) {
	// Create Trigger
	tr := Trigger{
		Properties: map[string]interface{}{
			"PropertyA": "1234",
			"Deep.Property": "5678",
			"SomeNumeric": float64(1),
		},
	}
	// Create JSON that should match it
	payload := []byte(`{
		"PropertyA": "1234",
		"SomeUselessThing": "blah blah blah",
		"Deep": {
			"Property": "5678"
		},
		"SomeNumeric": 1
	}`)
	// Does it spark joy?
	if !tr.IsMatch(payload) {
		t.Errorf("Expected trigger to match payload, but it didn't!")
	}

	// Create JSON that should not match it
	payload = []byte(`{
		"PropertyA": "4321",
		"SomeUselessThing": "blah blah blah",
		"Deep": {
			"Property": "5678"
		},
		"SomeNumeric": 1
	}`)
	// Does it spark joy?
	if tr.IsMatch(payload) {
		t.Errorf("Expected trigger to NOT match payload, but it did!")
	}
}