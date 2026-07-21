package main

import (
	"log"

	"github.com/masakurapa/cchist/internal/cc"
	"github.com/masakurapa/cchist/internal/gui"
	"github.com/masakurapa/cchist/internal/types"
)

func main() {
	paths, err := cc.SessionFilePaths()
	if err != nil {
		log.Fatal(err)
	}

	sessions, err := loadSessions(paths)
	if err != nil {
		log.Fatal(err)
	}
	if err := gui.Run(projectDir, sessions); err != nil {
		log.Fatal(err)
	}
}

func loadSessions(projectDirs []string) ([]types.Session, error) {
	var sessions []types.Session
	for _, f := range projectDirs {
		s, err := types.ParseSession(f)
		if err != nil {
			continue
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}
