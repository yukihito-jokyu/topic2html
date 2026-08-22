package codex

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/url"
	"strings"
)

// ClientはServerからBrokerへのprivate IPC境界である。
type Client struct {
	endpoint string
	dial     func(context.Context, string, string) (net.Conn, error)
}

// Endpointは検証済みのprivate Unix socket pathを返す。
func (c *Client) Endpoint() string {
	return c.endpoint
}

// NewClientは絶対Unix socket endpointだけを受け付ける。
func NewClient(endpoint string) (*Client, error) {
	path, err := unixSocketPath(endpoint)
	if err != nil {
		return nil, err
	}

	return &Client{
		endpoint: path,
		dial:     (&net.Dialer{}).DialContext,
	}, nil
}

// Executeは完了済みpromptだけをBrokerへ送る。
func (c *Client) Execute(ctx context.Context, prompt string) Result {
	if c == nil || strings.TrimSpace(prompt) == "" || len(prompt) > maxIPCMessageBytes {
		return Result{Unavailable: true}
	}
	connection, err := c.dial(ctx, "unix", c.endpoint)
	if err != nil {
		return Result{Unavailable: true}
	}
	defer connection.Close()
	if err := json.NewEncoder(connection).Encode(request{Prompt: prompt}); err != nil {
		return Result{Unavailable: true}
	}
	var response Result
	if err := json.NewDecoder(io.LimitReader(connection, maxIPCMessageBytes)).Decode(&response); err != nil {
		return Result{Unavailable: true}
	}

	return normalize(response)
}

type request struct {
	Prompt string `json:"prompt"`
}

func unixSocketPath(endpoint string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "unix" || u.Host != "" || u.RawQuery != "" || u.Fragment != "" || !strings.HasPrefix(u.Path, "/") {
		return "", errors.New("broker endpoint must be an absolute unix socket")
	}

	return u.Path, nil
}
