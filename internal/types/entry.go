package types

import (
	"bufio"
	"encoding/json"
	"os"
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

	var customTitle, agentName string
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

	if customTitle != "" {
		s.Name = customTitle
	} else {
		s.Name = agentName
	}
	return s, scanner.Err()
}
