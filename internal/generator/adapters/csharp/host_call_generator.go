package csharp

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
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
	if strings.TrimSpace(cfg.TypeName) == "" {
		return fmt.Errorf("generator: HostCallOpConfig.TypeName is required")
	}
	if err := os.MkdirAll(filepath.Dir(cfg.Output), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(cfg.Output, []byte(renderHostCallOps(cfg)), 0o644); err != nil {
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
	b.WriteString(csharpGeneratedBanner)
	b.WriteString("\n")
	b.WriteString("namespace ")
	b.WriteString(cfg.Package)
	b.WriteString(";\n\n")
	b.WriteString("public static class ")
	b.WriteString(cfg.TypeName)
	b.WriteString("\n{\n")
	for i, op := range cfg.Ops {
		fmt.Fprintf(&b, "    public const uint %s = %d;\n", op.Name, i+1)
	}
	b.WriteString("}\n")
	return b.String()
}
