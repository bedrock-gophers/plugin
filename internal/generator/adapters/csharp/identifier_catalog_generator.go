package csharp

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bedrock-gophers/plugin/internal/generator/ports"
)

func generateIdentifierCatalog(cfg ports.IdentifierCatalogConfig) error {
	if strings.TrimSpace(cfg.Output) == "" {
		return fmt.Errorf("generator: IdentifierCatalogConfig.Output is required")
	}
	if err := validateIdentifierGroups(cfg.Groups, false, true); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cfg.Output), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(cfg.Output, []byte(renderIdentifierCatalog(cfg)), 0o644); err != nil {
		return err
	}
	if err := writeGeneratedItemBlockAPI(cfg); err != nil {
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
	b.WriteString(csharpGeneratedBanner)
	b.WriteString("\n")
	for i, g := range cfg.Groups {
		if i > 0 {
			b.WriteString("\n\n")
		}
		b.WriteString("namespace ")
		b.WriteString(g.Package)
		b.WriteString("\n{\n")
		b.WriteString("public static ")
		if g.Partial {
			b.WriteString("partial ")
		}
		b.WriteString("class ")
		b.WriteString(g.TypeName)
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

type itemSymbol struct {
	Symbol string
	Value  string
}

func writeGeneratedItemBlockAPI(cfg ports.IdentifierCatalogConfig) error {
	itemGroup := (*ports.IdentifierGroup)(nil)
	blockGroup := (*ports.IdentifierGroup)(nil)

	for i := range cfg.Groups {
		group := &cfg.Groups[i]
		if strings.EqualFold(group.Package, "BedrockPlugin.Sdk.Guest") && strings.EqualFold(group.TypeName, "ItemIds") {
			itemGroup = group
		}
		if strings.EqualFold(group.TypeName, "BlockIds") {
			blockGroup = group
		}
	}

	if itemGroup == nil && blockGroup == nil {
		return nil
	}

	output := filepath.Join(filepath.Dir(cfg.Output), "ItemBlock.g.cs")
	return os.WriteFile(output, []byte(renderItemBlockAPI(itemGroup, blockGroup)), 0o644)
}

func renderItemBlockAPI(itemGroup, blockGroup *ports.IdentifierGroup) string {
	var b bytes.Buffer
	b.WriteString(csharpGeneratedBanner)
	b.WriteString("\n")
	b.WriteString("#nullable enable\n")
	b.WriteString("\n")
	b.WriteString("using BedrockPlugin.Sdk.Abi;\n")
	b.WriteString("using System.Collections.Generic;\n\n")
	b.WriteString("namespace BedrockPlugin.Sdk.Guest;\n\n")
	b.WriteString("public readonly record struct ItemSpec\n")
	b.WriteString("{\n")
	b.WriteString("    public string Name { get; init; }\n")
	b.WriteString("    public int Count { get; init; }\n")
	b.WriteString("    public int Meta { get; init; }\n")
	b.WriteString("    public string CustomName { get; init; }\n\n")
	b.WriteString("    public static implicit operator ItemStackData(ItemSpec value)\n")
	b.WriteString("    {\n")
	b.WriteString("        var count = value.Count <= 0 ? 1 : value.Count;\n")
	b.WriteString("        return Item.NewStack(value.Name ?? string.Empty, count, value.Meta, value.CustomName ?? string.Empty);\n")
	b.WriteString("    }\n")
	b.WriteString("}\n\n")
	b.WriteString("public readonly record struct BlockSpec\n")
	b.WriteString("{\n")
	b.WriteString("    private static readonly IReadOnlyDictionary<string, string> EmptyProperties = new Dictionary<string, string>();\n\n")
	b.WriteString("    public string Name { get; init; }\n")
	b.WriteString("    public IReadOnlyDictionary<string, string>? Properties { get; init; }\n\n")
	b.WriteString("    public static implicit operator BlockData(BlockSpec value)\n")
	b.WriteString("    {\n")
	b.WriteString("        var name = value.Name ?? string.Empty;\n")
	b.WriteString("        if (name.Length == 0)\n")
	b.WriteString("        {\n")
	b.WriteString("            return new BlockData(false, string.Empty, EmptyProperties);\n")
	b.WriteString("        }\n")
	b.WriteString("        return new BlockData(true, name, value.Properties ?? EmptyProperties);\n")
	b.WriteString("    }\n")
	b.WriteString("}\n\n")

	if itemGroup != nil {
		renderItemAPI(&b, *itemGroup)
		b.WriteString("\n")
	}
	if blockGroup != nil {
		renderBlockAPI(&b, *blockGroup)
	}
	return b.String()
}

func renderItemAPI(b *bytes.Buffer, group ports.IdentifierGroup) {
	items := make([]itemSymbol, 0, len(group.Entries))
	bySymbol := make(map[string]string, len(group.Entries))
	for _, entry := range group.Entries {
		items = append(items, itemSymbol{Symbol: entry.Symbol, Value: entry.Value})
		bySymbol[entry.Symbol] = entry.Value
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Symbol < items[j].Symbol
	})

	// Cooked variants: CookedBeef -> Beef{Cooked=true}
	cookedByBase := map[string]string{}
	for _, item := range items {
		if strings.HasPrefix(item.Symbol, "Cooked") {
			base := strings.TrimPrefix(item.Symbol, "Cooked")
			if base != "" {
				if _, ok := bySymbol[base]; ok {
					cookedByBase[base] = item.Symbol
				}
			}
		}
	}

	b.WriteString("public static partial class Item\n")
	b.WriteString("{\n")
	b.WriteString("    public static ItemSpec Of(string itemName, int count = 1, int itemMeta = 0, string customName = \"\")\n")
	b.WriteString("    {\n")
	b.WriteString("        return new ItemSpec\n")
	b.WriteString("        {\n")
	b.WriteString("            Name = itemName ?? string.Empty,\n")
	b.WriteString("            Count = count,\n")
	b.WriteString("            Meta = itemMeta,\n")
	b.WriteString("            CustomName = customName ?? string.Empty,\n")
	b.WriteString("        };\n")
	b.WriteString("    }\n\n")

	for _, item := range items {
		if cookedSymbol, ok := cookedByBase[item.Symbol]; ok {
			fmt.Fprintf(b, "    public static %sVariant %s => new();\n\n", item.Symbol, item.Symbol)
			fmt.Fprintf(b, "    public readonly record struct %sVariant\n", item.Symbol)
			b.WriteString("    {\n")
			b.WriteString("        public bool Cooked { get; init; }\n")
			b.WriteString("        public int Count { get; init; }\n")
			b.WriteString("        public int Meta { get; init; }\n")
			b.WriteString("        public string CustomName { get; init; }\n\n")
			b.WriteString("        public static implicit operator ItemStackData(")
			b.WriteString(item.Symbol)
			b.WriteString("Variant value)\n")
			b.WriteString("        {\n")
			b.WriteString("            var id = value.Cooked ? ItemIds.")
			b.WriteString(cookedSymbol)
			b.WriteString(" : ItemIds.")
			b.WriteString(item.Symbol)
			b.WriteString(";\n")
			b.WriteString("            var count = value.Count <= 0 ? 1 : value.Count;\n")
			b.WriteString("            return Item.NewStack(id, count, value.Meta, value.CustomName ?? string.Empty);\n")
			b.WriteString("        }\n")
			b.WriteString("    }\n\n")
			continue
		}

		fmt.Fprintf(b, "    public static ItemSpec %s => new ItemSpec\n", item.Symbol)
		b.WriteString("    {\n")
		b.WriteString("        Name = ItemIds.")
		b.WriteString(item.Symbol)
		b.WriteString(",\n")
		b.WriteString("        Count = 1,\n")
		b.WriteString("        Meta = 0,\n")
		b.WriteString("        CustomName = string.Empty,\n")
		b.WriteString("    };\n\n")
	}

	b.WriteString("}\n")
}

func renderBlockAPI(b *bytes.Buffer, group ports.IdentifierGroup) {
	blocks := make([]itemSymbol, 0, len(group.Entries))
	for _, entry := range group.Entries {
		blocks = append(blocks, itemSymbol{Symbol: entry.Symbol, Value: entry.Value})
	}
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Symbol < blocks[j].Symbol
	})

	b.WriteString("public static class Block\n")
	b.WriteString("{\n")
	b.WriteString("    private static readonly IReadOnlyDictionary<string, string> EmptyProperties = new Dictionary<string, string>();\n\n")
	b.WriteString("    public static BlockSpec Of(string blockName, IReadOnlyDictionary<string, string>? properties = null)\n")
	b.WriteString("    {\n")
	b.WriteString("        return new BlockSpec\n")
	b.WriteString("        {\n")
	b.WriteString("            Name = blockName ?? string.Empty,\n")
	b.WriteString("            Properties = properties ?? EmptyProperties,\n")
	b.WriteString("        };\n")
	b.WriteString("    }\n\n")

	for _, block := range blocks {
		fmt.Fprintf(b, "    public static BlockSpec %s => new BlockSpec\n", block.Symbol)
		b.WriteString("    {\n")
		b.WriteString("        Name = BlockIds.")
		b.WriteString(block.Symbol)
		b.WriteString(",\n")
		b.WriteString("        Properties = EmptyProperties,\n")
		b.WriteString("    };\n\n")
	}
	b.WriteString("}\n")
}
