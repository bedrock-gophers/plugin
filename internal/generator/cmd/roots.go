package main

import (
	"github.com/bedrock-gophers/plugin/internal/generator"
	"github.com/df-mc/dragonfly/server/player"
)

// rootTypes is the only project-specific generation hook.
// Keep all generator behaviour generic outside of this file.
func rootTypes() []any {
	return []any{
		(*player.Player)(nil),
	}
}

// csharpSDKConfigs defines project-specific C# SDK generation roots.
// The generator implementation itself remains generic in internal/generator.
func csharpSDKConfigs() []generator.CSharpSDKConfig {
	return []generator.CSharpSDKConfig{
		{
			PackageDir:    "../../plugin/sdk/go",
			Namespace:     "BedrockPlugin.Sdk.Guest",
			OutputDir:     "../../plugin/sdk/csharp/src/Guest",
			HostInterface: "Host",
			ExtraModelTypes: []string{
				"Vec3",
				"Rotation",
				"Pos",
				"WorldData",
				"BlockData",
				"EntityData",
				"DamageSourceData",
				"HealingSourceData",
				"SkinData",
				"CommandData",
				"DiagnosticsData",
				"PlayerIdentity",
			},
		},
	}
}
