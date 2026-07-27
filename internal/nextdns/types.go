package nextdns

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
