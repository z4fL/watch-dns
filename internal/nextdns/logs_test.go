package nextdns

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestStreamLogs(t *testing.T) {
	tests := []struct {
		name string
		body string

		contentType string
		statusCode  int

		lastEventID string

		wantEventID string
		wantDomain  string
		wantErr     error

		checkRequest func(t *testing.T, r *http.Request)
	}{
		{
			name: "successful event",
			body: "id: abc123\n" +
				"data: {\"timestamp\":\"2026-07-27T10:00:00Z\",\"domain\":\"example.com\",\"status\":\"default\"}\n\n",

			contentType: "text/event-stream",
			statusCode:  http.StatusOK,

			wantEventID: "abc123",
			wantDomain:  "example.com",

			checkRequest: func(t *testing.T, r *http.Request) {
				if got := r.Header.Get("X-Api-Key"); got != "test-api-key" {
					t.Fatalf("X-Api-Key = %q, want %q", got, "test-api-key")
				}

				if got := r.Header.Get("Accept"); got != "text/event-stream" {
					t.Fatalf("Accept = %q, want %q", got, "text/event-stream")
				}
			},
		},
		{
			name: "resume from last event id",
			body: "id: next123\n" +
				"data: {\"timestamp\":\"2026-07-27T10:00:00Z\",\"domain\":\"example.com\",\"status\":\"default\"}\n\n",

			contentType: "text/event-stream",
			statusCode:  http.StatusOK,

			lastEventID: "previous123",

			wantEventID: "next123",
			wantDomain:  "example.com",

			checkRequest: func(t *testing.T, r *http.Request) {
				if got := r.URL.Query().Get("id"); got != "previous123" {
					t.Fatalf("query id = %q, want %q", got, "previous123")
				}
			},
		},
		{
			name: "invalid content type",
			body: "id: abc123\n" +
				"data: {\"domain\":\"example.com\"}\n\n",

			contentType: "application/json",
			statusCode:  http.StatusOK,

			wantErr: ErrInvalidSSEContentType,
		},
		{
			name: "invalid JSON payload",
			body: "id: abc123\n" +
				"data: {invalid-json}\n\n",

			contentType: "text/event-stream",
			statusCode:  http.StatusOK,

			wantErr: ErrInvalidLogEvent,
		},
		{
			name: "invalid HTTP status",
			body: `{"errors":[{"code":"invalid","detail":"invalid profile"}]}`,

			contentType: "application/json",
			statusCode:  http.StatusBadRequest,

			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if tt.checkRequest != nil {
						tt.checkRequest(t, r)
					}

					w.Header().Set("Content-Type", tt.contentType)
					w.WriteHeader(tt.statusCode)

					_, _ = fmt.Fprint(w, tt.body)
				},
			))
			defer server.Close()

			client := &Client{
				baseUrl:    server.URL,
				apiKey:     "test-api-key",
				httpClient: server.Client(),
			}

			ctx := context.Background()

			stream, err := client.StreamLogs(
				ctx,
				"abc123",
				tt.lastEventID,
			)

			if tt.name == "invalid HTTP status" {
				if err == nil {
					t.Fatal("expected HTTP error")
				}

				return
			}

			if tt.wantErr != nil {
				if err == nil {
					if stream != nil {
						_ = stream.Close()
					}

					t.Fatal("expected error")
				}

				if !errors.Is(err, tt.wantErr) {
					t.Fatalf(
						"error = %v, want wrapped %v",
						err,
						tt.wantErr,
					)
				}

				return
			}

			if err != nil {
				t.Fatalf("StreamLogs() error = %v", err)
			}

			defer stream.Close()

			event, err := stream.Next()
			if err != nil {
				t.Fatalf("Next() error = %v", err)
			}

			if event.ID != tt.wantEventID {
				t.Fatalf(
					"event ID = %q, want %q",
					event.ID,
					tt.wantEventID,
				)
			}

			if event.Data.Domain != tt.wantDomain {
				t.Fatalf(
					"domain = %q, want %q",
					event.Data.Domain,
					tt.wantDomain,
				)
			}

			if stream.LastEventID() != tt.wantEventID {
				t.Fatalf(
					"LastEventID() = %q, want %q",
					stream.LastEventID(),
					tt.wantEventID,
				)
			}
		})
	}
}

