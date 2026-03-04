package main

import "github.com/bedrock-gophers/plugin/internal/generator/ports"

func bridgeConfigs() []ports.BridgeConfig {
	return []ports.BridgeConfig{
		{
			PrimaryArg:         ports.BridgeArgSpec{Name: "playerID", Type: "uint64"},
			GuestHostInterface: "playerHost",
			RefReceiverType:    "PlayerRef",
			RefReceiverName:    "p",
			GuestHostOutput:    "../../plugin/sdk/go/bridge_host_generated.go",
			GuestHostPackage:   "guest",
			GuestHostImports:   []string{"\"time\""},
			RefOutput:          "../../plugin/sdk/go/bridge_ref_generated.go",
			RefPackage:         "guest",
			RefImports:         []string{"\"time\""},
			ManagerOutput:      "../../plugin/manager_bridge_generated.go",
			ManagerPackage:     "plugin",
			ManagerImports: []string{
				"\"github.com/df-mc/dragonfly/server/player\"",
				"\"strings\"",
			},
			RuntimeOutput:  "../../plugin/csharp_runtime_bridge_generated.go",
			RuntimePackage: "plugin",
			Ops:            bridgeOps(),
		},
	}
}

func bridgeOps() []ports.BridgeOpSpec {
	return []ports.BridgeOpSpec{
		getter("Health", "float64", "float64(0)", "p.Health()"),
		withoutManager(setter("Health", "float64", "health", "p.SetHealth(health)")),

		getter("Food", "int32", "int32(0)", "int32(p.Food())"),
		setter("Food", "int32", "food", "p.SetFood(int(food))"),
		{
			HostMethod:    "PlayerName",
			HostReturn:    "string",
			GuestFunc:     "playerName",
			GuestReturn:   "string",
			GuestFallback: `""`,
			GuestExpr:     "h.PlayerName(playerID)",
			RefMethod:     "Name",
			RefReturn:     "string",
			RefExpr:       "playerName(p.id)",
			ManagerBody: "if name := m.commandTargetNameByID(playerID); name != \"\" {\n" +
				"return name\n" +
				"}\n" +
				"return playerValue(m, playerID, \"\", func(p *player.Player) string {\n" +
				"name := p.Name()\n" +
				"if name == \"\" || strings.EqualFold(name, \"default\") {\n" +
				"return \"\"\n" +
				"}\n" +
				"return name\n" +
				"})",
		},

		{
			HostMethod:    "PlayerGameMode",
			HostReturn:    "int32",
			GuestFunc:     "playerGameMode",
			GuestReturn:   "GameMode",
			GuestFallback: "GameMode(0)",
			GuestExpr:     "GameMode(h.PlayerGameMode(playerID))",
			RefMethod:     "GameMode",
			RefReturn:     "GameMode",
			RefExpr:       "playerGameMode(p.id)",
		},
		{
			HostMethod:  "SetPlayerGameMode",
			HostReturn:  "bool",
			HostArg:     &ports.BridgeArgSpec{Name: "mode", Type: "int32"},
			GuestFunc:   "setPlayerGameMode",
			GuestReturn: "bool",
			GuestArg:    &ports.BridgeArgSpec{Name: "mode", Type: "GameMode"},
			GuestExpr:   "h.SetPlayerGameMode(playerID, int32(mode))",
			RefMethod:   "SetGameMode",
			RefReturn:   "bool",
			RefArg:      &ports.BridgeArgSpec{Name: "mode", Type: "GameMode"},
			RefExpr:     "setPlayerGameMode(p.id, mode)",
		},

		getter("XUID", "string", `""`, "p.XUID()"),
		getter("DeviceID", "string", `""`, "p.DeviceID()"),
		getter("DeviceModel", "string", `""`, "p.DeviceModel()"),
		getter("SelfSignedID", "string", `""`, "p.SelfSignedID()"),

		getter("NameTag", "string", `""`, "p.NameTag()"),
		setter("NameTag", "string", "value", "p.SetNameTag(value)"),
		getter("ScoreTag", "string", `""`, "p.ScoreTag()"),
		setter("ScoreTag", "string", "value", "p.SetScoreTag(value)"),
		getter("Absorption", "float64", "float64(0)", "p.Absorption()"),
		setter("Absorption", "float64", "value", "p.SetAbsorption(value)"),
		getter("MaxHealth", "float64", "float64(0)", "p.MaxHealth()"),
		setter("MaxHealth", "float64", "value", "p.SetMaxHealth(value)"),
		getter("Speed", "float64", "float64(0)", "p.Speed()"),
		setter("Speed", "float64", "value", "p.SetSpeed(value)"),
		getter("FlightSpeed", "float64", "float64(0)", "p.FlightSpeed()"),
		setter("FlightSpeed", "float64", "value", "p.SetFlightSpeed(value)"),
		getter("VerticalFlightSpeed", "float64", "float64(0)", "p.VerticalFlightSpeed()"),
		setter("VerticalFlightSpeed", "float64", "value", "p.SetVerticalFlightSpeed(value)"),
		getter("Experience", "int32", "int32(0)", "int32(p.Experience())"),
		getter("ExperienceLevel", "int32", "int32(0)", "int32(p.ExperienceLevel())"),
		setter("ExperienceLevel", "int32", "value", "p.SetExperienceLevel(int(value))"),
		getter("ExperienceProgress", "float64", "float64(0)", "p.ExperienceProgress()"),
		setter("ExperienceProgress", "float64", "value", "p.SetExperienceProgress(value)"),

		getter("OnGround", "bool", "false", "p.OnGround()"),
		getter("Sneaking", "bool", "false", "p.Sneaking()"),
		toggle("Sneaking", "StartSneaking", "StopSneaking"),
		getter("Sprinting", "bool", "false", "p.Sprinting()"),
		toggle("Sprinting", "StartSprinting", "StopSprinting"),
		getter("Swimming", "bool", "false", "p.Swimming()"),
		toggle("Swimming", "StartSwimming", "StopSwimming"),
		getter("Flying", "bool", "false", "p.Flying()"),
		toggle("Flying", "StartFlying", "StopFlying"),
		getter("Gliding", "bool", "false", "p.Gliding()"),
		toggle("Gliding", "StartGliding", "StopGliding"),
		getter("Crawling", "bool", "false", "p.Crawling()"),
		toggle("Crawling", "StartCrawling", "StopCrawling"),
		getter("UsingItem", "bool", "false", "p.UsingItem()"),
		getter("Invisible", "bool", "false", "p.Invisible()"),
		toggle("Invisible", "SetInvisible", "SetVisible"),
		getter("Immobile", "bool", "false", "p.Immobile()"),
		toggle("Immobile", "SetImmobile", "SetMobile"),
		getter("Dead", "bool", "false", "p.Dead()"),

		{
			HostMethod:    "PlayerLatency",
			HostReturn:    "time.Duration",
			GuestFunc:     "playerLatency",
			GuestReturn:   "time.Duration",
			GuestFallback: "time.Duration(0)",
			GuestExpr:     "h.PlayerLatency(playerID)",
			RefMethod:     "Latency",
			RefReturn:     "time.Duration",
			RefExpr:       "playerLatency(p.id)",
		},

		withoutManager(setter("OnFireMillis", "int64", "millis", "p.SetOnFire(time.Duration(millis) * time.Millisecond)")),
		argBool("AddPlayerFood", "addPlayerFood", "AddFood", "points", "int32", "p.AddFood(int(points))"),
		action("UseItem"),
		action("Jump"),
		action("SwingArm"),
		action("Wake"),
		action("Extinguish"),
		toggle("ShowCoordinates", "ShowCoordinates", "HideCoordinates"),
		{
			HostMethod:    "PlayerMainHandItem",
			HostReturn:    "ItemStackData",
			GuestFunc:     "playerMainHandItem",
			GuestReturn:   "ItemStackData",
			GuestFallback: "ItemStackData{}",
			GuestExpr:     "h.PlayerMainHandItem(playerID)",
			RefMethod:     "MainHandItem",
			RefReturn:     "ItemStackData",
			RefExpr:       "playerMainHandItem(p.id)",
			ManagerBody:   "return playerValue(m, playerID, ItemStackData{}, func(p *player.Player) ItemStackData { return mainHandItemData(p) })",
		},
		{
			HostMethod:  "SetPlayerMainHandItem",
			HostReturn:  "bool",
			HostArg:     &ports.BridgeArgSpec{Name: "stack", Type: "ItemStackData"},
			GuestFunc:   "setPlayerMainHandItem",
			GuestReturn: "bool",
			GuestArg:    &ports.BridgeArgSpec{Name: "stack", Type: "ItemStackData"},
			GuestExpr:   "h.SetPlayerMainHandItem(playerID, stack)",
			RefMethod:   "SetMainHandItem",
			RefReturn:   "bool",
			RefArg:      &ports.BridgeArgSpec{Name: "stack", Type: "ItemStackData"},
			RefExpr:     "setPlayerMainHandItem(p.id, stack)",
			ManagerBody: "return playerValue(m, playerID, false, func(p *player.Player) bool { return setMainHandItemData(p, stack) })",
		},
		{
			HostMethod:    "PlayerOffHandItem",
			HostReturn:    "ItemStackData",
			GuestFunc:     "playerOffHandItem",
			GuestReturn:   "ItemStackData",
			GuestFallback: "ItemStackData{}",
			GuestExpr:     "h.PlayerOffHandItem(playerID)",
			RefMethod:     "OffHandItem",
			RefReturn:     "ItemStackData",
			RefExpr:       "playerOffHandItem(p.id)",
			ManagerBody:   "return playerValue(m, playerID, ItemStackData{}, func(p *player.Player) ItemStackData { return offHandItemData(p) })",
		},
		{
			HostMethod:  "SetPlayerOffHandItem",
			HostReturn:  "bool",
			HostArg:     &ports.BridgeArgSpec{Name: "stack", Type: "ItemStackData"},
			GuestFunc:   "setPlayerOffHandItem",
			GuestReturn: "bool",
			GuestArg:    &ports.BridgeArgSpec{Name: "stack", Type: "ItemStackData"},
			GuestExpr:   "h.SetPlayerOffHandItem(playerID, stack)",
			RefMethod:   "SetOffHandItem",
			RefReturn:   "bool",
			RefArg:      &ports.BridgeArgSpec{Name: "stack", Type: "ItemStackData"},
			RefExpr:     "setPlayerOffHandItem(p.id, stack)",
			ManagerBody: "return playerValue(m, playerID, false, func(p *player.Player) bool { return setOffHandItemData(p, stack) })",
		},
		{
			HostMethod:    "PlayerInventoryItems",
			HostReturn:    "[]ItemStackData",
			GuestFunc:     "playerInventoryItems",
			GuestReturn:   "[]ItemStackData",
			GuestFallback: "nil",
			GuestExpr:     "h.PlayerInventoryItems(playerID)",
			RefMethod:     "InventoryItems",
			RefReturn:     "[]ItemStackData",
			RefExpr:       "playerInventoryItems(p.id)",
			ManagerBody:   "return playerValue(m, playerID, nil, func(p *player.Player) []ItemStackData { return inventoryItemsData(p.Inventory()) })",
		},
		{
			HostMethod:  "SetPlayerInventoryItems",
			HostReturn:  "bool",
			HostArg:     &ports.BridgeArgSpec{Name: "items", Type: "[]ItemStackData"},
			GuestFunc:   "setPlayerInventoryItems",
			GuestReturn: "bool",
			GuestArg:    &ports.BridgeArgSpec{Name: "items", Type: "[]ItemStackData"},
			GuestExpr:   "h.SetPlayerInventoryItems(playerID, items)",
			RefMethod:   "SetInventoryItems",
			RefReturn:   "bool",
			RefArg:      &ports.BridgeArgSpec{Name: "items", Type: "[]ItemStackData"},
			RefExpr:     "setPlayerInventoryItems(p.id, items)",
			ManagerBody: "return playerValue(m, playerID, false, func(p *player.Player) bool { return setInventoryItemsData(p.Inventory(), items) })",
		},
		{
			HostMethod:    "PlayerEnderChestItems",
			HostReturn:    "[]ItemStackData",
			GuestFunc:     "playerEnderChestItems",
			GuestReturn:   "[]ItemStackData",
			GuestFallback: "nil",
			GuestExpr:     "h.PlayerEnderChestItems(playerID)",
			RefMethod:     "EnderChestItems",
			RefReturn:     "[]ItemStackData",
			RefExpr:       "playerEnderChestItems(p.id)",
			ManagerBody:   "return playerValue(m, playerID, nil, func(p *player.Player) []ItemStackData { return inventoryItemsData(p.EnderChestInventory()) })",
		},
		{
			HostMethod:  "SetPlayerEnderChestItems",
			HostReturn:  "bool",
			HostArg:     &ports.BridgeArgSpec{Name: "items", Type: "[]ItemStackData"},
			GuestFunc:   "setPlayerEnderChestItems",
			GuestReturn: "bool",
			GuestArg:    &ports.BridgeArgSpec{Name: "items", Type: "[]ItemStackData"},
			GuestExpr:   "h.SetPlayerEnderChestItems(playerID, items)",
			RefMethod:   "SetEnderChestItems",
			RefReturn:   "bool",
			RefArg:      &ports.BridgeArgSpec{Name: "items", Type: "[]ItemStackData"},
			RefExpr:     "setPlayerEnderChestItems(p.id, items)",
			ManagerBody: "return playerValue(m, playerID, false, func(p *player.Player) bool { return setInventoryItemsData(p.EnderChestInventory(), items) })",
		},
		{
			HostMethod:    "PlayerArmourItems",
			HostReturn:    "[]ItemStackData",
			GuestFunc:     "playerArmourItems",
			GuestReturn:   "[]ItemStackData",
			GuestFallback: "nil",
			GuestExpr:     "h.PlayerArmourItems(playerID)",
			RefMethod:     "ArmourItems",
			RefReturn:     "[]ItemStackData",
			RefExpr:       "playerArmourItems(p.id)",
			ManagerBody:   "return playerValue(m, playerID, nil, func(p *player.Player) []ItemStackData { return armourItemsData(p.Armour()) })",
		},
		{
			HostMethod:  "SetPlayerArmourItems",
			HostReturn:  "bool",
			HostArg:     &ports.BridgeArgSpec{Name: "items", Type: "[]ItemStackData"},
			GuestFunc:   "setPlayerArmourItems",
			GuestReturn: "bool",
			GuestArg:    &ports.BridgeArgSpec{Name: "items", Type: "[]ItemStackData"},
			GuestExpr:   "h.SetPlayerArmourItems(playerID, items)",
			RefMethod:   "SetArmourItems",
			RefReturn:   "bool",
			RefArg:      &ports.BridgeArgSpec{Name: "items", Type: "[]ItemStackData"},
			RefExpr:     "setPlayerArmourItems(p.id, items)",
			ManagerBody: "return playerValue(m, playerID, false, func(p *player.Player) bool { return setArmourItemsData(p.Armour(), items) })",
		},
		{
			HostMethod:  "SetPlayerHeldSlot",
			HostReturn:  "bool",
			HostArg:     &ports.BridgeArgSpec{Name: "slot", Type: "int32"},
			GuestFunc:   "setPlayerHeldSlot",
			GuestReturn: "bool",
			GuestArg:    &ports.BridgeArgSpec{Name: "slot", Type: "int32"},
			GuestExpr:   "h.SetPlayerHeldSlot(playerID, slot)",
			RefMethod:   "SetHeldSlot",
			RefReturn:   "bool",
			RefArg:      &ports.BridgeArgSpec{Name: "slot", Type: "int32"},
			RefExpr:     "setPlayerHeldSlot(p.id, slot)",
			ManagerBody: "return playerValue(m, playerID, false, func(p *player.Player) bool { return p.SetHeldSlot(int(slot)) == nil })",
		},
		action("MoveItemsToInventory"),
		action("CloseForm"),
		action("CloseDialogue"),
		{
			HostMethod:  "PlayerSendMenuForm",
			HostReturn:  "bool",
			HostArg:     &ports.BridgeArgSpec{Name: "value", Type: "MenuFormData"},
			GuestFunc:   "playerSendMenuForm",
			GuestReturn: "bool",
			GuestArg:    &ports.BridgeArgSpec{Name: "value", Type: "MenuFormData"},
			GuestExpr:   "h.PlayerSendMenuForm(playerID, value)",
			RefMethod:   "SendMenuForm",
			RefReturn:   "bool",
			RefArg:      &ports.BridgeArgSpec{Name: "value", Type: "MenuFormData"},
			RefExpr:     "playerSendMenuForm(p.id, value)",
			ManagerBody: "return playerValue(m, playerID, false, func(p *player.Player) bool { return sendMenuFormData(p, value) })",
		},
		{
			HostMethod:  "PlayerSendModalForm",
			HostReturn:  "bool",
			HostArg:     &ports.BridgeArgSpec{Name: "value", Type: "ModalFormData"},
			GuestFunc:   "playerSendModalForm",
			GuestReturn: "bool",
			GuestArg:    &ports.BridgeArgSpec{Name: "value", Type: "ModalFormData"},
			GuestExpr:   "h.PlayerSendModalForm(playerID, value)",
			RefMethod:   "SendModalForm",
			RefReturn:   "bool",
			RefArg:      &ports.BridgeArgSpec{Name: "value", Type: "ModalFormData"},
			RefExpr:     "playerSendModalForm(p.id, value)",
			ManagerBody: "return playerValue(m, playerID, false, func(p *player.Player) bool { return sendModalFormData(p, value) })",
		},
		argVoid("PlayerMessage", "playerMessage", "Message", "message", "string", "p.Message(message)"),
	}
}

