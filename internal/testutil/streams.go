package testutil

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

func BuildImageStreamFixture(t *testing.T) io.Reader {
	t.Helper()
	event := map[string]any{
		"event_type": "final",
		"image":      []byte("png"),
		"samp_ix":    0,
		"step_ix":    10,
	}
	payload, err := msgpack.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(payload))); err != nil {
		t.Fatal(err)
	}
	if _, err := buf.Write(payload); err != nil {
		t.Fatal(err)
	}
	return &buf
}
