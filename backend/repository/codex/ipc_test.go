package codex

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"testing"
)

func TestClient(t *testing.T) {
	cases := []struct {
		name string
		run  func(*testing.T)
	}{
		{
			name: "検証済みendpointを返す",
			run: func(t *testing.T) {
				client, err := NewClient("unix:///broker.sock")
				if err != nil {
					t.Fatal(err)
				}
				if got := client.Endpoint(); got != "/broker.sock" {
					t.Fatalf("Endpoint() = %q", got)
				}
			},
		},
		{
			name: "不正なendpointを拒否する",
			run: func(t *testing.T) {
				if _, err := NewClient("tcp://127.0.0.1:1"); err == nil {
					t.Fatal("NewClient() succeeded")
				}
			},
		},
		{
			name: "通信失敗を安全に扱う",
			run: func(t *testing.T) {
				client, err := NewClient("unix:///missing/broker.sock")
				if err != nil {
					t.Fatal(err)
				}
				client.dial = func(context.Context, string, string) (net.Conn, error) { return nil, errors.New("unavailable") }
				if got := client.Execute(context.Background(), "prompt"); !got.Unavailable {
					t.Fatalf("result=%+v", got)
				}
			},
		},
		{
			name: "完了済みpromptと安全な結果だけを中継する",
			run: func(t *testing.T) {
				client, err := NewClient("unix:///broker.sock")
				if err != nil {
					t.Fatal(err)
				}
				server, peer := net.Pipe()
				client.dial = func(context.Context, string, string) (net.Conn, error) { return server, nil }
				go func() {
					defer peer.Close()
					var request request
					_ = json.NewDecoder(peer).Decode(&request)
					_ = json.NewEncoder(peer).Encode(Result{HTML: "<html></html>"})
				}()
				if got := client.Execute(context.Background(), "finished prompt"); got.HTML != "<html></html>" {
					t.Fatalf("result=%+v", got)
				}
			},
		},
		{
			name: "不正な要求と通信を拒否する",
			run: func(t *testing.T) {
				if got := (*Client)(nil).Execute(context.Background(), "prompt"); !got.Unavailable {
					t.Fatalf("result=%+v", got)
				}
				client, err := NewClient("unix:///broker.sock")
				if err != nil {
					t.Fatal(err)
				}
				for _, prompt := range []string{"", string(make([]byte, maxIPCMessageBytes+1))} {
					if got := client.Execute(context.Background(), prompt); !got.Unavailable {
						t.Fatalf("result=%+v", got)
					}
				}
				closed, peer := net.Pipe()
				_ = peer.Close()
				client.dial = func(context.Context, string, string) (net.Conn, error) { return closed, nil }
				if got := client.Execute(context.Background(), "prompt"); !got.Unavailable {
					t.Fatalf("result=%+v", got)
				}
				server, peer := net.Pipe()
				client.dial = func(context.Context, string, string) (net.Conn, error) { return server, nil }
				go func() {
					defer peer.Close()
					var request request
					_ = json.NewDecoder(peer).Decode(&request)
				}()
				if got := client.Execute(context.Background(), "prompt"); !got.Unavailable {
					t.Fatalf("result=%+v", got)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, tc.run)
	}
}
