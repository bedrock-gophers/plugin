package main

import "github.com/bedrock-gophers/plugin/internal/generator"

func hostCallOpConfig() generator.HostCallOpConfig {
	return generator.HostCallOpConfig{
		GoOutput:        "../../plugin/host_call_ops.go",
		GoPackage:       "plugin",
		GoPrefix:        "hostCall",
		CSharpOutput:    "../../plugin/sdk/csharp/src/Abi/HostCallOp.cs",
		CSharpNamespace: "BedrockPlugin.Sdk.Abi",
		CSharpClass:     "HostCallOp",
		Ops:             hostCallOps(),
	}
}

func generateHostCallOps() error {
	cfg := hostCallOpConfig()
	if err := generator.GenerateHostCallOpsGo(cfg); err != nil {
		return err
	}
	if err := generator.GenerateHostCallOpsCSharp(cfg); err != nil {
		return err
	}
	return nil
}

func generateHostCallOpsGo() error {
	return generator.GenerateHostCallOpsGo(hostCallOpConfig())
}

func generateHostCallOpsCSharp() error {
	return generator.GenerateHostCallOpsCSharp(hostCallOpConfig())
}

func hostCallOps() []generator.HostCallOpEntry {
	return []generator.HostCallOpEntry{
		{Name: "BlockNames"},
		{Name: "ItemNames"},
		{Name: "WorldNames"},
		{Name: "PlayerMainHandItemGet"},
		{Name: "PlayerMainHandItemSet"},
		{Name: "PlayerOffHandItemGet"},
		{Name: "PlayerOffHandItemSet"},
		{Name: "PlayerInventoryItemsGet"},
		{Name: "PlayerInventoryItemsSet"},
		{Name: "PlayerEnderChestItemsGet"},
		{Name: "PlayerEnderChestItemsSet"},
		{Name: "PlayerArmourItemsGet"},
		{Name: "PlayerArmourItemsSet"},
		{Name: "PlayerSetHeldSlot"},
		{Name: "PlayerMoveItemsToInventory"},
		{Name: "PlayerCloseForm"},
		{Name: "PlayerCloseDialogue"},
		{Name: "PlayerSendMenuForm"},
		{Name: "PlayerSendModalForm"},
	}
}
