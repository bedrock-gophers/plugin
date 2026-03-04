package csharp

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

	SDKConfig          *generator.CSharpSDKConfig
	IdentifierConfig   *generator.IdentifierCatalogConfig
	HostCallConfig     *generator.HostCallOpConfig
	EventCatalogConfig *generator.EventCatalogConfig
}

func (a Adapter) Name() string {
	return a.PortName
}

func (a Adapter) Language() ports.Language {
	return ports.LanguageCSharp
}

func (a Adapter) Usage() ports.Usage {
	return a.PortUsage
}

func (a Adapter) Generate() error {
	if strings.TrimSpace(a.PortName) == "" {
		return fmt.Errorf("generator: csharp adapter name is required")
	}
	switch a.Operation {
	case ports.CSharpOpSDK:
		return a.generateSDK()
	case ports.CSharpOpIdentifiers:
		return a.generateIdentifiers()
	case ports.CSharpOpHostCallOps:
		return a.generateHostCallOps()
	case ports.CSharpOpEventCatalog:
		return a.generateEventCatalog()
	default:
		return fmt.Errorf("generator: csharp adapter %q has unsupported operation %q", a.PortName, a.Operation)
	}
}

func (a Adapter) generateSDK() error {
	if a.SDKConfig == nil {
		return fmt.Errorf("generator: csharp adapter %q sdk config is nil", a.PortName)
	}
	return generator.GenerateCSharpSDK(*a.SDKConfig)
}

func (a Adapter) generateIdentifiers() error {
	if a.IdentifierConfig == nil {
		return fmt.Errorf("generator: csharp adapter %q identifier config is nil", a.PortName)
	}
	return generator.GenerateIdentifierCatalogCSharp(*a.IdentifierConfig)
}

func (a Adapter) generateHostCallOps() error {
	if a.HostCallConfig == nil {
		return fmt.Errorf("generator: csharp adapter %q host call config is nil", a.PortName)
	}
	return generator.GenerateHostCallOpsCSharp(*a.HostCallConfig)
}

func (a Adapter) generateEventCatalog() error {
	if a.EventCatalogConfig == nil {
		return fmt.Errorf("generator: csharp adapter %q event catalog config is nil", a.PortName)
	}
	return generator.GenerateEventCatalogCSharp(*a.EventCatalogConfig)
}
