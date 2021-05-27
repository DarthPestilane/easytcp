package packet

type Codec interface {
	Encode()
	Decode(data []byte) (interface{}, error)
}

type DefaultCodec struct {
}

func (d *DefaultCodec) Encode() {
	panic("implement me")
}

func (d *DefaultCodec) Decode(data []byte) (interface{}, error) {
	return string(data), nil
}
