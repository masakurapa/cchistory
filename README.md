# cchistory

A GUI viewer for [Claude Code](https://claude.ai/code) conversation history.

`cchistory` reads the session files that Claude Code stores for the current directory and displays them in a browsable GUI.

## Features

- Lists all Claude Code sessions for the current directory, sorted by last updated
- Session detail view with summary and turn-by-turn message history

## Requirements

- Go 1.25 or later
- macOS (other platforms untested)

## Installation

### Option 1: Install with Go

```sh
go install github.com/masakurapa/cchistory@latest
```

### Option 2: Prebuilt binary (macOS)

```sh
curl -fsSL https://raw.githubusercontent.com/masakurapa/cchistory/main/install.sh | sh
```

## Usage

Run `cchistory` in any directory where you have used Claude Code:

```
cd /your/project
cchistory
```

Claude Code stores session files under `~/.claude/projects/<path>/`. `cchistory` automatically picks up sessions for the current working directory.
