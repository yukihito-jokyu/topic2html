package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"strings"
)

// RunWireは固定されたapp-server v2通信を実行する。
func RunWire(ctx context.Context, writer io.WriteCloser, reader io.Reader, version, workdir, prompt string) Result {
	defer writer.Close()
	encoder, decoder := json.NewEncoder(writer), json.NewDecoder(bufio.NewReader(reader))
	if !writeRequest(encoder, 1, "initialize", map[string]any{
		"clientInfo": map[string]string{
			"name":    "topic2html-server",
			"version": version,
		},
	}) {
		return unavailable()
	}
	initializeQueued, ok := response(decoder, 1, nil)
	if !ok || len(initializeQueued) != 0 || encoder.Encode(map[string]any{
		"jsonrpc": "2.0",
		"method":  "initialized",
	}) != nil {
		return unavailable()
	}
	var thread struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if !writeRequest(encoder, 2, "thread/start", map[string]any{
		"cwd":            workdir,
		"approvalPolicy": "never",
		"sandbox":        "read-only",
		"ephemeral":      true,
	}) {
		return unavailable()
	}
	threadQueued, ok := response(decoder, 2, &thread)
	if !ok || thread.Thread.ID == "" || !validQueued(threadQueued, thread.Thread.ID, "") {
		return unavailable()
	}
	var turn struct {
		Turn struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	input := []map[string]string{
		{
			"type": "text",
			"text": prompt,
		},
	}
	if !writeRequest(encoder, 3, "turn/start", map[string]any{
		"threadId": thread.Thread.ID,
		"input":    input,
	}) {
		return unavailable()
	}
	turnQueued, ok := response(decoder, 3, &turn)
	if !ok || turn.Turn.ID == "" || !validQueued(turnQueued, thread.Thread.ID, turn.Turn.ID) {
		return unavailable()
	}
	state := wireState{
		threadID: thread.Thread.ID,
		turnID:   turn.Turn.ID,
	}
	for _, m := range append(threadQueued, turnQueued...) {
		if !state.accept(m) {
			return unavailable()
		}
	}
	for {
		select {
		case <-ctx.Done():
			return unavailable()
		default:
		}
		var m wireMessage
		if decoder.Decode(&m) != nil || m.JSONRPC != "2.0" || m.Method == "" || !state.accept(m) {
			return unavailable()
		}
		if m.Method == "turn/completed" {
			return Result{
				HTML: state.html,
			}
		}
	}
}

type wireState struct {
	threadID  string
	turnID    string
	agentID   string
	html      string
	items     map[string]string
	completed map[string]struct{}
}

func (s *wireState) accept(m wireMessage) bool {
	if m.Method == "error" {
		return m.Params.ThreadID == s.threadID && m.Params.TurnID == s.turnID && m.Params.WillRetry
	}
	if m.Method == "thread/status/changed" {
		return m.Params.ThreadID == s.threadID && (m.Params.TurnID == "" || m.Params.TurnID == s.turnID) && (m.Params.Status == "active" || m.Params.Status == "idle" || m.Params.Status == "notLoaded")
	}
	if m.Method == "thread/settings/updated" || m.Method == "thread/tokenUsage/updated" {
		return m.Params.ThreadID == s.threadID && (m.Params.TurnID == "" || m.Params.TurnID == s.turnID)
	}
	if m.Method == "thread/started" {
		return m.Params.ThreadID == s.threadID && (m.Params.TurnID == "" || m.Params.TurnID == s.turnID)
	}
	if isIgnoredGlobal(m) {
		return true
	}
	if m.Params.ThreadID != s.threadID || m.Params.TurnID != s.turnID {
		return false
	}
	if strings.HasPrefix(m.Method, "item/reasoning/") {
		return m.Params.Item.ID != "" && s.items[m.Params.Item.ID] == "reasoning"
	}
	switch m.Method {
	case "turn/started":
		return true
	case "item/started":
		if m.Params.Item.ID == "" || (m.Params.Item.Type != "userMessage" && m.Params.Item.Type != "reasoning" && m.Params.Item.Type != "agentMessage") {
			return false
		}
		if m.Params.Item.Type == "agentMessage" {
			if s.agentID != "" {
				return false
			}
			s.agentID = m.Params.Item.ID
		}
		if s.items == nil {
			s.items = make(map[string]string)
		}
		if _, exists := s.items[m.Params.Item.ID]; exists {
			return false
		}
		s.items[m.Params.Item.ID] = m.Params.Item.Type

		return true
	case "item/agentMessage/delta":
		return s.agentID != "" && m.Params.Item.ID == s.agentID && s.items[m.Params.Item.ID] == "agentMessage"
	case "item/completed":
		if s.items[m.Params.Item.ID] != m.Params.Item.Type {
			return false
		}
		if s.completed == nil {
			s.completed = make(map[string]struct{})
		}
		if _, exists := s.completed[m.Params.Item.ID]; exists {
			return false
		}
		if m.Params.Item.Type != "agentMessage" {
			s.completed[m.Params.Item.ID] = struct{}{}

			return true
		}
		if m.Params.Item.ID != s.agentID || m.Params.Item.Text == "" || len(m.Params.Item.Text) > maxIPCMessageBytes || s.html != "" {
			return false
		}
		s.html = m.Params.Item.Text
		s.completed[m.Params.Item.ID] = struct{}{}

		return true
	case "turn/completed":
		return m.Params.Status == "completed" && s.html != ""
	default:
		return false
	}
}
func response(d *json.Decoder, id int, target any) ([]wireMessage, bool) {
	var queued []wireMessage
	for {
		var raw json.RawMessage
		if d.Decode(&raw) != nil {
			return nil, false
		}
		var r struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      *int            `json:"id"`
			Result  json.RawMessage `json:"result"`
			Error   json.RawMessage `json:"error"`
		}
		if json.Unmarshal(raw, &r) == nil && r.ID != nil {
			if r.JSONRPC != "2.0" || *r.ID != id || len(r.Result) == 0 || len(r.Error) != 0 {
				return nil, false
			}

			return queued, target == nil || json.Unmarshal(r.Result, target) == nil
		}
		var m wireMessage
		if json.Unmarshal(raw, &m) != nil || m.JSONRPC != "2.0" || m.Method == "" {
			return nil, false
		}
		queued = append(queued, m)
	}
}
func validQueued(ms []wireMessage, threadID, turnID string) bool {
	for _, m := range ms {
		if !validQueuedMessage(m, threadID, turnID) {
			return false
		}
	}

	return true
}

