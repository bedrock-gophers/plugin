package ports

type BridgeArgSpec struct {
	Name string
	Type string
}

type BridgeOpSpec struct {
	HostMethod string
	HostReturn string
	HostArg    *BridgeArgSpec

	GuestFunc     string
	GuestReturn   string
	GuestArg      *BridgeArgSpec
	GuestFallback string
	GuestExpr     string

	RefMethod string
	RefReturn string
	RefArg    *BridgeArgSpec
	RefExpr   string

	ManagerBody string
}

type BridgeConfig struct {
	Ops []BridgeOpSpec

	PrimaryArg BridgeArgSpec

	GuestHostOutput    string
	GuestHostPackage   string
	GuestHostImports   []string
	GuestHostInterface string

	RefOutput       string
	RefPackage      string
	RefImports      []string
	RefReceiverType string
	RefReceiverName string

	ManagerOutput  string
	ManagerPackage string
	ManagerImports []string

	RuntimeOutput  string
	RuntimePackage string
}
