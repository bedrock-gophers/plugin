package ports

import (
	"fmt"
	"strings"
)

type SDKProfile struct{}

func (SDKProfile) Name() string { return "sdk" }
func (SDKProfile) Allows(usage Usage) bool {
	return usage == UsageSDK
}

type RuntimeProfile struct{}

func (RuntimeProfile) Name() string { return "runtime" }
func (RuntimeProfile) Allows(usage Usage) bool {
	return usage == UsageRuntime
}

type SharedProfile struct{}

func (SharedProfile) Name() string { return "shared" }
func (SharedProfile) Allows(usage Usage) bool {
	return usage == UsageShared
}

type CompleteProfile struct{}

func (CompleteProfile) Name() string { return "complete" }
func (CompleteProfile) Allows(Usage) bool {
	return true
}

func ParseProfile(value string) (UsageProfile, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "complete", "all":
		return CompleteProfile{}, nil
	case "sdk":
		return SDKProfile{}, nil
	case "runtime":
		return RuntimeProfile{}, nil
	case "shared":
		return SharedProfile{}, nil
	default:
		return nil, fmt.Errorf("generator: unsupported profile %q", value)
	}
}
