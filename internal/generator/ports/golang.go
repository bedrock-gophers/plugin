package ports

const (
	GoOpBridgeSDK     = "bridge-sdk"
	GoOpBridgeRuntime = "bridge-runtime"
	GoOpIdentifiers   = "identifiers"
	GoOpHostCallOps   = "host-call-ops"
	GoOpEventCatalog  = "event-catalog"
	GoOpContextKeys   = "context-keys"
)

type ContextKeysConfig struct {
	KeysPath   string
	Package    string
	OutputPath string
}
