package common

import (
	"strings"
	"testing"
)

func TestRedactPath(t *testing.T) {
	cases := map[string]string{
		"/ws/events/1/orders?token=eyJhbGciOi.secret.sig": "/ws/events/1/orders?token=REDACTED",
		"/api/x?status=NEW&token=abc":                     "/api/x?status=NEW&token=REDACTED",
		"/api/warehouse/events/1/orders?status=NEW":       "/api/warehouse/events/1/orders?status=NEW",
		"/api/health": "/api/health",
	}
	for in, want := range cases {
		got := RedactPath(in)
		if got != want {
			t.Errorf("RedactPath(%q) = %q, want %q", in, got, want)
		}
		if strings.Contains(got, "secret") {
			t.Errorf("token leaked: %q", got)
		}
	}
}
