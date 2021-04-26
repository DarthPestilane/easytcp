package codec

import (
	"fmt"
	"reflect"
)

var _ Codec = &SimpleText{}
var DefaultSimpleText = &SimpleText{}

type SimpleText struct{}

func (s SimpleText) Marshal(msg interface{}) ([]byte, error) {
	str, ok := msg.(string)
	if !ok {
		return nil, fmt.Errorf("msg should be a Stringer")
	}
	return []byte(str), nil
}

func (s SimpleText) Unmarshal(b []byte, data interface{}) error {
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("invalid data type: %s", reflect.TypeOf(data))
	}
	rv.Elem().Set(reflect.ValueOf(string(b)))
	return nil
}
