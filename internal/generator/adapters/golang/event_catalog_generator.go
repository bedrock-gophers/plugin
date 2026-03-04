package golang

import (
	"bytes"
	"fmt"
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
	if cfg.EventDescriptor <= 0 {
		return fmt.Errorf("generator: EventCatalogConfig.EventDescriptor must be > 0")
	}
	if len(cfg.Flags) == 0 {
		return fmt.Errorf("generator: EventCatalogConfig.Flags is required")
	}
	if err := writeGeneratedGoFile(cfg.Output, renderEventCatalog(cfg)); err != nil {
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
	b.WriteString(generatedBanner)
	b.WriteString("\n")
	b.WriteString("package ")
	b.WriteString(cfg.Package)
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
