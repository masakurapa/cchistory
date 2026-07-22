package types

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

	var totalOutput int
	var last Usage
	for _, entry := range items {
		u := entry.Usage()
		if u.InputTokens+u.CacheReadInputTokens+u.CacheCreationInputTokens > 0 {
			last = u
		}
		totalOutput += u.OutputTokens
	}
	total := Usage{
		InputTokens:              last.InputTokens,
		OutputTokens:             totalOutput,
		CacheReadInputTokens:     last.CacheReadInputTokens,
		CacheCreationInputTokens: last.CacheCreationInputTokens,
	}

	title := session.ID
	if session.Name != "" {
		title = session.Name
	}

	ctx := total.InputTokens + total.CacheReadInputTokens + total.CacheCreationInputTokens
	metas := []Meta{
		{Name: "ID", Value: session.ID},
		{Name: "Context Token", Value: FormatTokens(ctx)},
		{Name: "Output Token", Value: FormatTokens(total.OutputTokens)},
	}
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
