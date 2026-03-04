package ports

import (
	"fmt"
	"strings"
)

type Language string

const (
	LanguageGo     Language = "go"
	LanguageCSharp Language = "csharp"
	LanguageMixed  Language = "mixed"
)

type Usage string

const (
	UsageSDK     Usage = "sdk"
	UsageRuntime Usage = "runtime"
	UsageShared  Usage = "shared"
)

type GenerationPort interface {
	Name() string
	Language() Language
	Usage() Usage
	Generate() error
}

type UsageProfile interface {
	Name() string
	Allows(usage Usage) bool
}

func ParseLanguage(value string) (Language, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "all":
		return "", nil
	case "go":
		return LanguageGo, nil
	case "csharp", "cs":
		return LanguageCSharp, nil
	case "mixed":
		return LanguageMixed, nil
	default:
		return "", fmt.Errorf("generator: unsupported language %q", value)
	}
}
