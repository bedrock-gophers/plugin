//go:build !cgo

package plugin

import (
	"fmt"

	"github.com/bedrock-gophers/plugin/plugin/abi"
)

type csharpRuntime struct {
	ctxID uintptr
}

func (m *Manager) startCSharpPlugin(plug *pluginRuntime) error {
	return m.startNativePlugin(plug, "csharp")
}

func (m *Manager) startRustPlugin(plug *pluginRuntime) error {
	return m.startNativePlugin(plug, "rust")
}

func (m *Manager) startNativePlugin(plug *pluginRuntime, kind string) error {
	if plug == nil {
		return fmt.Errorf("%s plugin runtime is nil", kind)
	}
	return fmt.Errorf("%s plugins require cgo; rebuild with CGO_ENABLED=1 and a working C toolchain", kind)
}

func (rt *csharpRuntime) dispatch(_ *Manager, _ *pluginRuntime, _ abi.EventDescriptor, _ []byte) error {
	return fmt.Errorf("csharp runtime is unavailable because cgo is disabled")
}

func (rt *csharpRuntime) close() {}

func registerCSharpMutableState(_ uintptr, _ *mutableState) uint64 { return 0 }

func unregisterCSharpMutableState(_ uintptr, _ uint64) {}
