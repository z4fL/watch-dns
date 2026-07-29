package nextdns

import (
	"net/http"
	"time"
)

type Client struct {
	baseUrl    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		baseUrl: "https://api.nextdns.io",
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}
