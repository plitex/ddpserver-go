package ddpserver

// Message used to decode all messages, it contains all posible fields
type Message struct {
	Msg     string        `json:"msg,omitempty"`
	Session string        `json:"session,omitempty"`
	Version string        `json:"version,omitempty"`
	Support []string      `json:"support,omitempty"`
	ID      string        `json:"id,omitempty"`
	Method  string        `json:"method,omitempty"`
	Params  []interface{} `json:"params,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
	Methods []string      `json:"methods,omitempty"`
	Name    string        `json:"name,omitempty"`
	Subs    []string      `json:"subs,omitempty"`
	Error   Error         `json:"error,omitempty"`
}
