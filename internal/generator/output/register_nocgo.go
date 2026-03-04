//go:build !cgo

package output

import (
	"sync"
	"sync/atomic"
)

var (
	nocgoHandleCounter atomic.Uint64
	nocgoHandleStore   sync.Map
)

func registerObject(value any) uint64 {
	if value == nil {
		return 0
	}
	handle := nocgoHandleCounter.Add(1)
	nocgoHandleStore.Store(handle, value)
	return handle
}
