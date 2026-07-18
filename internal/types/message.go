package types

import (
	"bufio"
	"encoding/json"
	"os"
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

type Message struct {
	Role      MessageRole
	Content   string
	Timestamp time.Time
	Model     string
	Effort    string
	Usage     Usage
}

type rawMessageEntry struct {
	Type      string          `json:"type"`
	IsMeta    bool            `json:"isMeta"`
	Timestamp string          `json:"timestamp"`
	Message   json.RawMessage `json:"message"`
	Effort    string          `json:"effort"`
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
	Type string `json:"type"`
	Text string `json:"text"`
}

func ParseMessages(path string) ([]Message, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var messages []Message
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)
	for scanner.Scan() {
		var entry rawMessageEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		if entry.IsMeta || len(entry.Message) == 0 {
			continue
		}
		if entry.Type != "user" && entry.Type != "assistant" {
			continue
		}

		var body rawMessageBody
		if err := json.Unmarshal(entry.Message, &body); err != nil {
			continue
		}

		content := extractTextContent(body.Content)
		if content == "" {
			continue
		}

		ts, _ := time.Parse(time.RFC3339Nano, entry.Timestamp)
		messages = append(messages, Message{
			Role:      MessageRole(body.Role),
			Content:   content,
			Timestamp: ts,
			Model:     body.Model,
			Effort:    entry.Effort,
			Usage: Usage{
				InputTokens:              body.Usage.InputTokens,
				OutputTokens:             body.Usage.OutputTokens,
				CacheReadInputTokens:     body.Usage.CacheReadInputTokens,
				CacheCreationInputTokens: body.Usage.CacheCreationInputTokens,
			},
		})
	}
	return messages, scanner.Err()
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
