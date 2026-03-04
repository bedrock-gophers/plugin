package ports

type HostCallOpEntry struct {
	Name string
}

type HostCallOpConfig struct {
	Output   string
	Package  string
	TypeName string
	Prefix   string

	Ops []HostCallOpEntry
}
