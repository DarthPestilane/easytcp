package fixture

import (
	"github.com/DarthPestilane/easytcp"
	"google.golang.org/protobuf/proto"
)

var _ easytcp.Codec = &ProtoCodec{}

type ProtoCodec struct{}

func (p *ProtoCodec) Encode(data interface{}) ([]byte, error) {
	return proto.Marshal(data.(proto.Message))
}

func (p *ProtoCodec) Decode(data []byte, v interface{}) error {
	return proto.Unmarshal(data, v.(proto.Message))
}
