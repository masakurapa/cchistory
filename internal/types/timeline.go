package types

import (
	"bufio"
	"encoding/json"
	"html"
	"os"
	"regexp"
	"strings"
	"time"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

type CompactBoundary struct {
	Time          time.Time
	Trigger       string
	PreTokens     int
	PostTokens    int
	DroppedTokens int
	DurationMs    int
	Summary       string
}

func (cb *CompactBoundary) Timestamp() time.Time    { return cb.Time }
func (cb *CompactBoundary) EndTimestamp() time.Time { return cb.Time }

func (cb *CompactBoundary) Headline() string {
	if cb.Trigger == "manual" || cb.Trigger == "user" {
		return "manual compacting"
	}
	return "auto compacting"
}

func (cb *CompactBoundary) Sections() []Section {
	if cb.Summary == "" {
		return nil
	}
	return []Section{{Label: "Summary", Text: cb.Summary}}
}

func (cb *CompactBoundary) Metadata() []Meta {
	metas := []Meta{
		{Name: "Timestamp", Value: cb.Time.Local().Format("2006-01-02 15:04:05")},
	}
	if cb.Trigger != "" {
		metas = append(metas, Meta{Name: "Trigger", Value: cb.Trigger})
	}
	metas = append(metas,
		Meta{Name: "Pre Tokens", Value: FormatTokens(cb.PreTokens)},
		Meta{Name: "Post Tokens", Value: FormatTokens(cb.PostTokens)},
	)
	if cb.DroppedTokens > 0 {
		metas = append(metas, Meta{Name: "Dropped", Value: FormatTokens(cb.DroppedTokens)})
	}
	return metas
}

func (cb *CompactBoundary) Usage() Usage { return Usage{} }

type LocalCommand struct {
	Time   time.Time
	Input  string
	Stdout string
	Stderr string
}

func (lc *LocalCommand) Timestamp() time.Time    { return lc.Time }
func (lc *LocalCommand) EndTimestamp() time.Time { return lc.Time }

func (lc *LocalCommand) Headline() string { return "$ " + lc.Input }

func (lc *LocalCommand) Sections() []Section {
	sections := []Section{{Label: "Command", Text: lc.Input}}
	if lc.Stdout != "" {
		sections = append(sections, Section{Label: "Stdout", Text: lc.Stdout})
	}
	if lc.Stderr != "" {
		sections = append(sections, Section{Label: "Stderr", Text: lc.Stderr})
	}
	return sections
}

func (lc *LocalCommand) Metadata() []Meta { return nil }

type SlashCommand struct {
	Time   time.Time
	Name   string
	Stdout string
}

func (sc *SlashCommand) Timestamp() time.Time    { return sc.Time }
func (sc *SlashCommand) EndTimestamp() time.Time { return sc.Time }
func (sc *SlashCommand) Headline() string        { return sc.Name }

func (sc *SlashCommand) Sections() []Section {
	sections := []Section{{Label: "Command", Text: sc.Name}}
	if sc.Stdout != "" {
		sections = append(sections, Section{Label: "Output", Text: sc.Stdout})
	}
	return sections
}

func (sc *SlashCommand) Metadata() []Meta { return nil }
func (sc *SlashCommand) Usage() Usage     { return Usage{} }

func (lc *LocalCommand) Usage() Usage { return Usage{} }

type rawPreservedSegment struct {
	AnchorUuid string `json:"anchorUuid"`
}

type rawCompactMetadata struct {
	Trigger                 string              `json:"trigger"`
	PreTokens               int                 `json:"preTokens"`
	PostTokens              int                 `json:"postTokens"`
	CumulativeDroppedTokens int                 `json:"cumulativeDroppedTokens"`
	DurationMs              int                 `json:"durationMs"`
	PreservedSegment        rawPreservedSegment `json:"preservedSegment"`
}

type rawTimelineEntry struct {
	Type             string             `json:"type"`
	Subtype          string             `json:"subtype"`
	UUID             string             `json:"uuid"`
	IsMeta           bool               `json:"isMeta"`
	IsCompactSummary bool               `json:"isCompactSummary"`
	Timestamp        string             `json:"timestamp"`
	Message          json.RawMessage    `json:"message"`
	Effort           string             `json:"effort"`
	CompactMetadata  rawCompactMetadata `json:"compactMetadata"`
}

func readLines(path string) ([][]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines [][]byte
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)
	for scanner.Scan() {
		line := make([]byte, len(scanner.Bytes()))
		copy(line, scanner.Bytes())
		lines = append(lines, line)
	}
	return lines, scanner.Err()
}

func extractTagContent(s, tag string) string {
	open := "<" + tag + ">"
	close := "</" + tag + ">"
	start := strings.Index(s, open)
	if start < 0 {
		return ""
	}
	start += len(open)
	end := strings.Index(s[start:], close)
	if end < 0 {
		return html.UnescapeString(strings.TrimSpace(s[start:]))
	}
	return html.UnescapeString(strings.TrimSpace(s[start : start+end]))
}

