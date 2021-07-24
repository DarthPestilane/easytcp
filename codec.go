package easytcp

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