func validQueuedMessage(m wireMessage, threadID, turnID string) bool {
	if isIgnoredGlobal(m) {
		return true
	}
	if m.Params.ThreadID != threadID {
		return false
	}
	if turnID == "" {
		return m.Params.TurnID == "" && (m.Method == "thread/started" || m.Method == "thread/status/changed" || m.Method == "thread/settings/updated")
	}
	if m.Method == "thread/status/changed" || m.Method == "thread/settings/updated" || m.Method == "thread/tokenUsage/updated" || m.Method == "thread/started" {
		return m.Params.TurnID == "" || m.Params.TurnID == turnID
	}

	return m.Params.TurnID == turnID && (m.Method == "turn/started" || m.Method == "item/started" || m.Method == "item/completed" || m.Method == "turn/completed" || m.Method == "error" || strings.HasPrefix(m.Method, "item/reasoning/") || m.Method == "item/agentMessage/delta")
}
func isIgnoredGlobal(m wireMessage) bool {
	if m.Params.ThreadID != "" || m.Params.TurnID != "" {
		return false
	}

	switch m.Method {
	case "warning", "config/updated", "account/updated", "model/list/updated":
		return true
	default:
		return false
	}
}
func unavailable() Result {
	return Result{
		Unavailable: true,
	}
}
func writeRequest(e *json.Encoder, id int, method string, params any) bool {
	return e.Encode(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}) == nil
}

type wireMessage struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		ThreadID  string `json:"threadId"`
		TurnID    string `json:"turnId"`
		Status    string `json:"status"`
		WillRetry bool   `json:"willRetry"`
		Item      struct {
			ID   string `json:"id"`
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"item"`
	} `json:"params"`
}
