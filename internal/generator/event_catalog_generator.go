package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type EventCatalogEntry struct {
	Name        string
	HandlerName string
}

type EventCatalogConfig struct {
	GoOutput        string
	GoPackage       string
	Version         uint16
	EventDescriptor int
	Flags           []string

	CSharpOutput    string
	CSharpNamespace string
	CSharpClass     string

	Events []EventCatalogEntry
}

func GenerateEventCatalog(cfg EventCatalogConfig) error {
	if err := GenerateEventCatalogGo(cfg); err != nil {
		return err
	}
	if err := GenerateEventCatalogCSharp(cfg); err != nil {
		return err
	}
	return nil
}

func GenerateEventCatalogGo(cfg EventCatalogConfig) error {
	if err := validateEventCatalogEntries(cfg.Events); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.GoOutput) == "" {
		return fmt.Errorf("generator: EventCatalogConfig.GoOutput is required")
	}
	if strings.TrimSpace(cfg.GoPackage) == "" {
		return fmt.Errorf("generator: EventCatalogConfig.GoPackage is required")
	}
	if cfg.EventDescriptor <= 0 {
		return fmt.Errorf("generator: EventCatalogConfig.EventDescriptor must be > 0")
	}
	if len(cfg.Flags) == 0 {
		return fmt.Errorf("generator: EventCatalogConfig.Flags is required")
	}
	if err := writeGeneratedGoFile(cfg.GoOutput, renderEventCatalogGo(cfg)); err != nil {
		return err
	}
	return nil
}

func GenerateEventCatalogCSharp(cfg EventCatalogConfig) error {
	if err := validateEventCatalogEntries(cfg.Events); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.CSharpOutput) == "" {
		return fmt.Errorf("generator: EventCatalogConfig.CSharpOutput is required")
	}
	if strings.TrimSpace(cfg.CSharpNamespace) == "" {
		return fmt.Errorf("generator: EventCatalogConfig.CSharpNamespace is required")
	}
	if strings.TrimSpace(cfg.CSharpClass) == "" {
		return fmt.Errorf("generator: EventCatalogConfig.CSharpClass is required")
	}
	if err := os.MkdirAll(filepath.Dir(cfg.CSharpOutput), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(cfg.CSharpOutput, []byte(renderEventCatalogCSharp(cfg)), 0o644); err != nil {
		return err
	}
	return nil
}

func validateEventCatalogEntries(events []EventCatalogEntry) error {
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

func renderEventCatalogGo(cfg EventCatalogConfig) string {
	var b bytes.Buffer
	b.WriteString(generatedBanner)
	b.WriteString("\n")
	b.WriteString("package ")
	b.WriteString(cfg.GoPackage)
	b.WriteString("\n\n")

	b.WriteString("const (\n")
	fmt.Fprintf(&b, "\tVersion uint16 = %d\n", cfg.Version)
	b.WriteString(")\n\n")

	b.WriteString("const (\n")
	for i, flag := range cfg.Flags {
		if i == 0 {
			fmt.Fprintf(&b, "\t%s uint32 = 1 << iota\n", flag)
			continue
		}
		fmt.Fprintf(&b, "\t%s\n", flag)
	}
	b.WriteString(")\n\n")

	fmt.Fprintf(&b, "const EventDescriptorSize = %d\n\n", cfg.EventDescriptor)

	b.WriteString("const (\n")
	for i, ev := range cfg.Events {
		if i == 0 {
			fmt.Fprintf(&b, "\tEvent%s uint16 = iota + 1\n", ev.Name)
			continue
		}
		fmt.Fprintf(&b, "\tEvent%s\n", ev.Name)
	}
	b.WriteString(")\n\n")

	b.WriteString("func EventName(id uint16) string {\n")
	b.WriteString("\tswitch id {\n")
	for _, ev := range cfg.Events {
		fmt.Fprintf(&b, "\tcase Event%s:\n", ev.Name)
		fmt.Fprintf(&b, "\t\treturn %q\n", ev.HandlerName)
	}
	b.WriteString("\tdefault:\n")
	b.WriteString("\t\treturn \"unknown\"\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n")

	return b.String()
}

func renderEventCatalogCSharp(cfg EventCatalogConfig) string {
	var b bytes.Buffer
	b.WriteString(csharpGeneratedBanner)
	b.WriteString("\n")
	b.WriteString("namespace ")
	b.WriteString(cfg.CSharpNamespace)
	b.WriteString(";\n\n")
	b.WriteString("public static class ")
	b.WriteString(cfg.CSharpClass)
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