func TestLogStream_MultipleEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(
				"Content-Type",
				"text/event-stream",
			)

			_, _ = fmt.Fprint(
				w,
				"id: first\n"+
					"data: {\"timestamp\":\"2026-07-27T10:00:00Z\",\"domain\":\"first.com\",\"status\":\"default\"}\n\n"+
					"id: second\n"+
					"data: {\"timestamp\":\"2026-07-27T10:01:00Z\",\"domain\":\"second.com\",\"status\":\"blocked\"}\n\n",
			)
		},
	))
	defer server.Close()

	client := &Client{
		baseUrl:    server.URL,
		apiKey:     "test-api-key",
		httpClient: server.Client(),
	}

	stream, err := client.StreamLogs(
		context.Background(),
		"abc123",
		"",
	)
	if err != nil {
		t.Fatalf("StreamLogs() error = %v", err)
	}

	defer stream.Close()

	first, err := stream.Next()
	if err != nil {
		t.Fatalf("first Next() error = %v", err)
	}

	second, err := stream.Next()
	if err != nil {
		t.Fatalf("second Next() error = %v", err)
	}

	if first.ID != "first" {
		t.Fatalf("first ID = %q, want %q", first.ID, "first")
	}

	if second.ID != "second" {
		t.Fatalf("second ID = %q, want %q", second.ID, "second")
	}

	if stream.LastEventID() != "second" {
		t.Fatalf(
			"LastEventID() = %q, want %q",
			stream.LastEventID(),
			"second",
		)
	}
}

func TestLogStream_MultilineData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(
				"Content-Type",
				"text/event-stream",
			)

			_, _ = fmt.Fprint(
				w,
				"id: abc123\n"+
					"data: {\"timestamp\":\"2026-07-27T10:00:00Z\",\n"+
					"data: \"domain\":\"example.com\",\"status\":\"default\"}\n\n",
			)
		},
	))
	defer server.Close()

	client := &Client{
		baseUrl:    server.URL,
		apiKey:     "test-api-key",
		httpClient: server.Client(),
	}

	stream, err := client.StreamLogs(
		context.Background(),
		"abc123",
		"",
	)
	if err != nil {
		t.Fatalf("StreamLogs() error = %v", err)
	}

	defer stream.Close()

	_, err = stream.Next()

	if err == nil {
		t.Fatal("expected invalid JSON because SSE data lines contain newline")
	}

	if !errors.Is(err, ErrInvalidLogEvent) {
		t.Fatalf("error = %v, want ErrInvalidLogEvent", err)
	}
}

func TestLogStream_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(
				"Content-Type",
				"text/event-stream",
			)

			<-r.Context().Done()
		},
	))
	defer server.Close()

	client := &Client{
		baseUrl:    server.URL,
		apiKey:     "test-api-key",
		httpClient: server.Client(),
	}

	ctx, cancel := context.WithCancel(context.Background())

	stream, err := client.StreamLogs(
		ctx,
		"abc123",
		"",
	)
	if err != nil {
		t.Fatalf("StreamLogs() error = %v", err)
	}

	done := make(chan error, 1)

	go func() {
		_, err := stream.Next()
		done <- err
	}()

	time.Sleep(50 * time.Millisecond)

	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected context cancellation error")
		}

	case <-time.After(time.Second):
		t.Fatal("stream did not stop after context cancellation")
	}

	_ = stream.Close()
}

func TestLogStream_Disconnect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(
				"Content-Type",
				"text/event-stream",
			)

			_, _ = fmt.Fprint(
				w,
				"id: abc123\n"+
					"data: {\"timestamp\":\"2026-07-27T10:00:00Z\",\"domain\":\"example.com\",\"status\":\"default\"}\n\n",
			)
		},
	))
	defer server.Close()

	client := &Client{
		baseUrl:    server.URL,
		apiKey:     "test-api-key",
		httpClient: server.Client(),
	}

	stream, err := client.StreamLogs(
		context.Background(),
		"abc123",
		"",
	)
	if err != nil {
		t.Fatalf("StreamLogs() error = %v", err)
	}

	defer stream.Close()

	_, err = stream.Next()
	if err != nil {
		t.Fatalf("first Next() error = %v", err)
	}

	_, err = stream.Next()

	if !errors.Is(err, ErrStreamDisconnected) {
		t.Fatalf(
			"error = %v, want ErrStreamDisconnected",
			err,
		)
	}

	if !errors.Is(err, io.EOF) {
		t.Fatalf(
			"error = %v, want io.EOF",
			err,
		)
	}
}

func TestParseSSEField(t *testing.T) {
	tests := []struct {
		line  string
		field string
		value string
		ok    bool
	}{
		{
			line:  "id: abc123",
			field: "id",
			value: "abc123",
			ok:    true,
		},
		{
			line:  "data: hello",
			field: "data",
			value: "hello",
			ok:    true,
		},
		{
			line:  "event: message",
			field: "event",
			value: "message",
			ok:    true,
		},
		{
			line: "malformed",
			ok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			field, value, ok := parseSSEField(tt.line)

			if field != tt.field {
				t.Fatalf(
					"field = %q, want %q",
					field,
					tt.field,
				)
			}

			if value != tt.value {
				t.Fatalf(
					"value = %q, want %q",
					value,
					tt.value,
				)
			}

			if ok != tt.ok {
				t.Fatalf(
					"ok = %v, want %v",
					ok,
					tt.ok,
				)
			}
		})
	}
}

