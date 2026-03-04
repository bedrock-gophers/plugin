package main

import "github.com/bedrock-gophers/plugin/internal/generator"

func eventCatalogConfig() generator.EventCatalogConfig {
	return generator.EventCatalogConfig{
		GoOutput:        "../../plugin/abi/events.go",
		GoPackage:       "abi",
		Version:         1,
		EventDescriptor: 24,
		Flags: []string{
			"FlagCancellable",
			"FlagSynchronous",
		},
		CSharpOutput:    "../../plugin/sdk/csharp/src/Abi/EventIds.cs",
		CSharpNamespace: "BedrockPlugin.Sdk.Abi",
		CSharpClass:     "EventIds",
		Events:          eventCatalogEntries(),
	}
}

func generateEventCatalog() error {
	cfg := eventCatalogConfig()
	if err := generator.GenerateEventCatalogGo(cfg); err != nil {
		return err
	}
	if err := generator.GenerateEventCatalogCSharp(cfg); err != nil {
		return err
	}
	return nil
}

func generateEventCatalogGo() error {
	return generator.GenerateEventCatalogGo(eventCatalogConfig())
}

func generateEventCatalogCSharp() error {
	return generator.GenerateEventCatalogCSharp(eventCatalogConfig())
}

func eventCatalogEntries() []generator.EventCatalogEntry {
	return []generator.EventCatalogEntry{
		{Name: "Move", HandlerName: "HandleMove"},
		{Name: "Jump", HandlerName: "HandleJump"},
		{Name: "Teleport", HandlerName: "HandleTeleport"},
		{Name: "ChangeWorld", HandlerName: "HandleChangeWorld"},
		{Name: "ToggleSprint", HandlerName: "HandleToggleSprint"},
		{Name: "ToggleSneak", HandlerName: "HandleToggleSneak"},
		{Name: "Chat", HandlerName: "HandleChat"},
		{Name: "FoodLoss", HandlerName: "HandleFoodLoss"},
		{Name: "Heal", HandlerName: "HandleHeal"},
		{Name: "Hurt", HandlerName: "HandleHurt"},
		{Name: "Death", HandlerName: "HandleDeath"},
		{Name: "Respawn", HandlerName: "HandleRespawn"},
		{Name: "SkinChange", HandlerName: "HandleSkinChange"},
		{Name: "FireExtinguish", HandlerName: "HandleFireExtinguish"},
		{Name: "StartBreak", HandlerName: "HandleStartBreak"},
		{Name: "BlockBreak", HandlerName: "HandleBlockBreak"},
		{Name: "BlockPlace", HandlerName: "HandleBlockPlace"},
		{Name: "BlockPick", HandlerName: "HandleBlockPick"},
		{Name: "ItemUse", HandlerName: "HandleItemUse"},
		{Name: "ItemUseOnBlock", HandlerName: "HandleItemUseOnBlock"},
		{Name: "ItemUseOnEntity", HandlerName: "HandleItemUseOnEntity"},
		{Name: "ItemRelease", HandlerName: "HandleItemRelease"},
		{Name: "ItemConsume", HandlerName: "HandleItemConsume"},
		{Name: "AttackEntity", HandlerName: "HandleAttackEntity"},
		{Name: "ExperienceGain", HandlerName: "HandleExperienceGain"},
		{Name: "PunchAir", HandlerName: "HandlePunchAir"},
		{Name: "SignEdit", HandlerName: "HandleSignEdit"},
		{Name: "Sleep", HandlerName: "HandleSleep"},
		{Name: "LecternPageTurn", HandlerName: "HandleLecternPageTurn"},
		{Name: "ItemDamage", HandlerName: "HandleItemDamage"},
		{Name: "ItemPickup", HandlerName: "HandleItemPickup"},
		{Name: "HeldSlotChange", HandlerName: "HandleHeldSlotChange"},
		{Name: "ItemDrop", HandlerName: "HandleItemDrop"},
		{Name: "Transfer", HandlerName: "HandleTransfer"},
		{Name: "CommandExecution", HandlerName: "HandleCommandExecution"},
		{Name: "Quit", HandlerName: "HandleQuit"},
		{Name: "Diagnostics", HandlerName: "HandleDiagnostics"},
		{Name: "Join", HandlerName: "HandleJoin"},
		{Name: "PluginCommand", HandlerName: "HandlePluginCommand"},
	}
}
