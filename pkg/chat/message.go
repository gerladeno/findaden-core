package chat

type Message struct {
	Sender    string `json:"sender"`
	Receiver  string `json:"receiver"`
	Timestamp string `json:"timestamp"`
	Body      string `json:"body"`
}

func (m *Message) String() string {
	return m.Sender + " at " + m.Timestamp + " says " + m.Body
}
