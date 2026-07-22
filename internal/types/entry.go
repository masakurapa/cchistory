package types

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type baseEntry struct {
	Type string `json:"type"`
}

type CustomTitleEntry struct {
	CustomTitle string `json:"customTitle"`
	SessionID   string `json:"sessionId"`
}

type AgentNameEntry struct {
	AgentName string `json:"agentName"`
	SessionID string `json:"sessionId"`
}

type AITitleEntry struct {
	AITitle   string `json:"aiTitle"`
	SessionID string `json:"sessionId"`
}

// Session holds parsed metadata from a single JSONL file.
type Session struct {
	ID      string
	Name    string
	ModTime time.Time
}

// ParseSession reads a JSONL file and extracts session metadata.
// custom-title takes priority over agent-name for the session name.
func ParseSession(path string) (Session, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Session{}, err
	}

	f, err := os.Open(path)
	if err != nil {
		return Session{}, err
	}
	defer f.Close()

	s := Session{ModTime: info.ModTime()}

	var customTitle, aiTitle, agentName string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		var base baseEntry
		if err := json.Unmarshal(scanner.Bytes(), &base); err != nil {
			continue
		}
		switch base.Type {
		case "custom-title":
			var e CustomTitleEntry
			if err := json.Unmarshal(scanner.Bytes(), &e); err == nil {
				customTitle = e.CustomTitle
				if s.ID == "" {
					s.ID = e.SessionID
				}
			}
		case "ai-title":
			var e AITitleEntry
			if err := json.Unmarshal(scanner.Bytes(), &e); err == nil {
				if e.AITitle != "" {
					aiTitle = e.AITitle
				}
				if s.ID == "" {
					s.ID = e.SessionID
				}
			}
		case "agent-name":
			var e AgentNameEntry
			if err := json.Unmarshal(scanner.Bytes(), &e); err == nil {
				agentName = e.AgentName
				if s.ID == "" {
					s.ID = e.SessionID
				}
			}
		}
	}

	if s.ID == "" {
		s.ID = strings.TrimSuffix(filepath.Base(path), ".jsonl")
	}
	switch {
	case customTitle != "":
		s.Name = customTitle
	case aiTitle != "":
		s.Name = aiTitle
	case agentName != "":
		s.Name = agentName
	}
	return s, scanner.Err()
}
