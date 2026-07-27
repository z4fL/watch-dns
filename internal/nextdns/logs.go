package nextdns

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Domain    string `json:"domain"`
	Root      string `json:"root,omitempty"`
	Status    string `json:"status"`
}

type LogListResponse = Response[[]LogEntry]

func (c *Client) GetLogs(ctx context.Context, profileID string, q LogsQuery) (*LogListResponse, error) {
	query := q.Values()
	var resp LogListResponse
	err := c.do(ctx, "GET", "/profiles/"+profileID+"/logs", query, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type LogsQuery struct {
	From   string
	To     string
	Sort   string
	Limit  int
	Cursor string
	Device string
	Status string
	Search string
	Raw    bool
}

func (q LogsQuery) Values() url.Values {
	v := url.Values{}
	if q.From != "" {
		v.Set("from", q.From)
	}
	if q.To != "" {
		v.Set("to", q.To)
	}
	if q.Sort != "" {
		v.Set("sort", q.Sort)
	}
	if q.Limit > 0 {
		v.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.Cursor != "" {
		v.Set("cursor", q.Cursor)
	}
	if q.Device != "" {
		v.Set("device", q.Device)
	}
	if q.Status != "" {
		v.Set("status", q.Status)
	}
	if q.Search != "" {
		v.Set("search", q.Search)
	}
	if q.Raw {
		v.Set("raw", "1")
	}
	return v
}

// Stream Log

const (
	logsStreamPath = "/profiles/%s/logs/stream"

	maxSSELineSize = 1024 * 1024
)

var (
	ErrInvalidSSEContentType = errors.New("invalid SSE content type")
	ErrInvalidSSEEvent       = errors.New("invalid SSE event")
	ErrInvalidLogEvent       = errors.New("invalid log event")
	ErrStreamDisconnected    = errors.New("logs stream disconnected")
)

type LogStream struct {
	body    io.ReadCloser
	scanner *bufio.Scanner

	lastEventID string
}

func (c *Client) StreamLogs(
	ctx context.Context,
	profileID string,
	lastEventID string,
) (*LogStream, error) {
	if profileID == "" {
		return nil, errors.New("profile ID is required")
	}

	path := fmt.Sprintf(logsStreamPath, url.PathEscape(profileID))

	query := url.Values{}

	if lastEventID != "" {
		query.Set("id", lastEventID)
	}

	resp, err := c.doStream(
		ctx,
		http.MethodGet,
		path,
		query,
	)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		return nil, fmt.Errorf("connect logs stream request: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")

	if !isSSEContentType(contentType) {
		resp.Body.Close()

		return nil, fmt.Errorf(
			"%w: got %q",
			ErrInvalidSSEContentType,
			contentType,
		)
	}

	scanner := bufio.NewScanner(resp.Body)

	buffer := make([]byte, 64*1024)
	scanner.Buffer(buffer, maxSSELineSize)

	return &LogStream{
		body:        resp.Body,
		scanner:     scanner,
		lastEventID: lastEventID,
	}, nil
}

func (s *LogStream) Next() (LogEvent, error) {
	var (
		eventID string
		data    []string
	)

	for s.scanner.Scan() {
		line := s.scanner.Text()

		if line == "" {
			if len(data) == 0 {
				continue
			}

			event, err := parseLogEvent(eventID, data)
			if err != nil {
				return LogEvent{}, err
			}

			if event.ID != "" {
				s.lastEventID = event.ID
			}

			return event, nil
		}

		if strings.HasPrefix(line, ":") {
			continue
		}

		field, value, ok := parseSSEField(line)
		if !ok {
			return LogEvent{}, fmt.Errorf(
				"%w: malformed line %q",
				ErrInvalidSSEEvent,
				line,
			)
		}

		switch field {
		case "id":
			eventID = value

		case "data":
			data = append(data, value)

		case "event":
			// NextDNS currently does not document a custom
			// event type. We intentionally ignore it.

		case "retry":
			// Reconnection policy belongs to the service layer.
			// The low-level client does not act on retry.

		default:
			// Unknown SSE fields are ignored according to SSE semantics.
		}
	}

	if err := s.scanner.Err(); err != nil {
		return LogEvent{}, fmt.Errorf(
			"%w: %v",
			ErrStreamDisconnected,
			err,
		)
	}

	return LogEvent{}, fmt.Errorf(
		"%w: %w",
		ErrStreamDisconnected,
		io.EOF,
	)
}

func (s *LogStream) LastEventID() string {
	return s.lastEventID
}

func (s *LogStream) Close() error {
	return s.body.Close()
}

func parseSSEField(line string) (field, value string, ok bool) {
	index := strings.IndexByte(line, ':')

	if index == -1 {
		return "", "", false
	}

	field = line[:index]
	value = line[index+1:]

	value = strings.TrimPrefix(value, " ")

	return field, value, true
}

func parseLogEvent(eventID string, data []string) (LogEvent, error) {
	payload := strings.Join(data, "\n")

	var log Log

	if err := json.Unmarshal([]byte(payload), &log); err != nil {
		return LogEvent{}, fmt.Errorf(
			"%w: %w",
			ErrInvalidLogEvent,
			err,
		)
	}

	if log.Domain == "" {
		return LogEvent{}, fmt.Errorf(
			"%w: domain is empty",
			ErrInvalidLogEvent,
		)
	}

	return LogEvent{
		ID:   eventID,
		Data: log,
	}, nil
}

func isSSEContentType(contentType string) bool {
	mediaType := strings.TrimSpace(
		strings.SplitN(contentType, ";", 2)[0],
	)

	return strings.EqualFold(mediaType, "text/event-stream")
}
