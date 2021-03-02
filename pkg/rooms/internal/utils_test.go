package internal_test

import (
	"testing"

	"github.com/soapboxsocial/soapbox/pkg/rooms/internal"
)

func TestTrimRoomNameToLimit(t *testing.T) {
	var tests = []struct {
		in  string
		out string
	}{
		{
			"Test ",
			"Test",
		},
		{
			" Test ",
			"Test",
		},
		{
			"1037619662938620156055447100655",
			"103761966293862015605544710065",
		},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {

			result := internal.TrimRoomNameToLimit(tt.in)
			if tt.out != result {
				t.Fatalf("expected: %s did not match actual: %s", tt.out, result)
			}

		})
	}
}
