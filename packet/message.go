package packet

//go:generate mockgen -destination mock/message_mock.go -package mock . Message

// Message is an interface for a message object after unpacked.
type Message interface {
	// GetSize returns the size,
	// which is the size of message data, or of the whole message.
	GetSize() uint

	// GetID returns the message ID.
	GetID() uint

	// GetData returns the data of message.
	GetData() []byte
}

var _ Message = &DefaultMsg{}

// DefaultMsg implements the Message interface.
// DefaultMsg is returned in DefaultPacker.Unpack() method.
type DefaultMsg struct {
	ID   uint32
	Size uint32
	Data []byte
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
