package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type HostCallOpEntry struct {
	Name string
}

type HostCallOpConfig struct {
	GoOutput  string
	GoPackage string
	GoPrefix  string

	CSharpOutput    string
	CSharpNamespace string
	CSharpClass     string

	Ops []HostCallOpEntry
}

func GenerateHostCallOps(cfg HostCallOpConfig) error {
	if err := GenerateHostCallOpsGo(cfg); err != nil {
		return err
	}
	if err := GenerateHostCallOpsCSharp(cfg); err != nil {
		return err
	}
	return nil
}

func GenerateHostCallOpsGo(cfg HostCallOpConfig) error {
	if err := validateHostCallOps(cfg.Ops); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.GoOutput) == "" {
		return fmt.Errorf("generator: HostCallOpConfig.GoOutput is required")
	}
	if strings.TrimSpace(cfg.GoPackage) == "" {
		return fmt.Errorf("generator: HostCallOpConfig.GoPackage is required")
	}
	if strings.TrimSpace(cfg.GoPrefix) == "" {
		return fmt.Errorf("generator: HostCallOpConfig.GoPrefix is required")
	}
	if err := writeGeneratedGoFile(cfg.GoOutput, renderHostCallOpsGo(cfg)); err != nil {
		return err
	}
	return nil
}

func GenerateHostCallOpsCSharp(cfg HostCallOpConfig) error {
	if err := validateHostCallOps(cfg.Ops); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.CSharpOutput) == "" {
		return fmt.Errorf("generator: HostCallOpConfig.CSharpOutput is required")
	}
	if strings.TrimSpace(cfg.CSharpNamespace) == "" {
		return fmt.Errorf("generator: HostCallOpConfig.CSharpNamespace is required")
	}
	if strings.TrimSpace(cfg.CSharpClass) == "" {
		return fmt.Errorf("generator: HostCallOpConfig.CSharpClass is required")
	}
	if err := os.MkdirAll(filepath.Dir(cfg.CSharpOutput), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(cfg.CSharpOutput, []byte(renderHostCallOpsCSharp(cfg)), 0o644); err != nil {
		return err
	}
	return nil
}

func validateHostCallOps(ops []HostCallOpEntry) error {
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

func renderHostCallOpsGo(cfg HostCallOpConfig) string {
	var b bytes.Buffer
	b.WriteString(generatedBanner)
	b.WriteString("\n")
	b.WriteString("package ")
	b.WriteString(cfg.GoPackage)
	b.WriteString("\n\n")
	b.WriteString("const (\n")
	for i, op := range cfg.Ops {
		b.WriteString("\t")
		b.WriteString(cfg.GoPrefix)
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

func renderHostCallOpsCSharp(cfg HostCallOpConfig) string {
	var b bytes.Buffer
	b.WriteString(csharpGeneratedBanner)
	b.WriteString("\n")
	b.WriteString("namespace ")
	b.WriteString(cfg.CSharpNamespace)
	b.WriteString(";\n\n")
	b.WriteString("public static class ")
	b.WriteString(cfg.CSharpClass)
	b.WriteString("\n{\n")
	for i, op := range cfg.Ops {
		fmt.Fprintf(&b, "    public const uint %s = %d;\n", op.Name, i+1)
	}
	b.WriteString("}\n")
	return b.String()
}
