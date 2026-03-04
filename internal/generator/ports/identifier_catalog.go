package ports

import (
	"fmt"
	"strings"
	"unicode"
)

type IdentifierEntry struct {
	Symbol string
	Value  string
}

type IdentifierGroup struct {
	Package string

	TypeName string
	Prefix   string
	AllVar   string
	Partial  bool

	Entries []IdentifierEntry
}

type IdentifierCatalogConfig struct {
	Output  string
	Package string

	Groups []IdentifierGroup
}

func BuildIdentifierEntries(ids []string) []IdentifierEntry {
	out := make([]IdentifierEntry, 0, len(ids))
	used := map[string]int{}
	for _, id := range ids {
		symbol := symbolForIdentifier(id)
		count := used[symbol]
		used[symbol] = count + 1
		if count > 0 {
			symbol = fmt.Sprintf("%s_%d", symbol, count+1)
		}
		out = append(out, IdentifierEntry{Symbol: symbol, Value: id})
	}
	return out
}

func symbolForIdentifier(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	v = strings.TrimPrefix(v, "minecraft:")

	var parts []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		parts = append(parts, cur.String())
		cur.Reset()
	}

	for _, r := range v {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			cur.WriteRune(r)
			continue
		}
		flush()
	}
	flush()

	if len(parts) == 0 {
		return "Unknown"
	}

	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		rs := []rune(p)
		rs[0] = unicode.ToUpper(rs[0])
		b.WriteString(string(rs))
	}
	name := b.String()
	if name == "" {
		name = "Unknown"
	}
	if name[0] >= '0' && name[0] <= '9' {
		name = "N" + name
	}
	return name
}
