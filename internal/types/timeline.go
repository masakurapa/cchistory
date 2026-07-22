package types

import (
	"bufio"
	"encoding/json"
	"html"
	"os"
	"strings"
	"time"
)

type CompactBoundary struct {
	Timestamp     time.Time
	Trigger       string
	PreTokens     int
	PostTokens    int
	DroppedTokens int
	DurationMs    int
	Summary       string
}

type LocalCommand struct {
	Timestamp time.Time
	Input     string
	Stdout    string
	Stderr    string
}

type TimelineItem struct {
	Turn            *Turn
	CompactBoundary *CompactBoundary
	LocalCommand    *LocalCommand
}

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

func ParseTimeline(path string) ([]TimelineItem, error) {
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
	var items []TimelineItem
	var currentTurn *Turn
	var lastTurn *Turn // last flushed turn, for post-compact orphan assistant messages
	var pendingLocalCmd *LocalCommand

	flushTurn := func() {
		if currentTurn != nil {
			items = append(items, TimelineItem{Turn: currentTurn})
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
			flushTurn()
			ts, _ := time.Parse(time.RFC3339Nano, entry.Timestamp)
			summary := summaries[entry.CompactMetadata.PreservedSegment.AnchorUuid]
			if summary != "" && lastTurn != nil {
				lastTurn.Assistant = append(lastTurn.Assistant, Message{
					Role:      RoleAssistant,
					Content:   summary,
					Timestamp: ts,
				})
			}
			items = append(items, TimelineItem{CompactBoundary: &CompactBoundary{
				Timestamp:     ts,
				Trigger:       entry.CompactMetadata.Trigger,
				PreTokens:     entry.CompactMetadata.PreTokens,
				PostTokens:    entry.CompactMetadata.PostTokens,
				DroppedTokens: entry.CompactMetadata.CumulativeDroppedTokens,
				DurationMs:    entry.CompactMetadata.DurationMs,
				Summary:       summary,
			}})
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

		// local command handling (user entries only)
		if entry.Type == "user" {
			if strings.HasPrefix(content, "<local-command-caveat>") ||
				strings.HasPrefix(content, "<command-message>") ||
				strings.HasPrefix(content, "<command-name>") ||
				strings.HasPrefix(content, "<command-args>") ||
				strings.HasPrefix(content, "<local-command-stdout>") ||
				strings.HasPrefix(content, "<local-command-stderr>") {
				continue
			}
			if strings.Contains(content, "<bash-input>") {
				ts, _ := time.Parse(time.RFC3339Nano, entry.Timestamp)
				pendingLocalCmd = &LocalCommand{
					Timestamp: ts,
					Input:     extractTagContent(content, "bash-input"),
				}
				continue
			}
			if strings.Contains(content, "<bash-stdout>") || strings.Contains(content, "<bash-stderr>") {
				stdout := extractTagContent(content, "bash-stdout")
				stderr := extractTagContent(content, "bash-stderr")
				if pendingLocalCmd != nil {
					pendingLocalCmd.Stdout = stdout
					pendingLocalCmd.Stderr = stderr
					flushTurn()
					items = append(items, TimelineItem{LocalCommand: pendingLocalCmd})
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
			currentTurn = &Turn{User: msg}
		case RoleAssistant:
			if currentTurn != nil {
				currentTurn.Assistant = append(currentTurn.Assistant, msg)
			} else if lastTurn != nil {
				// compact 後など currentTurn が nil のとき、直前のターンに追加する
				lastTurn.Assistant = append(lastTurn.Assistant, msg)
			}
		}
	}

	if pendingLocalCmd != nil {
		flushTurn()
		items = append(items, TimelineItem{LocalCommand: pendingLocalCmd})
	}
	flushTurn()

	return items, nil
}
