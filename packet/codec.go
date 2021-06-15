package packet

import (
	"encoding/json"
	"fmt"
)

//go:generate mockgen -destination mock/codec_mock.go -package mock . Codec

// Codec is a generic codec for encoding and decoding data.
type Codec interface {
	// Encode encodes data into []byte.
	// Returns error when error occurred.
	Encode(data interface{}) ([]byte, error)

	// Decode decodes data into v.
	// Returns error when error occurred.
	Decode(data []byte, v interface{}) error
}

var _ Codec = &StringCodec{}

// StringCodec implements the Codec interface.
// StringCodec encodes string into []byte, and decodes data into string.
type StringCodec struct{}

// Encode implements the Codec Encode method.
func (c *StringCodec) Encode(data interface{}) ([]byte, error) {
	return []byte(data.(string)), nil
}

// Decode implements the Codec Decode method.
// Parameter v should be a String pointer, or an error will return.
func (c *StringCodec) Decode(data []byte, v interface{}) error {
	if _, ok := v.(*string); !ok {
		return fmt.Errorf("v must be a string pointer")
	}
	*v.(*string) = string(data)
	return nil
}

// JsonCodec implements the Codec interface.
// JsonCodec encodes and decodes data in json way.
type JsonCodec struct{}

// Encode implements the Codec Encode method.
func (c *JsonCodec) Encode(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// Decode implements the Codec Decode method.
func (c *JsonCodec) Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
