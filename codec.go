package easytcp

import (
	"encoding/json"
)

//go:generate mockgen -destination mock/codec_mock.go -package mock . Codec

// Codec is a generic codec for encoding and decoding data.
type Codec interface {
	// Encode encodes data into []byte.
	// Returns error when error occurred.
	Encode(v interface{}) ([]byte, error)

	// Decode decodes data into v.
	// Returns error when error occurred.
	Decode(data []byte, v interface{}) error
}

var _ Codec = &JsonCodec{}

// JsonCodec implements the Codec interface.
// JsonCodec encodes and decodes data in json way.
type JsonCodec struct{}

// Encode implements the Codec Encode method.
func (c *JsonCodec) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Decode implements the Codec Decode method.
func (c *JsonCodec) Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
