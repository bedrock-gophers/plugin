package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type IdentifierEntry struct {
	Symbol string
	Value  string
}

type IdentifierGroup struct {
	GoType        string
	GoConstPrefix string
	GoAllVar      string

	CSharpNamespace string
	CSharpClass     string
	CSharpPartial   bool

	Entries []IdentifierEntry
}

type IdentifierCatalogConfig struct {
	GoOutput     string
	GoPackage    string
	CSharpOutput string

	Groups []IdentifierGroup
}

func GenerateIdentifierCatalog(cfg IdentifierCatalogConfig) error {
	if err := GenerateIdentifierCatalogGo(cfg); err != nil {
		return err
	}
	if err := GenerateIdentifierCatalogCSharp(cfg); err != nil {
		return err
	}
	return nil
}

func GenerateIdentifierCatalogGo(cfg IdentifierCatalogConfig) error {
	if strings.TrimSpace(cfg.GoOutput) == "" {
		return fmt.Errorf("generator: IdentifierCatalogConfig.GoOutput is required")
	}
	if strings.TrimSpace(cfg.GoPackage) == "" {
		return fmt.Errorf("generator: IdentifierCatalogConfig.GoPackage is required")
	}
	if err := validateIdentifierGroups(cfg.Groups, true, false); err != nil {
		return err
	}
	if err := writeGeneratedGoFile(cfg.GoOutput, renderIdentifierCatalogGo(cfg)); err != nil {
		return err
	}
	return nil
}

func GenerateIdentifierCatalogCSharp(cfg IdentifierCatalogConfig) error {
	if strings.TrimSpace(cfg.CSharpOutput) == "" {
		return fmt.Errorf("generator: IdentifierCatalogConfig.CSharpOutput is required")
	}
	if err := validateIdentifierGroups(cfg.Groups, false, true); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cfg.CSharpOutput), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(cfg.CSharpOutput, []byte(renderIdentifierCatalogCSharp(cfg)), 0o644); err != nil {
		return err
	}
	return nil
}

func validateIdentifierGroups(groups []IdentifierGroup, requireGo bool, requireCSharp bool) error {
	if len(groups) == 0 {
		return fmt.Errorf("generator: IdentifierCatalogConfig.Groups is required")
	}
	for i, g := range groups {
		id := fmt.Sprintf("groups[%d]", i)
		if requireGo {
			if strings.TrimSpace(g.GoType) == "" {
				return fmt.Errorf("generator: %s GoType is required", id)
			}
			if strings.TrimSpace(g.GoConstPrefix) == "" {
				return fmt.Errorf("generator: %s GoConstPrefix is required", id)
			}
			if strings.TrimSpace(g.GoAllVar) == "" {
				return fmt.Errorf("generator: %s GoAllVar is required", id)
			}
		}
		if requireCSharp {
			if strings.TrimSpace(g.CSharpNamespace) == "" {
				return fmt.Errorf("generator: %s CSharpNamespace is required", id)
			}
			if strings.TrimSpace(g.CSharpClass) == "" {
				return fmt.Errorf("generator: %s CSharpClass is required", id)
			}
		}
		if len(g.Entries) == 0 {
			return fmt.Errorf("generator: %s Entries is required", id)
		}

		seenSymbols := map[string]struct{}{}
		seenValues := map[string]struct{}{}
		for j, e := range g.Entries {
			entryID := fmt.Sprintf("%s entries[%d]", id, j)
			symbol := strings.TrimSpace(e.Symbol)
			value := strings.TrimSpace(e.Value)
			if symbol == "" {
				return fmt.Errorf("generator: %s Symbol is required", entryID)
			}
			if value == "" {
				return fmt.Errorf("generator: %s Value is required", entryID)
			}
			if _, ok := seenSymbols[symbol]; ok {
				return fmt.Errorf("generator: %s duplicate Symbol %q", id, symbol)
			}
			if _, ok := seenValues[value]; ok {
				return fmt.Errorf("generator: %s duplicate Value %q", id, value)
			}
			seenSymbols[symbol] = struct{}{}
			seenValues[value] = struct{}{}
		}
	}
	return nil
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

func renderIdentifierCatalogGo(cfg IdentifierCatalogConfig) string {
	var b bytes.Buffer
	b.WriteString(generatedBanner)
	b.WriteString("\n")
	b.WriteString("package ")
	b.WriteString(cfg.GoPackage)
	b.WriteString("\n\n")

	for _, g := range cfg.Groups {
		b.WriteString("type ")
		b.WriteString(g.GoType)
		b.WriteString(" string\n")
	}
	b.WriteString("\n")

	for _, g := range cfg.Groups {
		b.WriteString("const (\n")
		for _, e := range g.Entries {
			fmt.Fprintf(&b, "\t%s%s %s = %q\n", g.GoConstPrefix, e.Symbol, g.GoType, e.Value)
		}
		b.WriteString(")\n\n")
	}

	for i, g := range cfg.Groups {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("var ")
		b.WriteString(g.GoAllVar)
		b.WriteString(" = []")
		b.WriteString(g.GoType)
		b.WriteString("{\n")
		for _, e := range g.Entries {
			fmt.Fprintf(&b, "\t%s%s,\n", g.GoConstPrefix, e.Symbol)
		}
		b.WriteString("}\n")
	}

	return b.String()
}

func renderIdentifierCatalogCSharp(cfg IdentifierCatalogConfig) string {
	var b bytes.Buffer
	b.WriteString(csharpGeneratedBanner)
	b.WriteString("\n")
	for i, g := range cfg.Groups {
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString("namespace ")
		b.WriteString(g.CSharpNamespace)
		b.WriteString("\n{\n")
		b.WriteString("public static ")
		if g.CSharpPartial {
			b.WriteString("partial ")
		}
		b.WriteString("class ")
		b.WriteString(g.CSharpClass)
		b.WriteString("\n{\n")
		for _, e := range g.Entries {
			fmt.Fprintf(&b, "    public const string %s = %q;\n", e.Symbol, e.Value)
		}
		b.WriteString("}\n")
		b.WriteString("}")
	}
	b.WriteString("\n")
	return b.String()
}
