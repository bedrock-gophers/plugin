package csharp

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bedrock-gophers/plugin/internal/generator/ports"
)

func generateEventCatalog(cfg ports.EventCatalogConfig) error {
	if err := validateEventCatalogEntries(cfg.Events); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Output) == "" {
		return fmt.Errorf("generator: EventCatalogConfig.Output is required")
	}
	if strings.TrimSpace(cfg.Package) == "" {
		return fmt.Errorf("generator: EventCatalogConfig.Package is required")
	}
	if strings.TrimSpace(cfg.TypeName) == "" {
		return fmt.Errorf("generator: EventCatalogConfig.TypeName is required")
	}
	if err := os.MkdirAll(filepath.Dir(cfg.Output), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(cfg.Output, []byte(renderEventCatalog(cfg)), 0o644); err != nil {
		return err
	}
	return nil
}

func validateEventCatalogEntries(events []ports.EventCatalogEntry) error {
	if len(events) == 0 {
		return fmt.Errorf("generator: EventCatalogConfig.Events is required")
	}
	seenNames := map[string]struct{}{}
	for i, ev := range events {
		id := fmt.Sprintf("events[%d]", i)
		if strings.TrimSpace(ev.Name) == "" {
			return fmt.Errorf("generator: %s Name is required", id)
		}
		if strings.TrimSpace(ev.HandlerName) == "" {
			return fmt.Errorf("generator: %s HandlerName is required", id)
		}
		if _, ok := seenNames[ev.Name]; ok {
			return fmt.Errorf("generator: duplicate event name %q", ev.Name)
		}
		seenNames[ev.Name] = struct{}{}
	}
	return nil
}

func renderEventCatalog(cfg ports.EventCatalogConfig) string {
	var b bytes.Buffer
	b.WriteString(csharpGeneratedBanner)
	b.WriteString("\n")
	b.WriteString("namespace ")
	b.WriteString(cfg.Package)
	b.WriteString(";\n\n")
	b.WriteString("public static class ")
	b.WriteString(cfg.TypeName)
	b.WriteString("\n{\n")
	for i, ev := range cfg.Events {
		fmt.Fprintf(&b, "    public const ushort Event%s = %d;\n", ev.Name, i+1)
	}
	b.WriteString("\n")
	b.WriteString("    public static string EventName(ushort id) => id switch\n")
	b.WriteString("    {\n")
	for _, ev := range cfg.Events {
		fmt.Fprintf(&b, "        Event%s => %q,\n", ev.Name, ev.HandlerName)
	}
	b.WriteString("        _ => \"unknown\",\n")
	b.WriteString("    };\n")
	b.WriteString("}\n")
	return b.String()
}
