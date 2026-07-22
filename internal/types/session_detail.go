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

	var total Usage
	for _, entry := range items {
		total = total.Add(entry.Usage())
	}

	title := session.ID
	if session.Name != "" {
		title = session.Name
	}

	metas := []Meta{
		{Name: "ID", Value: session.ID},
		{Name: "Input Token", Value: FormatTokens(total.InputTokens)},
		{Name: "Output Token", Value: FormatTokens(total.OutputTokens)},
	}
	if total.CacheReadInputTokens > 0 {
		metas = append(metas, Meta{Name: "Cache Read", Value: FormatTokens(total.CacheReadInputTokens)})
	}
	if total.CacheCreationInputTokens > 0 {
		metas = append(metas, Meta{Name: "Cache Creation", Value: FormatTokens(total.CacheCreationInputTokens)})
	}

	return SessionDetail{
		Title: title,
		Metas: metas,
		Items: items,
	}, nil
}
