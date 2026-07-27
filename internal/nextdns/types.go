package nextdns

import "time"

type Response[T any] struct {
	Data T             `json:"data"`
	Meta *ResponseMeta `json:"meta,omitempty"`
}

type ResponseMeta struct {
	Pagination *Pagination `json:"pagination,omitempty"`
	Stream     *StreamMeta `json:"stream,omitempty"`
}

type Pagination struct {
	Cursor *string `json:"cursor,omitempty"`
}

type StreamMeta struct {
	ID string `json:"id"`
}

type LogEvent struct {
	ID   string
	Data Log
}

type Log struct {
	Timestamp time.Time `json:"timestamp"`
	Domain    string    `json:"domain"`
	Root      string    `json:"root,omitempty"`
	Tracker   string    `json:"tracker,omitempty"`
	Encrypted bool      `json:"encrypted"`
	Protocol  string    `json:"protocol"`
	ClientIP  string    `json:"clientIp,omitempty"`
	Client    string    `json:"client,omitempty"`
	Device    *Device   `json:"device,omitempty"`
	Status    string    `json:"status"`
	Reasons   []Reason  `json:"reasons"`
}

type Device struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Model   string `json:"model,omitempty"`
	LocalIP string `json:"localIp,omitempty"`
}

type Reason struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
