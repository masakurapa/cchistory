package types

import "time"

type Meta struct {
	Name  string
	Value string
}

type Section struct {
	Label string
	Text  string
}

type TimelineEntry interface {
	Timestamp() time.Time
	Headline() string
	Sections() []Section
	Metadata() []Meta
	Usage() Usage
}
