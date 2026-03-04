package golang

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bedrock-gophers/plugin/internal/generator/ports"
)

func generateIdentifierCatalog(cfg ports.IdentifierCatalogConfig) error {
	if strings.TrimSpace(cfg.Output) == "" {
		return fmt.Errorf("generator: IdentifierCatalogConfig.Output is required")
	}
	if strings.TrimSpace(cfg.Package) == "" {
		return fmt.Errorf("generator: IdentifierCatalogConfig.Package is required")
	}
	if err := validateIdentifierGroups(cfg.Groups, true, false); err != nil {
		return err
	}
	if err := writeGeneratedGoFile(cfg.Output, renderIdentifierCatalog(cfg)); err != nil {
		return err
	}
	return nil
}

func validateIdentifierGroups(groups []ports.IdentifierGroup, requireGo bool, requireCSharp bool) error {
	if len(groups) == 0 {
		return fmt.Errorf("generator: IdentifierCatalogConfig.Groups is required")
	}
	for i, g := range groups {
		id := fmt.Sprintf("groups[%d]", i)
		if requireGo {
			if strings.TrimSpace(g.TypeName) == "" {
				return fmt.Errorf("generator: %s TypeName is required", id)
			}
			if strings.TrimSpace(g.Prefix) == "" {
				return fmt.Errorf("generator: %s Prefix is required", id)
			}
			if strings.TrimSpace(g.AllVar) == "" {
				return fmt.Errorf("generator: %s AllVar is required", id)
			}
		}
		if requireCSharp {
			if strings.TrimSpace(g.Package) == "" {
				return fmt.Errorf("generator: %s Package is required", id)
			}
			if strings.TrimSpace(g.TypeName) == "" {
				return fmt.Errorf("generator: %s TypeName is required", id)
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

func renderIdentifierCatalog(cfg ports.IdentifierCatalogConfig) string {
	var b bytes.Buffer
	b.WriteString(generatedBanner)
	b.WriteString("\n")
	b.WriteString("package ")
	b.WriteString(cfg.Package)
	b.WriteString("\n\n")

	for _, g := range cfg.Groups {
		b.WriteString("type ")
		b.WriteString(g.TypeName)
		b.WriteString(" string\n")
	}
	b.WriteString("\n")

	for _, g := range cfg.Groups {
		b.WriteString("const (\n")
		for _, e := range g.Entries {
			fmt.Fprintf(&b, "\t%s%s %s = %q\n", g.Prefix, e.Symbol, g.TypeName, e.Value)
		}
		b.WriteString(")\n\n")
	}

	for i, g := range cfg.Groups {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("var ")
		b.WriteString(g.AllVar)
		b.WriteString(" = []")
		b.WriteString(g.TypeName)
		b.WriteString("{\n")
		for _, e := range g.Entries {
			fmt.Fprintf(&b, "\t%s%s,\n", g.Prefix, e.Symbol)
		}
		b.WriteString("}\n")
	}

	return b.String()
}
