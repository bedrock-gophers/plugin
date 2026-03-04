package mixed

import (
	"fmt"
	"strings"

	"github.com/bedrock-gophers/plugin/internal/generator"
	"github.com/bedrock-gophers/plugin/internal/generator/ports"
)

type Adapter struct {
	PortName  string
	PortUsage ports.Usage
	OutDir    string
	Roots     []any
}

func (a Adapter) Name() string {
	return a.PortName
}

func (a Adapter) Language() ports.Language {
	return ports.LanguageMixed
}

func (a Adapter) Usage() ports.Usage {
	return a.PortUsage
}

func (a Adapter) Generate() error {
	if strings.TrimSpace(a.PortName) == "" {
		return fmt.Errorf("generator: mixed adapter name is required")
	}
	if strings.TrimSpace(a.OutDir) == "" {
		return fmt.Errorf("generator: mixed adapter %q out dir is required", a.PortName)
	}
	if len(a.Roots) == 0 {
		return fmt.Errorf("generator: mixed adapter %q requires at least one root", a.PortName)
	}

	for i, root := range a.Roots {
		if err := generateCLibrary(root, a.OutDir); err != nil {
			return fmt.Errorf("interop root[%d]: %w", i, err)
		}
	}
	return nil
}

func generateCLibrary(root any, outDir string) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			switch v := recovered.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("%v", recovered)
			}
		}
	}()

	g := generator.Generator{Type: root}
	g.CLibrary(outDir)
	return nil
}
