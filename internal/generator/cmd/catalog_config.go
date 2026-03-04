package main

import (
	"sort"
	"strings"

	"github.com/bedrock-gophers/plugin/internal/generator"
	_ "github.com/df-mc/dragonfly/server/block"
	_ "github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

func identifierCatalogConfig() generator.IdentifierCatalogConfig {
	itemIDs := collectSortedUniqueIDs(func(add func(string)) {
		for _, it := range world.Items() {
			if it == nil {
				continue
			}
			name, _ := it.EncodeItem()
			add(name)
		}
	})
	blockIDs := collectSortedUniqueIDs(func(add func(string)) {
		for _, b := range world.Blocks() {
			if b == nil {
				continue
			}
			name, _ := b.EncodeBlock()
			add(name)
		}
	})
	worldIDs := []string{"overworld", "nether", "end"}

	return generator.IdentifierCatalogConfig{
		GoOutput:     "../../plugin/sdk/go/ids_generated.go",
		GoPackage:    "guest",
		CSharpOutput: "../../plugin/sdk/csharp/src/Abi/Ids.g.cs",
		Groups: []generator.IdentifierGroup{
			{
				GoType:          "ItemID",
				GoConstPrefix:   "Item",
				GoAllVar:        "AllItemIDs",
				CSharpNamespace: "BedrockPlugin.Sdk.Guest",
				CSharpClass:     "Item",
				CSharpPartial:   true,
				Entries:         generator.BuildIdentifierEntries(itemIDs),
			},
			{
				GoType:          "BlockID",
				GoConstPrefix:   "Block",
				GoAllVar:        "AllBlockIDs",
				CSharpNamespace: "BedrockPlugin.Sdk.Abi",
				CSharpClass:     "BlockIds",
				Entries:         generator.BuildIdentifierEntries(blockIDs),
			},
			{
				GoType:          "WorldID",
				GoConstPrefix:   "World",
				GoAllVar:        "AllWorldIDs",
				CSharpNamespace: "BedrockPlugin.Sdk.Abi",
				CSharpClass:     "WorldIds",
				Entries:         generator.BuildIdentifierEntries(worldIDs),
			},
		},
	}
}

func generateIdentifierCatalogs() error {
	cfg := identifierCatalogConfig()
	if err := generator.GenerateIdentifierCatalogGo(cfg); err != nil {
		return err
	}
	if err := generator.GenerateIdentifierCatalogCSharp(cfg); err != nil {
		return err
	}
	return nil
}

func generateIdentifierCatalogsGo() error {
	return generator.GenerateIdentifierCatalogGo(identifierCatalogConfig())
}

func generateIdentifierCatalogsCSharp() error {
	return generator.GenerateIdentifierCatalogCSharp(identifierCatalogConfig())
}

func collectSortedUniqueIDs(feed func(add func(string))) []string {
	seen := map[string]struct{}{}
	values := make([]string, 0, 1024)
	feed(func(v string) {
		v = strings.TrimSpace(v)
		if v == "" {
			return
		}
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		values = append(values, v)
	})
	sort.Strings(values)
	return values
}
