package easytcp

import (
	"github.com/vmihailenco/msgpack/v5"
)

// MsgpackCodec implements the Codec interface.
type MsgpackCodec struct{}

// Encode implements the Codec Encode method.
func (m *MsgpackCodec) Encode(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

// Decode implements the Codec Decode method.
func (m *MsgpackCodec) Decode(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}
