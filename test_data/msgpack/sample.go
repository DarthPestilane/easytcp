package msgpack

// Sample is a sample struct for test only.
type Sample struct {
	Foo string         `msgpack:"foo"`
	Bar int64          `msgpack:"bar"`
	Baz map[int]string `msgpack:"baz"`
}
