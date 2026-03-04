package ports

type EventCatalogEntry struct {
	Name        string
	HandlerName string
}

type EventCatalogConfig struct {
	Output          string
	Package         string
	TypeName        string
	Version         uint16
	EventDescriptor int
	Flags           []string

	Events []EventCatalogEntry
}
