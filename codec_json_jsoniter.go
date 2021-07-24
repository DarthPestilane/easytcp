// +build jsoniter

package easytcp

import (
	jsoniter "github.com/json-iterator/go"
)

var _ Codec = &JsonCodec{}

// JsonCodec implements the Codec interface.
// JsonCodec encodes and decodes data in json way.
type JsonCodec struct{}

// Encode implements the Codec Encode method.
func (c *JsonCodec) Encode(v interface{}) ([]byte, error) {
	return jsoniter.Marshal(v)
}

// Decode implements the Codec Decode method.
func (c *JsonCodec) Decode(data []byte, v interface{}) error {
	return jsoniter.Unmarshal(data, v)
}
