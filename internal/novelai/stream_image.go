package novelai

import (
	"encoding/binary"
	"io"
)

type ImageStreamEvent struct {
	EventType string `msgpack:"event_type"`
	Image     []byte `msgpack:"image"`
	SampIX    int    `msgpack:"samp_ix"`
	StepIX    int    `msgpack:"step_ix"`
	Message   string `msgpack:"message"`
	Code      string `msgpack:"code"`
}

func ParseImageStream(r io.Reader) ([]ImageStreamEvent, error) {
	var out []ImageStreamEvent
	for {
		var size uint32
		if err := binary.Read(r, binary.BigEndian, &size); err != nil {
			if err == io.EOF {
				return out, nil
			}
			return nil, err
		}
		payload := make([]byte, size)
		if _, err := io.ReadFull(r, payload); err != nil {
			return nil, err
		}
		var event ImageStreamEvent
		if err := DecodeMsgpack(payload, &event); err != nil {
			return nil, err
		}
		out = append(out, event)
	}
}
