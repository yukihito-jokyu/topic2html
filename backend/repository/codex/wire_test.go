package codex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type closeBuffer struct{ bytes.Buffer }

func (*closeBuffer) Close() error { return nil }

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func (failingWriter) Close() error              { return nil }

type limitedWriter struct {
	closeBuffer
	failAt int
	writes int
}

func (w *limitedWriter) Write(value []byte) (int, error) {
	w.writes++
	if w.writes == w.failAt {
		return 0, errors.New("write failed")
	}

	return w.closeBuffer.Write(value)
}

func TestRunWire(t *testing.T) {
	responses := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"result":{}}`,
		`{"jsonrpc":"2.0","method":"thread/settings/updated","params":{"threadId":"thread"}}`,
		`{"jsonrpc":"2.0","id":2,"result":{"thread":{"id":"thread"}}}`,
		`{"jsonrpc":"2.0","method":"turn/started","params":{"threadId":"thread","turnId":"turn"}}`,
		`{"jsonrpc":"2.0","id":3,"result":{"turn":{"id":"turn"}}}`,
		`{"jsonrpc":"2.0","method":"item/started","params":{"threadId":"thread","turnId":"turn","item":{"id":"item","type":"agentMessage"}}}`,
		`{"jsonrpc":"2.0","method":"item/completed","params":{"threadId":"thread","turnId":"turn","item":{"id":"item","type":"agentMessage","text":"<html></html>"}}}`,
		`{"jsonrpc":"2.0","method":"turn/completed","params":{"threadId":"thread","turnId":"turn","status":"completed"}}`,
	}, "\n")
	t.Run("returns the completed agent message", func(t *testing.T) {
		writer := &closeBuffer{}
		got := RunWire(context.Background(), writer, strings.NewReader(responses), "v1", "/work", "prompt")
		if got.HTML != "<html></html>" {
			t.Fatalf("result=%+v", got)
		}
		var requests []struct {
			JSONRPC string         `json:"jsonrpc"`
			ID      *int           `json:"id"`
			Method  string         `json:"method"`
			Params  map[string]any `json:"params"`
		}
		decoder := json.NewDecoder(strings.NewReader(writer.String()))
		for decoder.More() {
			var request struct {
				JSONRPC string         `json:"jsonrpc"`
				ID      *int           `json:"id"`
				Method  string         `json:"method"`
				Params  map[string]any `json:"params"`
			}
			if err := decoder.Decode(&request); err != nil {
				t.Fatal(err)
			}
			requests = append(requests, request)
		}
		if len(requests) != 4 || requests[0].JSONRPC != "2.0" || requests[0].ID == nil || *requests[0].ID != 1 || requests[0].Method != "initialize" || requests[1].ID != nil || requests[1].Method != "initialized" || requests[2].ID == nil || *requests[2].ID != 2 || requests[2].Method != "thread/start" || requests[3].ID == nil || *requests[3].ID != 3 || requests[3].Method != "turn/start" {
			t.Fatalf("requests=%+v", requests)
		}
		if requests[0].Params["clientInfo"].(map[string]any)["name"] != "topic2html-server" || requests[0].Params["clientInfo"].(map[string]any)["version"] != "v1" || requests[2].Params["cwd"] != "/work" || requests[2].Params["approvalPolicy"] != "never" || requests[2].Params["sandbox"] != "read-only" || requests[2].Params["ephemeral"] != true || requests[3].Params["threadId"] != "thread" {
			t.Fatalf("requests=%+v", requests)
		}
	})
	t.Run("replays queued reasoning notifications", func(t *testing.T) {
		queued := strings.Replace(responses,
			`{"jsonrpc":"2.0","id":3,"result":{"turn":{"id":"turn"}}}`,
			strings.Join([]string{
				`{"jsonrpc":"2.0","method":"item/started","params":{"threadId":"thread","turnId":"turn","item":{"id":"reason","type":"reasoning"}}}`,
				`{"jsonrpc":"2.0","method":"item/reasoning/textDelta","params":{"threadId":"thread","turnId":"turn","item":{"id":"reason"}}}`,
				`{"jsonrpc":"2.0","id":3,"result":{"turn":{"id":"turn"}}}`,
			}, "\n"), 1)
		if got := RunWire(context.Background(), &closeBuffer{}, strings.NewReader(queued), "v1", "/work", "prompt"); got.HTML != "<html></html>" {
			t.Fatalf("result=%+v", got)
		}
	})
	t.Run("fails safely for malformed exchanges", func(t *testing.T) {
		for _, reader := range []string{
			"",
			`{"id":9,"result":{}}`,
			`{"id":1,"result":{}}`,
			`{"jsonrpc":"2.0","method":"warning"}` + "\n" + responses,
		} {
			if got := RunWire(context.Background(), &closeBuffer{}, strings.NewReader(reader), "v1", "/work", "prompt"); !got.Unavailable {
				t.Fatalf("reader=%q result=%+v", reader, got)
			}
		}
	})
	t.Run("rejects invalid live JSON-RPC notifications", func(t *testing.T) {
		prefix := strings.Join([]string{
			`{"jsonrpc":"2.0","id":1,"result":{}}`,
			`{"jsonrpc":"2.0","id":2,"result":{"thread":{"id":"thread"}}}`,
			`{"jsonrpc":"2.0","id":3,"result":{"turn":{"id":"turn"}}}`,
		}, "\n")
		for _, notification := range []string{
			`{"method":"turn/started","params":{"threadId":"thread","turnId":"turn"}}`,
			`{"jsonrpc":"1.0","method":"turn/started","params":{"threadId":"thread","turnId":"turn"}}`,
		} {
			if got := RunWire(context.Background(), &closeBuffer{}, strings.NewReader(prefix+"\n"+notification), "v1", "/work", "prompt"); !got.Unavailable {
				t.Fatalf("notification=%q result=%+v", notification, got)
			}
		}
	})
	t.Run("fails safely when writing or context fails", func(t *testing.T) {
		if got := RunWire(context.Background(), failingWriter{}, strings.NewReader(responses), "v1", "/work", "prompt"); !got.Unavailable {
			t.Fatalf("result=%+v", got)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if got := RunWire(ctx, &closeBuffer{}, strings.NewReader(strings.Join([]string{
			`{"jsonrpc":"2.0","id":1,"result":{}}`, `{"jsonrpc":"2.0","id":2,"result":{"thread":{"id":"thread"}}}`, `{"jsonrpc":"2.0","id":3,"result":{"turn":{"id":"turn"}}}`,
		}, "\n")), "v1", "/work", "prompt"); !got.Unavailable {
			t.Fatalf("result=%+v", got)
		}
	})
	t.Run("fails safely at each request boundary", func(t *testing.T) {
		for _, test := range []struct {
			failAt int
			reader string
		}{
			{failAt: 3, reader: `{"jsonrpc":"2.0","id":1,"result":{}}`},
			{failAt: 4, reader: `{"jsonrpc":"2.0","id":1,"result":{}}` + "\n" + `{"jsonrpc":"2.0","id":2,"result":{"thread":{"id":"thread"}}}`},
			{failAt: 99, reader: `{"jsonrpc":"2.0","id":1,"result":{}}` + "\n" + `{"jsonrpc":"2.0","id":2,"result":{"thread":{"id":""}}}`},
			{failAt: 99, reader: `{"jsonrpc":"2.0","id":1,"result":{}}` + "\n" + `{"jsonrpc":"2.0","id":2,"result":{"thread":{"id":"thread"}}}` + "\n" + `{"jsonrpc":"2.0","id":3,"result":{"turn":{"id":""}}}`},
			{failAt: 99, reader: `{"jsonrpc":"2.0","id":1,"result":{}}` + "\n" + `{"jsonrpc":"2.0","method":"thread/status/changed","params":{"threadId":"thread","status":"systemError"}}` + "\n" + `{"jsonrpc":"2.0","id":2,"result":{"thread":{"id":"thread"}}}` + "\n" + `{"jsonrpc":"2.0","id":3,"result":{"turn":{"id":"turn"}}}`},
		} {
			writer := &limitedWriter{failAt: test.failAt}
			if got := RunWire(context.Background(), writer, strings.NewReader(test.reader), "v1", "/work", "prompt"); !got.Unavailable {
				t.Fatalf("test=%+v result=%+v", test, got)
			}
		}
	})
	t.Run("requires a nonempty message before completion", func(t *testing.T) {
		reader := strings.NewReader(strings.Join([]string{
			`{"jsonrpc":"2.0","id":1,"result":{}}`, `{"jsonrpc":"2.0","id":2,"result":{"thread":{"id":"thread"}}}`, `{"jsonrpc":"2.0","id":3,"result":{"turn":{"id":"turn"}}}`, `{"jsonrpc":"2.0","method":"turn/completed"}`,
		}, "\n"))
		if got := RunWire(context.Background(), &closeBuffer{}, reader, "v1", "/work", "prompt"); !got.Unavailable {
			t.Fatalf("result=%+v", got)
		}
	})
}

func message(method, thread, turn, itemID, itemType, text, status string) wireMessage {
	var value wireMessage
	value.JSONRPC = "2.0"
	value.Method, value.Params.ThreadID, value.Params.TurnID, value.Params.Item.ID, value.Params.Item.Type, value.Params.Item.Text, value.Params.Status = method, thread, turn, itemID, itemType, text, status

	return value
}

func TestWireState(t *testing.T) {
	tests := []struct {
		name    string
		message wireMessage
		want    bool
	}{
		{
			name:    "accepts a known global notification",
			message: message("warning", "", "", "", "", "", ""),
			want:    true,
		},
		{
			name:    "rejects an unknown global notification",
			message: message("unknown", "", "", "", "", "", ""),
		},
		{
			name:    "rejects a notification for another thread",
			message: message("item/started", "other", "turn", "item", "agentMessage", "", ""),
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			state := wireState{
				threadID: "thread",
				turnID:   "turn",
			}
			if got := state.accept(testCase.message); got != testCase.want {
				t.Fatalf("accepted=%t want=%t message=%#v", got, testCase.want, testCase.message)
			}
		})
	}
	state := wireState{threadID: "thread", turnID: "turn"}
	for _, value := range []wireMessage{
		message("thread/started", "thread", "", "", "", "", ""),
		message("thread/status/changed", "thread", "", "", "", "", "active"),
		message("thread/tokenUsage/updated", "thread", "turn", "", "", "", ""),
		message("turn/started", "thread", "turn", "", "", "", ""),
		message("item/started", "thread", "turn", "user", "userMessage", "", ""),
		message("item/started", "thread", "turn", "reason", "reasoning", "", ""),
		message("item/reasoning/textDelta", "thread", "turn", "reason", "", "", ""),
		message("item/completed", "thread", "turn", "reason", "reasoning", "", ""),
		message("item/started", "thread", "turn", "agent", "agentMessage", "", ""),
		message("item/agentMessage/delta", "thread", "turn", "agent", "", "", ""),
		message("item/completed", "thread", "turn", "agent", "agentMessage", "<html>", ""),
		message("turn/completed", "thread", "turn", "", "", "", "completed"),
	} {
		if !state.accept(value) {
			t.Fatalf("rejected %#v", value)
		}
	}
	for _, value := range []wireMessage{
		message("unknown", "thread", "turn", "", "", "", ""),
		message("item/started", "other", "turn", "x", "agentMessage", "", ""),
		message("turn/completed", "thread", "turn", "", "", "", "failed"),
	} {
		if state.accept(value) {
			t.Fatalf("accepted %#v", value)
		}
	}
	if state.accept(message("thread/status/changed", "thread", "", "", "", "", "systemError")) {
		t.Fatal("system error accepted")
	}
	if !state.accept(message("thread/status/changed", "thread", "turn", "", "", "", "notLoaded")) {
		t.Fatal("notLoaded status rejected")
	}
	if state.accept(message("thread/status/changed", "thread", "other", "", "", "", "active")) {
		t.Fatal("foreign turn status accepted")
	}
	if state.accept(message("thread/started", "thread", "other", "", "", "", "")) {
		t.Fatal("foreign turn start accepted")
	}
	if !state.accept(message("warning", "", "", "", "", "", "")) {
		t.Fatal("known global warning rejected")
	}
	if state.accept(message("warning", "thread", "", "", "", "", "")) {
		t.Fatal("scoped global warning accepted")
	}
	if state.accept(message("unknown", "", "", "", "", "", "")) {
		t.Fatal("unknown global notification accepted")
	}
	if state.accept(message("thread/settings/updated", "other", "", "", "", "", "")) {
		t.Fatal("foreign setting accepted")
	}
	if state.accept(message("item/started", "thread", "turn", "invalid", "tool", "", "")) {
		t.Fatal("unsupported item accepted")
	}
	if state.accept(message("item/reasoning/delta", "thread", "turn", "other", "", "", "")) {
		t.Fatal("unstarted reasoning accepted")
	}
	errorMessage := message("error", "thread", "turn", "", "", "", "")
	errorMessage.Params.WillRetry = true
	if !state.accept(errorMessage) {
		t.Fatal("retryable error rejected")
	}
	errorMessage.Params.WillRetry = false
	if state.accept(errorMessage) {
		t.Fatal("terminal error accepted")
	}
	if !validQueued([]wireMessage{message("thread/settings/updated", "thread", "", "", "", "", "")}, "thread", "") {
		t.Fatal("valid queued notification rejected")
	}
	if !validQueued([]wireMessage{message("config/updated", "", "", "", "", "", "")}, "thread", "") {
		t.Fatal("known global queued notification rejected")
	}
	if validQueued([]wireMessage{message("turn/started", "thread", "turn", "", "", "", "")}, "thread", "") {
		t.Fatal("turn notification accepted before turn response")
	}
	if validQueued([]wireMessage{message("thread/tokenUsage/updated", "thread", "", "", "", "", "")}, "thread", "") {
		t.Fatal("token usage accepted before turn response")
	}
	for _, value := range []wireMessage{
		message("thread/started", "thread", "", "", "", "", ""),
		message("thread/status/changed", "thread", "turn", "", "", "", ""),
		message("item/started", "thread", "turn", "item", "agentMessage", "", ""),
		message("item/agentMessage/delta", "thread", "turn", "item", "", "", ""),
		message("item/reasoning/textDelta", "thread", "turn", "item", "", "", ""),
	} {
		if !validQueued([]wireMessage{value}, "thread", "turn") {
			t.Fatalf("valid queued notification rejected %#v", value)
		}
	}
	if validQueued([]wireMessage{message("unknown", "thread", "turn", "", "", "", "")}, "thread", "turn") {
		t.Fatal("unknown queued notification accepted")
	}
	if validQueued([]wireMessage{message("thread/started", "other", "", "", "", "", "")}, "thread", "") {
		t.Fatal("foreign queued notification accepted")
	}
	if validQueued([]wireMessage{message("thread/started", "thread", "other", "", "", "", "")}, "thread", "turn") {
		t.Fatal("wrong thread event turn accepted")
	}
	duplicate := wireState{threadID: "thread", turnID: "turn"}
	if !duplicate.accept(message("item/started", "thread", "turn", "agent", "agentMessage", "", "")) {
		t.Fatal("agent start rejected")
	}
	if duplicate.accept(message("item/started", "thread", "turn", "agent", "agentMessage", "", "")) {
		t.Fatal("duplicate agent accepted")
	}
	if !duplicate.accept(message("item/started", "thread", "turn", "user", "userMessage", "", "")) || duplicate.accept(message("item/started", "thread", "turn", "user", "userMessage", "", "")) {
		t.Fatal("duplicate item handling failed")
	}
	if duplicate.accept(message("item/completed", "thread", "turn", "missing", "agentMessage", "", "")) {
		t.Fatal("unknown completion accepted")
	}
	if duplicate.accept(message("item/completed", "thread", "turn", "agent", "agentMessage", "", "")) {
		t.Fatal("empty agent completion accepted")
	}
	if !duplicate.accept(message("item/completed", "thread", "turn", "agent", "agentMessage", "<html>", "")) {
		t.Fatal("agent completion rejected")
	}
	if duplicate.accept(message("item/completed", "thread", "turn", "agent", "agentMessage", "<html>", "")) {
		t.Fatal("duplicate completion accepted")
	}
	large := wireState{threadID: "thread", turnID: "turn"}
	if !large.accept(message("item/started", "thread", "turn", "agent", "agentMessage", "", "")) || large.accept(message("item/completed", "thread", "turn", "agent", "agentMessage", string(make([]byte, maxIPCMessageBytes+1)), "")) {
		t.Fatal("oversized output accepted")
	}
}

func TestResponse(t *testing.T) {
	decoder := json.NewDecoder(strings.NewReader("{\"jsonrpc\":\"2.0\",\"method\":\"thread/settings/updated\",\"params\":{\"threadId\":\"thread\"}}\n{\"jsonrpc\":\"2.0\",\"id\":2,\"result\":{\"thread\":{\"id\":\"thread\"}}}"))
	var target struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	queued, ok := response(decoder, 2, &target)
	if !ok || len(queued) != 1 || target.Thread.ID != "thread" {
		t.Fatalf("queued=%#v target=%#v", queued, target)
	}
	for _, raw := range []string{"", `{"id":3,"result":{}}`, `{"id":2,"error":{}}`, `{"method":""}`} {
		if _, ok := response(json.NewDecoder(strings.NewReader(raw)), 2, nil); ok {
			t.Fatalf("accepted %q", raw)
		}
	}
}