func getter(name, typ, fallback, managerExpr string) ports.BridgeOpSpec {
	hostMethod := "Player" + name
	guestFunc := "player" + name
	return ports.BridgeOpSpec{
		HostMethod:    hostMethod,
		HostReturn:    typ,
		GuestFunc:     guestFunc,
		GuestReturn:   typ,
		GuestFallback: fallback,
		GuestExpr:     "h." + hostMethod + "(playerID)",
		RefMethod:     name,
		RefReturn:     typ,
		RefExpr:       guestFunc + "(p.id)",
		ManagerBody:   "return playerValue(m, playerID, " + fallback + ", func(p *player.Player) " + typ + " { return " + managerExpr + " })",
	}
}

func setter(name, argType, argName, managerStmt string) ports.BridgeOpSpec {
	hostMethod := "SetPlayer" + name
	guestFunc := "setPlayer" + name
	return ports.BridgeOpSpec{
		HostMethod:  hostMethod,
		HostReturn:  "bool",
		HostArg:     &ports.BridgeArgSpec{Name: argName, Type: argType},
		GuestFunc:   guestFunc,
		GuestReturn: "bool",
		GuestArg:    &ports.BridgeArgSpec{Name: argName, Type: argType},
		GuestExpr:   "h." + hostMethod + "(playerID, " + argName + ")",
		RefMethod:   "Set" + name,
		RefReturn:   "bool",
		RefArg:      &ports.BridgeArgSpec{Name: argName, Type: argType},
		RefExpr:     guestFunc + "(p.id, " + argName + ")",
		ManagerBody: "return m.playerUpdate(playerID, func(p *player.Player) { " + managerStmt + " })",
	}
}

