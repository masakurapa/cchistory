package cc

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/masakurapa/cchistory/internal/types"
)

func SessionFilePaths() (string, []string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil, err
	}
	dirName := strings.ReplaceAll(pwd, "/", "-")
	prjDir := filepath.Join(home, ".claude", "projects", dirName)
	paths, err := sessionFilePaths(prjDir)
	return prjDir, paths, err
}

func sessionFilePaths(projectDir string) ([]string, error) {
	files, err := os.ReadDir(projectDir)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".jsonl" {
			continue
		}
		p := filepath.Join(projectDir, f.Name())
		paths = append(paths, p)
	}
	return paths, nil
}

func LoadSessions(paths []string) []types.Session {
	var sessions []types.Session
	for _, p := range paths {
		s, err := types.ParseSession(p)
		if err != nil {
			continue
		}
		sessions = append(sessions, s)
	}
	return sessions
}
