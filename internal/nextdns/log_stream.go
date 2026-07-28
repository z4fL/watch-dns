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
	"strings"
	"time"
)

const (
	logsStreamPath = "/profiles/%s/logs/stream"

	maxSSELineSize    = 1024 * 1024
	initialRetryDelay = 1 * time.Second
	maxRetryDelay     = 30 * time.Second
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
	before, after, ok := strings.Cut(line, ":")

	if !ok {
		return "", "", false
	}

	field = before
	value = after

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

func (c *Client) StreamLogsWithReconnect(
	ctx context.Context,
	profileID string,
	handler func(LogEvent) error,
) error {
	if profileID == "" {
		return errors.New("profile ID is required")
	}

	if handler == nil {
		return errors.New("log event handler is required")
	}

	var lastEventID string
	retryDelay := initialRetryDelay

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		stream, err := c.StreamLogs(
			ctx,
			profileID,
			lastEventID,
		)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err := waitForRetry(ctx, retryDelay); err != nil {
				return err
			}

			retryDelay = nextRetryDelay(retryDelay)
			continue
		}

		retryDelay = initialRetryDelay

		err = consumeLogStream(
			ctx,
			stream,
			handler,
			&lastEventID,
		)

		closeErr := stream.Close()

		if err == nil {
			err = closeErr
		}

		if err != nil && !errors.Is(err, ErrStreamDisconnected) {
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err := waitForRetry(ctx, retryDelay); err != nil {
			return err
		}

		retryDelay = nextRetryDelay(retryDelay)
	}
}

func consumeLogStream(
	ctx context.Context,
	stream *LogStream,
	handler func(LogEvent) error,
	lastEventID *string,
) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		event, err := stream.Next()
		if err != nil {
			return err
		}

		if event.ID != "" {
			*lastEventID = event.ID
		}

		if err := handler(event); err != nil {
			return fmt.Errorf(
				"handle log event: %w",
				err,
			)
		}
	}
}

func waitForRetry(
	ctx context.Context,
	delay time.Duration,
) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-timer.C:
		return nil
	}
}

func nextRetryDelay(current time.Duration) time.Duration {
	next := current * 2

	if next > maxRetryDelay {
		return maxRetryDelay
	}

	return next
}
