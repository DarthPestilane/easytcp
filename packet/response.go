package packet

// Response is the response sent to session.Session.
// Response contains message's ID and Data.
// Usually the Data is not encoded yet at this moment, and will be encoded in session.Session's SendResp() method.
type Response struct {
	// ID is the message's ID
	ID uint

	// Data is the message's data, usually the Data is not encoded yet.
	Data interface{}
}
