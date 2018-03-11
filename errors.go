package ddpserver

type Error struct {
	Type    string `json:"errorType,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

func NewError(err string, reason string, message string) *Error {
	return &Error{
		Type:    "Server.Error",
		Error:   err,
		Reason:  reason,
		Message: message,
	}
}