func ParseTimeline(path string) ([]TimelineEntry, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	// first pass: collect compact summaries keyed by uuid
	summaries := make(map[string]string)
	for _, line := range lines {
		var entry rawTimelineEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		if !entry.IsCompactSummary || len(entry.Message) == 0 {
			continue
		}
		var body rawMessageBody
		if err := json.Unmarshal(entry.Message, &body); err != nil {
			continue
		}
		summaries[entry.UUID] = extractTextContent(body.Content)
	}

	// second pass: build timeline
	var items []TimelineEntry
	var currentTurn *Turn
	var lastTurn *Turn
	var pendingLocalCmd *LocalCommand
	var pendingSlashCmd *SlashCommand

	flushTurn := func() {
		if currentTurn != nil {
			items = append(items, currentTurn)
			lastTurn = currentTurn
			currentTurn = nil
		}
	}

	for _, line := range lines {
		var entry rawTimelineEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		if entry.IsCompactSummary {
			continue
		}

		if entry.Type == "system" && entry.Subtype == "compact_boundary" {
			// Discard slash command turns with no assistant response — they'll be
			// replayed in <command-name> format in the preserved context.
			if currentTurn != nil && len(currentTurn.AssistantMsgs) == 0 &&
				strings.HasPrefix(currentTurn.UserMsg.Content, "/") {
				currentTurn = nil
			}
			flushTurn()
			ts, _ := time.Parse(time.RFC3339Nano, entry.Timestamp)
			summary := summaries[entry.CompactMetadata.PreservedSegment.AnchorUuid]
			if summary != "" && lastTurn != nil {
				lastTurn.AssistantMsgs = append(lastTurn.AssistantMsgs, Message{
					Role:      RoleAssistant,
					Content:   summary,
					Timestamp: ts,
				})
			}
			items = append(items, &CompactBoundary{
				Time:          ts,
				Trigger:       entry.CompactMetadata.Trigger,
				PreTokens:     entry.CompactMetadata.PreTokens,
				PostTokens:    entry.CompactMetadata.PostTokens,
				DroppedTokens: entry.CompactMetadata.CumulativeDroppedTokens,
				DurationMs:    entry.CompactMetadata.DurationMs,
				Summary:       summary,
			})
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

		// tool_result entries: attach results to last assistant message's tool calls
		if entry.Type == "user" && hasToolResultBlocks(body.Content) {
			if currentTurn != nil && len(currentTurn.AssistantMsgs) > 0 {
				attachToolResults(&currentTurn.AssistantMsgs[len(currentTurn.AssistantMsgs)-1], body.Content)
			}
			continue
		}

		content := extractTextContent(body.Content)
		if entry.Type == "user" && content == "" {
			continue
		}

		if entry.Type == "user" {
			if strings.HasPrefix(content, "<local-command-caveat>") ||
				strings.HasPrefix(content, "<command-message>") ||
				strings.HasPrefix(content, "<command-args>") {
				continue
			}
			if strings.Contains(content, "<command-name>") {
				ts, _ := time.Parse(time.RFC3339Nano, entry.Timestamp)
				pendingSlashCmd = &SlashCommand{
					Time: ts,
					Name: extractTagContent(content, "command-name"),
				}
				continue
			}
			if strings.HasPrefix(content, "<local-command-stdout>") || strings.HasPrefix(content, "<local-command-stderr>") {
				if pendingSlashCmd != nil {
					out := ansiEscape.ReplaceAllString(extractTagContent(content, "local-command-stdout"), "")
					pendingSlashCmd.Stdout = strings.TrimSpace(out)
					flushTurn()
					items = append(items, pendingSlashCmd)
					pendingSlashCmd = nil
				}
				continue
			}
			if strings.Contains(content, "<bash-input>") {
				ts, _ := time.Parse(time.RFC3339Nano, entry.Timestamp)
				pendingLocalCmd = &LocalCommand{
					Time:  ts,
					Input: extractTagContent(content, "bash-input"),
				}
				continue
			}
			if strings.Contains(content, "<bash-stdout>") || strings.Contains(content, "<bash-stderr>") {
				if pendingLocalCmd != nil {
					pendingLocalCmd.Stdout = extractTagContent(content, "bash-stdout")
					pendingLocalCmd.Stderr = extractTagContent(content, "bash-stderr")
					flushTurn()
					items = append(items, pendingLocalCmd)
					pendingLocalCmd = nil
				}
				continue
			}
		}

		ts, _ := time.Parse(time.RFC3339Nano, entry.Timestamp)
		msg := Message{
			Role:      MessageRole(body.Role),
			Content:   content,
			Timestamp: ts,
			Model:     body.Model,
			Effort:    entry.Effort,
			Thinking:  hasThinkingBlock(body.Content),
			ToolCalls: extractToolCalls(body.Content),
			Usage: Usage{
				InputTokens:              body.Usage.InputTokens,
				OutputTokens:             body.Usage.OutputTokens,
				CacheReadInputTokens:     body.Usage.CacheReadInputTokens,
				CacheCreationInputTokens: body.Usage.CacheCreationInputTokens,
			},
		}

		switch msg.Role {
		case RoleUser:
			flushTurn()
			currentTurn = &Turn{UserMsg: msg}
		case RoleAssistant:
			if currentTurn != nil {
				currentTurn.AssistantMsgs = append(currentTurn.AssistantMsgs, msg)
			} else if lastTurn != nil {
				lastTurn.AssistantMsgs = append(lastTurn.AssistantMsgs, msg)
			}
		}
	}

	if pendingLocalCmd != nil {
		flushTurn()
		items = append(items, pendingLocalCmd)
	}
	flushTurn()

	return items, nil
}
