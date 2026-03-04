package golang

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bedrock-gophers/plugin/internal/generator/ports"
)

func generateHostCallOps(cfg ports.HostCallOpConfig) error {
	if err := validateHostCallOps(cfg.Ops); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Output) == "" {
		return fmt.Errorf("generator: HostCallOpConfig.Output is required")
	}
	if strings.TrimSpace(cfg.Package) == "" {
		return fmt.Errorf("generator: HostCallOpConfig.Package is required")
	}
	if strings.TrimSpace(cfg.Prefix) == "" {
		return fmt.Errorf("generator: HostCallOpConfig.Prefix is required")
	}
	if err := writeGeneratedGoFile(cfg.Output, renderHostCallOps(cfg)); err != nil {
		return err
	}
	return nil
}

func validateHostCallOps(ops []ports.HostCallOpEntry) error {
	if len(ops) == 0 {
		return fmt.Errorf("generator: HostCallOpConfig.Ops is required")
	}
	seen := map[string]struct{}{}
	for i, op := range ops {
		name := strings.TrimSpace(op.Name)
		if name == "" {
			return fmt.Errorf("generator: ops[%d].Name is required", i)
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("generator: duplicate op name %q", name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

func renderHostCallOps(cfg ports.HostCallOpConfig) string {
	var b bytes.Buffer
	b.WriteString(generatedBanner)
	b.WriteString("\n")
	b.WriteString("package ")
	b.WriteString(cfg.Package)
	b.WriteString("\n\n")
	b.WriteString("const (\n")
	for i, op := range cfg.Ops {
		b.WriteString("\t")
		b.WriteString(cfg.Prefix)
		b.WriteString(op.Name)
		if i == 0 {
			b.WriteString(" uint32 = iota + 1\n")
			continue
		}
		b.WriteString("\n")
	}
	b.WriteString(")\n")
	return b.String()
}
