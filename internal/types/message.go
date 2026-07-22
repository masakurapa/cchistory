package types

import (
	"encoding/json"
	"strings"
	"time"
)

type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

type Usage struct {
	InputTokens              int
	OutputTokens             int
	CacheReadInputTokens     int
	CacheCreationInputTokens int
}

type ToolCall struct {
	ID     string
	Name   string
	Input  string
	Result string
}

type Message struct {
	Role      MessageRole
	Content   string
	Timestamp time.Time
	Model     string
	Effort    string
	Thinking  bool
	ToolCalls []ToolCall
	Usage     Usage
}

type rawUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
}

type rawMessageBody struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
	Model   string          `json:"model"`
	Usage   rawUsage        `json:"usage"`
}

type rawContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text"`
	Thinking  string          `json:"thinking"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
	ToolUseID string          `json:"tool_use_id"`
	Content   json.RawMessage `json:"content"`
}

var toolPrimaryField = map[string]string{
	"Bash":      "command",
	"Read":      "file_path",
	"Edit":      "file_path",
	"Write":     "file_path",
	"WebSearch": "query",
	"WebFetch":  "url",
	"Grep":      "pattern",
	"Glob":      "pattern",
	"LS":        "path",
	"Agent":     "description",
}

func formatToolInput(name string, raw json.RawMessage) string {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil || len(m) == 0 {
		return strings.TrimSpace(string(raw))
	}
	if field, ok := toolPrimaryField[name]; ok {
		if v, ok := m[field]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err == nil {
				return strings.TrimSpace(s)
			}
		}
	}
	b, _ := json.MarshalIndent(raw, "", "  ")
	return string(b)
}

func hasThinkingBlock(raw json.RawMessage) bool {
	var blocks []rawContentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return false
	}
	for _, b := range blocks {
		if b.Type == "thinking" {
			return true
		}
	}
	return false
}

func hasToolResultBlocks(raw json.RawMessage) bool {
	var blocks []rawContentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return false
	}
	for _, b := range blocks {
		if b.Type == "tool_result" {
			return true
		}
	}
	return false
}

func extractToolCalls(raw json.RawMessage) []ToolCall {
	var blocks []rawContentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return nil
	}
	var calls []ToolCall
	for _, b := range blocks {
		if b.Type == "tool_use" {
			calls = append(calls, ToolCall{
				ID:    b.ID,
				Name:  b.Name,
				Input: formatToolInput(b.Name, b.Input),
			})
		}
	}
	return calls
}

func attachToolResults(msg *Message, raw json.RawMessage) {
	var blocks []rawContentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return
	}
	results := make(map[string]string, len(blocks))
	for _, b := range blocks {
		if b.Type == "tool_result" {
			results[b.ToolUseID] = extractTextContent(b.Content)
		}
	}
	for i := range msg.ToolCalls {
		if r, ok := results[msg.ToolCalls[i].ID]; ok {
			msg.ToolCalls[i].Result = r
		}
	}
}

func extractTextContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}
	var blocks []rawContentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return ""
	}
	var parts []string
	for _, b := range blocks {
		if b.Type == "text" {
			if t := strings.TrimSpace(b.Text); t != "" {
				parts = append(parts, t)
			}
		}
	}
	return strings.Join(parts, "\n")
}
