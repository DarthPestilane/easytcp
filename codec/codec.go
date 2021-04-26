package codec

// Codec 解码器
type Codec interface {
	// Marshal encode msg data into byte
	Marshal(msg interface{}) ([]byte, error)

	// Unmarshal decode byte and bind to data
	Unmarshal(b []byte, data interface{}) error
}