func toggle(name, onMethod, offMethod string) ports.BridgeOpSpec {
	op := setter(name, "bool", "value", "")
	op.ManagerBody = "return m.playerToggle(playerID, value, (*player.Player)." + onMethod + ", (*player.Player)." + offMethod + ")"
	return op
}

func action(name string) ports.BridgeOpSpec {
	hostMethod := "Player" + name
	guestFunc := "player" + name
	return ports.BridgeOpSpec{
		HostMethod:  hostMethod,
		HostReturn:  "bool",
		GuestFunc:   guestFunc,
		GuestReturn: "bool",
		GuestExpr:   "h." + hostMethod + "(playerID)",
		RefMethod:   name,
		RefReturn:   "bool",
		RefExpr:     guestFunc + "(p.id)",
		ManagerBody: "return m.playerUpdate(playerID, (*player.Player)." + name + ")",
	}
}

func argBool(hostMethod, guestFunc, refMethod, argName, argType, managerStmt string) ports.BridgeOpSpec {
	return ports.BridgeOpSpec{
		HostMethod:  hostMethod,
		HostReturn:  "bool",
		HostArg:     &ports.BridgeArgSpec{Name: argName, Type: argType},
		GuestFunc:   guestFunc,
		GuestReturn: "bool",
		GuestArg:    &ports.BridgeArgSpec{Name: argName, Type: argType},
		GuestExpr:   "h." + hostMethod + "(playerID, " + argName + ")",
		RefMethod:   refMethod,
		RefReturn:   "bool",
		RefArg:      &ports.BridgeArgSpec{Name: argName, Type: argType},
		RefExpr:     guestFunc + "(p.id, " + argName + ")",
		ManagerBody: "return m.playerUpdate(playerID, func(p *player.Player) { " + managerStmt + " })",
	}
}

func argVoid(hostMethod, guestFunc, refMethod, argName, argType, managerStmt string) ports.BridgeOpSpec {
	return ports.BridgeOpSpec{
		HostMethod: hostMethod,
		HostArg:    &ports.BridgeArgSpec{Name: argName, Type: argType},
		GuestFunc:  guestFunc,
		GuestArg:   &ports.BridgeArgSpec{Name: argName, Type: argType},
		GuestExpr:  "h." + hostMethod + "(playerID, " + argName + ")",
		RefMethod:  refMethod,
		RefArg:     &ports.BridgeArgSpec{Name: argName, Type: argType},
		RefExpr:    guestFunc + "(p.id, " + argName + ")",
		ManagerBody: "_ = m.playerUpdate(playerID, func(p *player.Player) { " +
			managerStmt + " })",
	}
}

func withoutManager(op ports.BridgeOpSpec) ports.BridgeOpSpec {
	op.ManagerBody = ""
	return op
}
