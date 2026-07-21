package cc

import (
	"os"
	"path/filepath"
	"strings"
)

func SessionFilePaths() ([]string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dirName := strings.ReplaceAll(pwd, "/", "-")
	prjDir := filepath.Join(home, ".claude", "projects", dirName)
	return loadSessions(prjDir)
}

func loadSessions(projectDir string) ([]string, error) {
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
