package http_test

import (
	"net/url"
	"testing"

	"github.com/soapboxsocial/soapbox/pkg/http"
)

func TestGetInt(t *testing.T) {
	var tests = []struct {
		value        string
		expected     int
		defaultValue int
	}{
		{
			"poop",
			10,
			10,
		},
		{
			"1",
			1,
			10,
		},
		{
			"",
			10,
			10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {

			values := url.Values{}
			values.Set("key", tt.value)

			result := http.GetInt(values, "key", tt.defaultValue)
			if result != tt.expected {
				t.Fatalf("expected %d does not match actual %d", tt.expected, result)
			}
		})
	}
}
