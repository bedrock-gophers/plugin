package golang

import (
	"fmt"
	"strings"

	"github.com/bedrock-gophers/plugin/internal/generator"
	"github.com/bedrock-gophers/plugin/internal/generator/ports"
)

type Adapter struct {
	PortName  string
	PortUsage ports.Usage
	Operation string

	BridgeConfig         *ports.BridgeConfig
	IdentifierConfig     *generator.IdentifierCatalogConfig
	HostCallConfig       *generator.HostCallOpConfig
	EventCatalogConfig   *generator.EventCatalogConfig
	ContextKeysConfigRef *ports.ContextKeysConfig
}

func (a Adapter) Name() string {
	return a.PortName
}

func (a Adapter) Language() ports.Language {
	return ports.LanguageGo
}

func (a Adapter) Usage() ports.Usage {
	return a.PortUsage
}

func (a Adapter) Generate() error {
	if strings.TrimSpace(a.PortName) == "" {
		return fmt.Errorf("generator: go adapter name is required")
	}
	switch a.Operation {
	case ports.GoOpBridgeSDK:
		return a.generateBridgeSDK()
	case ports.GoOpBridgeRuntime:
		return a.generateBridgeRuntime()
	case ports.GoOpIdentifiers:
		return a.generateIdentifiers()
	case ports.GoOpHostCallOps:
		return a.generateHostCallOps()
	case ports.GoOpEventCatalog:
		return a.generateEventCatalog()
	case ports.GoOpContextKeys:
		return a.generateContextKeys()
	default:
		return fmt.Errorf("generator: go adapter %q has unsupported operation %q", a.PortName, a.Operation)
	}
}

func (a Adapter) generateBridgeSDK() error {
	if a.BridgeConfig == nil {
		return fmt.Errorf("generator: go adapter %q bridge config is nil", a.PortName)
	}
	return generateBridgeSDK(*a.BridgeConfig)
}

func (a Adapter) generateBridgeRuntime() error {
	if a.BridgeConfig == nil {
		return fmt.Errorf("generator: go adapter %q bridge config is nil", a.PortName)
	}
	return generateBridgeRuntime(*a.BridgeConfig)
}

func (a Adapter) generateIdentifiers() error {
	if a.IdentifierConfig == nil {
		return fmt.Errorf("generator: go adapter %q identifier config is nil", a.PortName)
	}
	return generator.GenerateIdentifierCatalogGo(*a.IdentifierConfig)
}

func (a Adapter) generateHostCallOps() error {
	if a.HostCallConfig == nil {
		return fmt.Errorf("generator: go adapter %q host call config is nil", a.PortName)
	}
	return generator.GenerateHostCallOpsGo(*a.HostCallConfig)
}

func (a Adapter) generateEventCatalog() error {
	if a.EventCatalogConfig == nil {
		return fmt.Errorf("generator: go adapter %q event catalog config is nil", a.PortName)
	}
	return generator.GenerateEventCatalogGo(*a.EventCatalogConfig)
}

func (a Adapter) generateContextKeys() error {
	if a.ContextKeysConfigRef == nil {
		return fmt.Errorf("generator: go adapter %q context keys config is nil", a.PortName)
	}
	cfg := a.ContextKeysConfigRef
	entries, err := readKeySetEntries(cfg.KeysPath)
	if err != nil {
		return err
	}
	return generateKeySet(ports.KeySetConfig{
		Package:    cfg.Package,
		OutputPath: cfg.OutputPath,
		Entries:    entries,
	})
}
