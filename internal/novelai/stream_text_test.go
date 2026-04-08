package novelai

import (
	"strings"
	"testing"
)

func TestParseCompletionStreamReadsDone(t *testing.T) {
	src := strings.NewReader("data: {\"choices\":[{\"text\":\"hi\"}]}\n\ndata: [DONE]\n\n")
	chunks, err := ParseCompletionStream(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 1 {
		t.Fatalf("len = %d", len(chunks))
	}
}
