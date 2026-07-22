package main

import (
	"log"

	"github.com/masakurapa/cchistory/internal/cc"
	"github.com/masakurapa/cchistory/internal/gui"
)

func main() {
	projectDir, paths, err := cc.SessionFilePaths()
	if err != nil {
		log.Fatal(err)
	}

	if err := gui.Run(projectDir, cc.LoadSessions(paths)); err != nil {
		log.Fatal(err)
	}
}
