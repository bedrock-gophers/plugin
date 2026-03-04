package ports

type KeySetEntry struct {
	ConstName string
	Key       string
}

type KeySetConfig struct {
	Package    string
	OutputPath string
	Entries    []KeySetEntry

	ConstType  string
	MapName    string
	LookupFunc string
	StartAt    uint32
}
