//go:build !cgo

package plugin

import (
	"fmt"

	"github.com/bedrock-gophers/plugin/plugin/abi"
)

type nativeRuntime struct {
	ctxID uintptr
}

func (m *Manager) startNativePlugin(plug *pluginRuntime, kind string) error {
	if plug == nil {
		return fmt.Errorf("%s plugin runtime is nil", kind)
	}
	return fmt.Errorf("%s plugins require cgo; rebuild with CGO_ENABLED=1 and a working C toolchain", kind)
}

func (rt *nativeRuntime) dispatch(_ *Manager, _ *pluginRuntime, _ abi.EventDescriptor, _ []byte) error {
	return fmt.Errorf("native runtime is unavailable because cgo is disabled")
}

func (rt *nativeRuntime) close() {}

func registerNativeMutableState(_ uintptr, _ *mutableState) uint64 { return 0 }

func unregisterNativeMutableState(_ uintptr, _ uint64) {}
