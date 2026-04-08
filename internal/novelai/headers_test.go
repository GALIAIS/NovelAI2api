package novelai

import (
	"strings"
	"testing"
)

func TestBuildHeadersAddsCorrelationValues(t *testing.T) {
	headers := BuildHeaders("token")
	if headers.Get("Authorization") != "Bearer token" {
		t.Fatalf("Authorization = %q", headers.Get("Authorization"))
	}
	correlationID := headers.Get("x-correlation-id")
	if correlationID == "" {
		t.Fatal("missing x-correlation-id")
	}
	if len(correlationID) != 6 {
		t.Fatalf("x-correlation-id length = %d", len(correlationID))
	}
	if strings.ContainsAny(correlationID, "-_ ") {
		t.Fatalf("x-correlation-id contains non-alphanumeric characters: %q", correlationID)
	}
}
