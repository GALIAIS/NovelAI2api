package novelai

import (
	"testing"

	"novelai/internal/testutil"
)

func TestParseImageStreamReadsFinalFrame(t *testing.T) {
	stream := testutil.BuildImageStreamFixture(t)
	events, err := ParseImageStream(stream)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("expected events")
	}
}
