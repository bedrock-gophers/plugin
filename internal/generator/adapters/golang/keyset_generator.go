package golang

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bedrock-gophers/plugin/internal/generator/ports"
)

func readKeySetEntries(path string) ([]ports.KeySetEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	var out []ports.KeySetEntry
	lineNo := 0
	seenConst := map[string]struct{}{}
	seenKey := map[string]struct{}{}
	for s.Scan() {
		lineNo++
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d: expected '<ConstName> <key>'", lineNo)
		}
		e := ports.KeySetEntry{ConstName: parts[0], Key: parts[1]}
		if _, ok := seenConst[e.ConstName]; ok {
			return nil, fmt.Errorf("line %d: duplicate const %q", lineNo, e.ConstName)
		}
		if _, ok := seenKey[e.Key]; ok {
			return nil, fmt.Errorf("line %d: duplicate key %q", lineNo, e.Key)
		}
		seenConst[e.ConstName] = struct{}{}
		seenKey[e.Key] = struct{}{}
		out = append(out, e)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func generateKeySet(cfg ports.KeySetConfig) error {
	if strings.TrimSpace(cfg.Package) == "" {
		return fmt.Errorf("generator: KeySetConfig.Package is required")
	}
	if strings.TrimSpace(cfg.OutputPath) == "" {
		return fmt.Errorf("generator: KeySetConfig.OutputPath is required")
	}
	if len(cfg.Entries) == 0 {
		return fmt.Errorf("generator: KeySetConfig.Entries is required")
	}

	constType := strings.TrimSpace(cfg.ConstType)
	if constType == "" {
		constType = "uint32"
	}
	mapName := strings.TrimSpace(cfg.MapName)
	if mapName == "" {
		mapName = "byName"
	}
	lookup := strings.TrimSpace(cfg.LookupFunc)
	if lookup == "" {
		lookup = "ID"
	}
	start := cfg.StartAt
	if start == 0 {
		start = 1
	}

	seenConst := map[string]struct{}{}
	seenKey := map[string]struct{}{}
	for i, e := range cfg.Entries {
		id := fmt.Sprintf("entries[%d]", i)
		if strings.TrimSpace(e.ConstName) == "" {
			return fmt.Errorf("generator: %s ConstName is required", id)
		}
		if strings.TrimSpace(e.Key) == "" {
			return fmt.Errorf("generator: %s Key is required", id)
		}
		if _, ok := seenConst[e.ConstName]; ok {
			return fmt.Errorf("generator: duplicate const %q", e.ConstName)
		}
		if _, ok := seenKey[e.Key]; ok {
			return fmt.Errorf("generator: duplicate key %q", e.Key)
		}
		seenConst[e.ConstName] = struct{}{}
		seenKey[e.Key] = struct{}{}
	}

	var b bytes.Buffer
	b.WriteString(generatedBanner)
	b.WriteString("\n")
	b.WriteString("package ")
	b.WriteString(cfg.Package)
	b.WriteString("\n\n")

	b.WriteString("const (\n")
	for i, e := range cfg.Entries {
		id := uint64(start) + uint64(i)
		b.WriteString("\t")
		b.WriteString(e.ConstName)
		b.WriteString(" ")
		b.WriteString(constType)
		b.WriteString(" = ")
		b.WriteString(strconv.FormatUint(id, 10))
		b.WriteString("\n")
	}
	b.WriteString(")\n\n")

	b.WriteString("var ")
	b.WriteString(mapName)
	b.WriteString(" = map[string]")
	b.WriteString(constType)
	b.WriteString("{\n")
	for _, e := range cfg.Entries {
		b.WriteString("\t\"")
		b.WriteString(e.Key)
		b.WriteString("\": ")
		b.WriteString(e.ConstName)
		b.WriteString(",\n")
	}
	b.WriteString("}\n\n")

	b.WriteString("func ")
	b.WriteString(lookup)
	b.WriteString("(name string) (")
	b.WriteString(constType)
	b.WriteString(", bool) {\n")
	b.WriteString("\tid, ok := ")
	b.WriteString(mapName)
	b.WriteString("[name]\n")
	b.WriteString("\treturn id, ok\n")
	b.WriteString("}\n")

	return writeGeneratedGoFile(cfg.OutputPath, b.String())
}
