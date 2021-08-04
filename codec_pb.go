package easytcp

import (
	"fmt"
	"google.golang.org/protobuf/proto"
)

// ProtobufCodec implements the Codec interface.
type ProtobufCodec struct{}

// Encode implements the Codec Encode method.
func (p *ProtobufCodec) Encode(v interface{}) ([]byte, error) {
	m, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("v should be proto.Message but %T", v)
	}
	return proto.Marshal(m)
}

// Decode implements the Codec Decode method.
func (p *ProtobufCodec) Decode(data []byte, v interface{}) error {
	m, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("v should be proto.Message but %T", v)
	}
	return proto.Unmarshal(data, m)
}
