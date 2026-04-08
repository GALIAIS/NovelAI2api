package novelai

import "github.com/vmihailenco/msgpack/v5"

func DecodeMsgpack(payload []byte, out any) error {
	return msgpack.Unmarshal(payload, out)
}
