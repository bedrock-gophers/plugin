//go:build !cgo

package plugin

import (
	"fmt"

	"github.com/bedrock-gophers/plugin/plugin/abi"
)

type csharpRuntime struct{}

func (m *Manager) startCSharpPlugin(plug *pluginRuntime) error {
	if plug == nil {
		return fmt.Errorf("csharp plugin runtime is nil")
	}
	return fmt.Errorf("csharp plugins require cgo; rebuild with CGO_ENABLED=1 and a working C toolchain")
}

func (rt *csharpRuntime) dispatch(_ *Manager, _ *pluginRuntime, _ abi.EventDescriptor, _ []byte) error {
	return fmt.Errorf("csharp runtime is unavailable because cgo is disabled")
}

func (rt *csharpRuntime) close() {}
