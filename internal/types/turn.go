package types

import (
	"strings"
	"time"
)

type Turn struct {
	User      Message
	Assistant []Message
}

func (t Turn) AssistantContent() string {
	var parts []string
	for _, a := range t.Assistant {
		if a.Content != "" {
			parts = append(parts, a.Content)
		}
	}
	return strings.Join(parts, "\n\n")
}

func (t Turn) FinishedAt() time.Time {
	if len(t.Assistant) == 0 {
		return time.Time{}
	}
	return t.Assistant[len(t.Assistant)-1].Timestamp
}

func (t Turn) Model() string {
	for _, a := range t.Assistant {
		if a.Model != "" {
			return a.Model
		}
	}
	return ""
}

func (t Turn) Effort() string {
	for _, a := range t.Assistant {
		if a.Effort != "" {
			return a.Effort
		}
	}
	return ""
}

func (t Turn) Cancelled() bool {
	return len(t.Assistant) == 0
}

func (t Turn) TotalUsage() Usage {
	var total Usage
	for _, a := range t.Assistant {
		total.InputTokens += a.Usage.InputTokens
		total.OutputTokens += a.Usage.OutputTokens
		total.CacheReadInputTokens += a.Usage.CacheReadInputTokens
		total.CacheCreationInputTokens += a.Usage.CacheCreationInputTokens
	}
	return total
}

func ParseTurns(path string) ([]Turn, error) {
	messages, err := ParseMessages(path)
	if err != nil {
		return nil, err
	}
	var turns []Turn
	for _, msg := range messages {
		switch msg.Role {
		case RoleUser:
			turns = append(turns, Turn{User: msg})
		case RoleAssistant:
			if len(turns) > 0 {
				turns[len(turns)-1].Assistant = append(turns[len(turns)-1].Assistant, msg)
			}
		}
	}
	return turns, nil
}
