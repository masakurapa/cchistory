package types

import (
	"fmt"
	"time"
)

type SessionDetail struct {
	Title string
	Metas []Meta
	Items []TimelineEntry
}

func ParseSessionDetail(path string, session Session) (SessionDetail, error) {
	items, err := ParseTimeline(path)
	if err != nil {
		return SessionDetail{}, err
	}

	var totalInput, totalOutput, totalCacheRead, totalCacheCreation, turnCount int
	var start, end time.Time
	for _, entry := range items {
		if _, ok := entry.(*Turn); ok {
			turnCount++
		}
		ts := entry.Timestamp()
		if start.IsZero() {
			start = ts
		}
		if et := entry.EndTimestamp(); et.After(end) {
			end = et
		}
		u := entry.Usage()
		totalInput += u.InputTokens
		totalCacheRead += u.CacheReadInputTokens
		totalCacheCreation += u.CacheCreationInputTokens
		totalOutput += u.OutputTokens
	}
	total := Usage{
		InputTokens:              totalInput,
		OutputTokens:             totalOutput,
		CacheReadInputTokens:     totalCacheRead,
		CacheCreationInputTokens: totalCacheCreation,
	}

	title := session.ID
	if session.Name != "" {
		title = session.Name
	}

	ctx := total.InputTokens + total.CacheReadInputTokens + total.CacheCreationInputTokens
	metas := []Meta{
		{Name: "ID", Value: session.ID},
	}
	if session.Name != "" {
		metas = append(metas, Meta{Name: "Name", Value: session.Name})
	}
	if !start.IsZero() {
		metas = append(metas, Meta{Name: "Started", Value: start.Local().Format("2006-01-02 15:04:05")})
	}
	if !start.IsZero() && !end.IsZero() {
		metas = append(metas, Meta{Name: "Duration", Value: FormatDuration(end.Sub(start))})
	}
	metas = append(metas,
		Meta{Name: "Turns", Value: fmt.Sprintf("%d", turnCount)},
		Meta{Name: "Context Token", Value: FormatTokens(ctx)},
		Meta{Name: "Output Token", Value: FormatTokens(total.OutputTokens)},
	)
	if total.CacheReadInputTokens > 0 {
		metas = append(metas, Meta{Name: "Cache Read Token", Value: FormatTokens(total.CacheReadInputTokens)})
	}
	if total.CacheCreationInputTokens > 0 {
		metas = append(metas, Meta{Name: "Cache Creation Token", Value: FormatTokens(total.CacheCreationInputTokens)})
	}

	return SessionDetail{
		Title: title,
		Metas: metas,
		Items: items,
	}, nil
}
