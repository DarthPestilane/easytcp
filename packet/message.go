package packet

//go:generate mockgen -destination mock/message_mock.go -package mock . Message

// Message is an interface for a message object after unpacked.
type Message interface {
	// Getters

	// GetSize returns the size,
	// which is the size of message data, or of the whole message.
	GetSize() uint

	// GetID returns the message ID.
	GetID() uint

	// GetData returns the data of message.
	GetData() []byte

	// Setters

	Setup(id uint, data []byte)

	Duplicate() Message
}

var _ Message = &DefaultMsg{}

// DefaultMsg implements the Message interface.
// DefaultMsg is of the format as:
// 	(Size)(ID)(Data)
// 	(4 bytes)(4 bytes)(n bytes)
// DefaultMsg will be returned in DefaultPacker.Unpack() method.
type DefaultMsg struct {
	ID   uint32
	Size uint32
	Data []byte
}

func (d *DefaultMsg) Duplicate() Message {
	return &DefaultMsg{}
}

func (d *DefaultMsg) Setup(id uint, data []byte) {
	d.ID = uint32(id)
	d.Data = data
	d.Size = uint32(len(data))
}

// GetID implements the Message GetID method.
func (d *DefaultMsg) GetID() uint {
	return uint(d.ID)
}

// GetSize implements the Message GetSize method.
func (d *DefaultMsg) GetSize() uint {
	return uint(d.Size)
}

// GetData implements the Message GetData method.
func (d *DefaultMsg) GetData() []byte {
	return d.Data
}
