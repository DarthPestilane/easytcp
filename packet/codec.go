package packet

import (
	"fmt"
	"reflect"
)

// Codec 编码解码器
// 对原始消息的 data 进行编码和解码处理
type Codec interface {
	// Encode 编码
	// data 为需要编码的数据, 通常是 Response 的 Data
	// 编码后的结果，通常应当经过 Packer.Pack() 打包成待发送的消息
	Encode(data interface{}) ([]byte, error) // 编码

	// Decode 解码
	// data 为需要解码的数据, 通常是 Message.GetData() 返回的数据
	Decode(data []byte, v interface{}) error // 解码
}

var _ Codec = &DefaultCodec{}

type DefaultCodec struct {
}

func (d *DefaultCodec) Encode(data interface{}) ([]byte, error) {
	return []byte(data.(string)), nil
}

func (d *DefaultCodec) Decode(data []byte, v interface{}) error {
	vv := reflect.ValueOf(v)
	if vv.Kind() != reflect.Ptr || vv.Elem().Kind() != reflect.String {
		return fmt.Errorf("v must be a string pointer")
	}
	vv.Elem().SetString(string(data))
	return nil
}
