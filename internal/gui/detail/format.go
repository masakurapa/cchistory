package detail

import (
	"fmt"
	"strings"

	"github.com/masakurapa/cchist/internal/types"
)

func formatTokens(n int) string {
	s := fmt.Sprintf("%d", n)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}

func formatUsage(u types.Usage) string {
	return fmt.Sprintf(
		"Input Token: %s  Output Token: %s\nCache Read Token: %s  Cache Creation Token: %s",
		formatTokens(u.InputTokens),
		formatTokens(u.OutputTokens),
		formatTokens(u.CacheReadInputTokens),
		formatTokens(u.CacheCreationInputTokens),
	)
}
