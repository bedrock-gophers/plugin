package output

// RegisterExternalObject stores an existing Go object in the generated handle registry
// and returns a stable interop handle.
func RegisterExternalObject(value any) uint64 {
	return uint64(registerObject(value))
}
