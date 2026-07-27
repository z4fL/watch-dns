package nextdns

import (
	"encoding/json"
	"fmt"
)

type APIError struct {
	StatusCode int
	Errors     []ErrorItem
}

type ErrorItem struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
	Source struct {
		Parameter string `json:"parameter,omitempty"`
		Pointer   string `json:"pointer,omitempty"`
	} `json:"source"`
}

func (e *APIError) Error() string {
	if len(e.Errors) == 0 {
		return fmt.Sprintf("nextdns api error: status %d", e.StatusCode)
	}
	return fmt.Sprintf("nextdns api error: status %d, detail=%s", e.StatusCode, e.Errors[0].Detail)
}

type errorEnvelope struct {
	Errors []ErrorItem `json:"errors"`
}

func parseAPIError(raw []byte, statusCode int) error {
	var env errorEnvelope
	if err := json.Unmarshal(raw, &env); err == nil && len(env.Errors) > 0 {
		return &APIError{StatusCode: statusCode, Errors: env.Errors}
	}
	return fmt.Errorf("nextdns api error: status %d, body=%s", statusCode, string(raw))
}
