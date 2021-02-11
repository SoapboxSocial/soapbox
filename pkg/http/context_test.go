package http_test

import (
	"context"
	"testing"

	"github.com/soapboxsocial/soapbox/pkg/http"
)

func TestWithUserID(t *testing.T) {
	ctx := context.Background()
	id := 12

	with := http.WithUserID(ctx, id)

	val, ok := http.GetUserIDFromContext(with)
	if !ok {
		t.Fatal("no user ID stored")
	}

	if val != id {
		t.Fatalf("%d does not match %d", val, id)
	}
}
