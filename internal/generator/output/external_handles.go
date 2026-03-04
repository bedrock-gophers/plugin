package output

// RegisterExternalObject publishes an existing Go object in the interop handle store.
func RegisterExternalObject(value any) uint64 {
	return uint64(registerObject(value))
}
