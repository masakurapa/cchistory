package types

import (
	"strings"
	"time"
)

type Turn struct {
	UserMsg       Message
	AssistantMsgs []Message
}

func (t Turn) Timestamp() time.Time { return t.UserMsg.Timestamp }

func (t Turn) EndTimestamp() time.Time {
	if ft := t.finishedAt(); !ft.IsZero() {
		return ft
	}
	return t.UserMsg.Timestamp
}

func (t Turn) Headline() string {
	line := t.UserMsg.Content
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	return line
}

func (t Turn) Sections() []Section {
	sections := []Section{{Label: "User", Text: t.UserMsg.Content}}
	for _, a := range t.AssistantMsgs {
		if a.Content != "" {
			sections = append(sections, Section{Label: "Assistant", Text: a.Content})
		}
		for _, tc := range a.ToolCalls {
			text := tc.Input
			if tc.Result != "" {
				text += "\n\n" + tc.Result
			}
			sections = append(sections, Section{Label: tc.Name, Text: text})
		}
	}
	return sections
}

func (t Turn) Metadata() []Meta {
	var metas []Meta
	if ft := t.finishedAt(); !ft.IsZero() {
		metas = append(metas, Meta{Name: "Finished", Value: ft.Local().Format("2006-01-02 15:04:05")})
		metas = append(metas, Meta{Name: "Response Time", Value: FormatDuration(ft.Sub(t.UserMsg.Timestamp))})
	}
	if m := t.model(); m != "" {
		metas = append(metas, Meta{Name: "Model", Value: m})
	}
	if e := t.effort(); e != "" {
		metas = append(metas, Meta{Name: "Effort", Value: e})
	}
	if t.thinking() {
		metas = append(metas, Meta{Name: "Thinking", Value: "ON"})
	}
	u := t.TotalUsage()
	ctx := u.InputTokens + u.CacheReadInputTokens + u.CacheCreationInputTokens
	metas = append(metas,
		Meta{Name: "Context Token", Value: FormatTokens(ctx)},
		Meta{Name: "Output Token", Value: FormatTokens(u.OutputTokens)},
	)
	if u.CacheReadInputTokens > 0 {
		metas = append(metas, Meta{Name: "Cache Read Token", Value: FormatTokens(u.CacheReadInputTokens)})
	}
	if u.CacheCreationInputTokens > 0 {
		metas = append(metas, Meta{Name: "Cache Creation Token", Value: FormatTokens(u.CacheCreationInputTokens)})
	}
	return metas
}

func (t Turn) Usage() Usage { return t.TotalUsage() }

func (t Turn) TotalUsage() Usage {
	var output int
	var last Usage
	for _, a := range t.AssistantMsgs {
		if a.Usage.InputTokens+a.Usage.CacheReadInputTokens+a.Usage.CacheCreationInputTokens > 0 {
			last = a.Usage
		}
		output += a.Usage.OutputTokens
	}
	return Usage{
		InputTokens:              last.InputTokens,
		OutputTokens:             output,
		CacheReadInputTokens:     last.CacheReadInputTokens,
		CacheCreationInputTokens: last.CacheCreationInputTokens,
	}
}


func (t Turn) finishedAt() time.Time {
	if len(t.AssistantMsgs) == 0 {
		return time.Time{}
	}
	return t.AssistantMsgs[len(t.AssistantMsgs)-1].Timestamp
}

func (t Turn) model() string {
	for _, a := range t.AssistantMsgs {
		if a.Model != "" {
			return a.Model
		}
	}
	return ""
}

func (t Turn) effort() string {
	for _, a := range t.AssistantMsgs {
		if a.Effort != "" {
			return a.Effort
		}
	}
	return ""
}

func (t Turn) thinking() bool {
	for _, a := range t.AssistantMsgs {
		if a.Thinking {
			return true
		}
	}
	return false
}

