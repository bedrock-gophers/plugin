package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	csharpadapter "github.com/bedrock-gophers/plugin/internal/generator/adapters/csharp"
	golangadapter "github.com/bedrock-gophers/plugin/internal/generator/adapters/golang"
	mixedadapter "github.com/bedrock-gophers/plugin/internal/generator/adapters/mixed"
	"github.com/bedrock-gophers/plugin/internal/generator/ports"
)

func main() {
	outDir := flag.String("out", "output", "directory for generated Go/C/C# bindings")
	profileValue := flag.String("profile", "complete", "generation profile: sdk|runtime|shared|complete")
	languageValue := flag.String("lang", "all", "language filter: all|go|csharp|mixed")
	flag.Parse()

	roots := rootTypes()
	if len(roots) == 0 {
		log.Print("no roots configured; nothing to generate")
		return
	}

	profile, err := ports.ParseProfile(*profileValue)
	if err != nil {
		log.Fatal(err)
	}
	language, err := ports.ParseLanguage(*languageValue)
	if err != nil {
		log.Fatal(err)
	}

	entries := ports.Select(buildGenerationPorts(*outDir, roots), profile, language)
	if len(entries) == 0 {
		log.Printf("no generation ports selected for profile=%s language=%s", profile.Name(), languageLabel(language))
		return
	}
	if err := ports.Run(entries, profile); err != nil {
		log.Fatal(err)
	}
}

func buildGenerationPorts(outDir string, roots []any) []ports.GenerationPort {
	entries := make([]ports.GenerationPort, 0, 16)

	if len(roots) > 0 {
		entries = append(entries, mixedadapter.Adapter{
			PortName:  "interop",
			PortUsage: ports.UsageShared,
			OutDir:    outDir,
			Roots:     roots,
		})
	}

	for i, cfg := range csharpSDKConfigs() {
		cfg := cfg
		entries = append(entries, csharpadapter.Adapter{
			PortName:  fmt.Sprintf("csharp-sdk:%d", i+1),
			PortUsage: ports.UsageSDK,
			Operation: ports.CSharpOpSDK,
			SDKConfig: &cfg,
		})
	}

	for i, cfg := range bridgeConfigs() {
		cfg := cfg
		entries = append(entries, golangadapter.Adapter{
			PortName:     fmt.Sprintf("bridge-sdk:%d", i+1),
			PortUsage:    ports.UsageSDK,
			Operation:    ports.GoOpBridgeSDK,
			BridgeConfig: &cfg,
		})
		entries = append(entries, golangadapter.Adapter{
			PortName:     fmt.Sprintf("bridge-runtime:%d", i+1),
			PortUsage:    ports.UsageRuntime,
			Operation:    ports.GoOpBridgeRuntime,
			BridgeConfig: &cfg,
		})
	}

	identifierCfg := identifierCatalogConfig()
	entries = append(entries, golangadapter.Adapter{
		PortName:         "identifiers-go",
		PortUsage:        ports.UsageSDK,
		Operation:        ports.GoOpIdentifiers,
		IdentifierConfig: &identifierCfg,
	})
	entries = append(entries, csharpadapter.Adapter{
		PortName:         "identifiers-csharp",
		PortUsage:        ports.UsageSDK,
		Operation:        ports.CSharpOpIdentifiers,
		IdentifierConfig: &identifierCfg,
	})

	hostCallCfg := hostCallOpConfig()
	entries = append(entries, golangadapter.Adapter{
		PortName:       "host-call-go",
		PortUsage:      ports.UsageRuntime,
		Operation:      ports.GoOpHostCallOps,
		HostCallConfig: &hostCallCfg,
	})
	entries = append(entries, csharpadapter.Adapter{
		PortName:       "host-call-csharp",
		PortUsage:      ports.UsageSDK,
		Operation:      ports.CSharpOpHostCallOps,
		HostCallConfig: &hostCallCfg,
	})

	eventCfg := eventCatalogConfig()
	entries = append(entries, golangadapter.Adapter{
		PortName:           "event-catalog-go",
		PortUsage:          ports.UsageRuntime,
		Operation:          ports.GoOpEventCatalog,
		EventCatalogConfig: &eventCfg,
	})
	entries = append(entries, csharpadapter.Adapter{
		PortName:           "event-catalog-csharp",
		PortUsage:          ports.UsageSDK,
		Operation:          ports.CSharpOpEventCatalog,
		EventCatalogConfig: &eventCfg,
	})

	entries = append(entries, golangadapter.Adapter{
		PortName:  "context-keys",
		PortUsage: ports.UsageRuntime,
		Operation: ports.GoOpContextKeys,
		ContextKeysConfigRef: &ports.ContextKeysConfig{
			KeysPath:   filepath.Join("cmd", "ctxkey_keys.txt"),
			Package:    "ctxkey",
			OutputPath: "../../plugin/internal/ctxkey/generated.go",
		},
	})

	return entries
}

func languageLabel(language ports.Language) string {
	if language == "" {
		return "all"
	}
	return string(language)
}