func TestIsSSEContentType(t *testing.T) {
	tests := []struct {
		contentType string
		want        bool
	}{
		{
			contentType: "text/event-stream",
			want:        true,
		},
		{
			contentType: "text/event-stream; charset=utf-8",
			want:        true,
		},
		{
			contentType: "application/json",
			want:        false,
		},
		{
			contentType: "",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			if got := isSSEContentType(tt.contentType); got != tt.want {
				t.Fatalf(
					"isSSEContentType() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestStreamLogs_InvalidProfileID(t *testing.T) {
	client := &Client{
		baseUrl:    "http://localhost",
		apiKey:     "test-api-key",
		httpClient: http.DefaultClient,
	}

	_, err := client.StreamLogs(
		context.Background(),
		"",
		"",
	)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStreamLogs_ContextAlreadyCanceled(t *testing.T) {
	client := &Client{
		baseUrl:    "http://localhost",
		apiKey:     "test-api-key",
		httpClient: http.DefaultClient,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.StreamLogs(
		ctx,
		"abc123",
		"",
	)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"error = %v, want context.Canceled",
			err,
		)
	}
}

func TestStreamLogs_ResumeQueryEscaping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			expected := "abc/123?test"

			if got := r.URL.Query().Get("id"); got != expected {
				t.Fatalf(
					"id = %q, want %q",
					got,
					expected,
				)
			}

			w.Header().Set(
				"Content-Type",
				"text/event-stream",
			)

			_, _ = fmt.Fprint(
				w,
				"id: next\n"+
					"data: {\"domain\":\"example.com\"}\n\n",
			)
		},
	))
	defer server.Close()

	client := &Client{
		baseUrl:    server.URL,
		apiKey:     "test-api-key",
		httpClient: server.Client(),
	}

	stream, err := client.StreamLogs(
		context.Background(),
		"abc123",
		"abc/123?test",
	)
	if err != nil {
		t.Fatalf("StreamLogs() error = %v", err)
	}

	defer stream.Close()

	_, err = stream.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
}

func TestLogStream_LastEventID(t *testing.T) {
	stream := &LogStream{
		lastEventID: "initial",
	}

	if got := stream.LastEventID(); got != "initial" {
		t.Fatalf(
			"LastEventID() = %q, want %q",
			got,
			"initial",
		)
	}
}

func TestParseLogEvent_InvalidDomain(t *testing.T) {
	_, err := parseLogEvent(
		"abc123",
		[]string{`{"status":"default"}`},
	)

	if !errors.Is(err, ErrInvalidLogEvent) {
		t.Fatalf(
			"error = %v, want ErrInvalidLogEvent",
			err,
		)
	}
}

func TestParseLogEvent_Timestamp(t *testing.T) {
	event, err := parseLogEvent(
		"abc123",
		[]string{
			`{"timestamp":"2026-07-27T10:00:00Z","domain":"example.com","status":"default"}`,
		},
	)
	if err != nil {
		t.Fatalf("parseLogEvent() error = %v", err)
	}

	want := time.Date(
		2026,
		time.July,
		27,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	if !event.Data.Timestamp.Equal(want) {
		t.Fatalf(
			"Timestamp = %v, want %v",
			event.Data.Timestamp,
			want,
		)
	}
}

func TestStreamLogs_ProfileIDEscaping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(
				"Content-Type",
				"text/event-stream",
			)

			if r.URL.Path == "/profiles/abc%2F123/logs/stream" {
				t.Fatal("Path has already been unescaped")
			}

			_, _ = fmt.Fprint(
				w,
				"id: abc\n"+
					"data: {\"domain\":\"example.com\"}\n\n",
			)
		},
	))
	defer server.Close()

	client := &Client{
		baseUrl:    server.URL,
		apiKey:     "test-api-key",
		httpClient: server.Client(),
	}

	stream, err := client.StreamLogs(
		context.Background(),
		"abc/123",
		"",
	)
	if err != nil {
		t.Fatalf("StreamLogs() error = %v", err)
	}

	defer stream.Close()

	_, err = stream.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
}

func TestLogStream_Close(t *testing.T) {
	stream := &LogStream{
		body: io.NopCloser(nil),
	}

	if err := stream.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestNewRequestURL(t *testing.T) {
	values := url.Values{}
	values.Set("id", "abc123")

	if got := values.Get("id"); got != "abc123" {
		t.Fatalf("id = %q, want %q", got, "abc123")
	}
}
