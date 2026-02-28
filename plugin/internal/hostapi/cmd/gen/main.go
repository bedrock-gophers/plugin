package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

type argSpec struct {
	Name string
	Type string
}

type opSpec struct {
	HostMethod string
	HostReturn string
	HostArg    *argSpec

	GuestFunc     string
	GuestReturn   string
	GuestArg      *argSpec
	GuestFallback string
	GuestExpr     string

	RefMethod string
	RefReturn string
	RefArg    *argSpec
	RefExpr   string

	ManagerBody string
}

func main() {
	ops := operations()
	if err := validateOps(ops); err != nil {
		fail("validate ops", err)
	}
	if err := writeFormatted(filepath.Join("..", "..", "sdk", "go", "player_host_generated.go"), renderGuestHost(ops)); err != nil {
		fail("generate sdk/go host wrappers", err)
	}
	if err := writeFormatted(filepath.Join("..", "..", "sdk", "go", "player_ref_generated.go"), renderPlayerRef(ops)); err != nil {
		fail("generate player ref wrappers", err)
	}
	if err := writeFormatted(filepath.Join("..", "..", "manager_player_generated.go"), renderManager(ops)); err != nil {
		fail("generate manager wrappers", err)
	}
}

func operations() []opSpec {
	ops := []opSpec{
		getter("Health", "float64", "float64(0)", "p.Health()"),
		noManager(setter("Health", "float64", "health", "p.SetHealth(health)")),

		getter("Food", "int32", "int32(0)", "int32(p.Food())"),
		setter("Food", "int32", "food", "p.SetFood(int(food))"),
		getter("Name", "string", "\"\"", "p.Name()"),

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
			HostArg:     &argSpec{Name: "mode", Type: "int32"},
			GuestFunc:   "setPlayerGameMode",
			GuestReturn: "bool",
			GuestArg:    &argSpec{Name: "mode", Type: "GameMode"},
			GuestExpr:   "h.SetPlayerGameMode(playerID, int32(mode))",
			RefMethod:   "SetGameMode",
			RefReturn:   "bool",
			RefArg:      &argSpec{Name: "mode", Type: "GameMode"},
			RefExpr:     "setPlayerGameMode(p.id, mode)",
		},

		getter("XUID", "string", "\"\"", "p.XUID()"),
		getter("DeviceID", "string", "\"\"", "p.DeviceID()"),
		getter("DeviceModel", "string", "\"\"", "p.DeviceModel()"),
		getter("SelfSignedID", "string", "\"\"", "p.SelfSignedID()"),

		getter("NameTag", "string", "\"\"", "p.NameTag()"),
		setter("NameTag", "string", "value", "p.SetNameTag(value)"),
		getter("ScoreTag", "string", "\"\"", "p.ScoreTag()"),
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

		noManager(setter("OnFireMillis", "int64", "millis", "p.SetOnFire(time.Duration(millis) * time.Millisecond)")),
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
			HostArg:     &argSpec{Name: "stack", Type: "ItemStackData"},
			GuestFunc:   "setPlayerMainHandItem",
			GuestReturn: "bool",
			GuestArg:    &argSpec{Name: "stack", Type: "ItemStackData"},
			GuestExpr:   "h.SetPlayerMainHandItem(playerID, stack)",
			RefMethod:   "SetMainHandItem",
			RefReturn:   "bool",
			RefArg:      &argSpec{Name: "stack", Type: "ItemStackData"},
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
			HostArg:     &argSpec{Name: "stack", Type: "ItemStackData"},
			GuestFunc:   "setPlayerOffHandItem",
			GuestReturn: "bool",
			GuestArg:    &argSpec{Name: "stack", Type: "ItemStackData"},
			GuestExpr:   "h.SetPlayerOffHandItem(playerID, stack)",
			RefMethod:   "SetOffHandItem",
			RefReturn:   "bool",
			RefArg:      &argSpec{Name: "stack", Type: "ItemStackData"},
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
			HostArg:     &argSpec{Name: "items", Type: "[]ItemStackData"},
			GuestFunc:   "setPlayerInventoryItems",
			GuestReturn: "bool",
			GuestArg:    &argSpec{Name: "items", Type: "[]ItemStackData"},
			GuestExpr:   "h.SetPlayerInventoryItems(playerID, items)",
			RefMethod:   "SetInventoryItems",
			RefReturn:   "bool",
			RefArg:      &argSpec{Name: "items", Type: "[]ItemStackData"},
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
			HostArg:     &argSpec{Name: "items", Type: "[]ItemStackData"},
			GuestFunc:   "setPlayerEnderChestItems",
			GuestReturn: "bool",
			GuestArg:    &argSpec{Name: "items", Type: "[]ItemStackData"},
			GuestExpr:   "h.SetPlayerEnderChestItems(playerID, items)",
			RefMethod:   "SetEnderChestItems",
			RefReturn:   "bool",
			RefArg:      &argSpec{Name: "items", Type: "[]ItemStackData"},
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
			HostArg:     &argSpec{Name: "items", Type: "[]ItemStackData"},
			GuestFunc:   "setPlayerArmourItems",
			GuestReturn: "bool",
			GuestArg:    &argSpec{Name: "items", Type: "[]ItemStackData"},
			GuestExpr:   "h.SetPlayerArmourItems(playerID, items)",
			RefMethod:   "SetArmourItems",
			RefReturn:   "bool",
			RefArg:      &argSpec{Name: "items", Type: "[]ItemStackData"},
			RefExpr:     "setPlayerArmourItems(p.id, items)",
			ManagerBody: "return playerValue(m, playerID, false, func(p *player.Player) bool { return setArmourItemsData(p.Armour(), items) })",
		},
		{
			HostMethod:  "SetPlayerHeldSlot",
			HostReturn:  "bool",
			HostArg:     &argSpec{Name: "slot", Type: "int32"},
			GuestFunc:   "setPlayerHeldSlot",
			GuestReturn: "bool",
			GuestArg:    &argSpec{Name: "slot", Type: "int32"},
			GuestExpr:   "h.SetPlayerHeldSlot(playerID, slot)",
			RefMethod:   "SetHeldSlot",
			RefReturn:   "bool",
			RefArg:      &argSpec{Name: "slot", Type: "int32"},
			RefExpr:     "setPlayerHeldSlot(p.id, slot)",
			ManagerBody: "return playerValue(m, playerID, false, func(p *player.Player) bool { return p.SetHeldSlot(int(slot)) == nil })",
		},
		action("MoveItemsToInventory"),
		action("CloseForm"),
		action("CloseDialogue"),
		{
			HostMethod:  "PlayerSendMenuForm",
			HostReturn:  "bool",
			HostArg:     &argSpec{Name: "value", Type: "MenuFormData"},
			GuestFunc:   "playerSendMenuForm",
			GuestReturn: "bool",
			GuestArg:    &argSpec{Name: "value", Type: "MenuFormData"},
			GuestExpr:   "h.PlayerSendMenuForm(playerID, value)",
			RefMethod:   "SendMenuForm",
			RefReturn:   "bool",
			RefArg:      &argSpec{Name: "value", Type: "MenuFormData"},
			RefExpr:     "playerSendMenuForm(p.id, value)",
			ManagerBody: "return playerValue(m, playerID, false, func(p *player.Player) bool { return sendMenuFormData(p, value) })",
		},
		{
			HostMethod:  "PlayerSendModalForm",
			HostReturn:  "bool",
			HostArg:     &argSpec{Name: "value", Type: "ModalFormData"},
			GuestFunc:   "playerSendModalForm",
			GuestReturn: "bool",
			GuestArg:    &argSpec{Name: "value", Type: "ModalFormData"},
			GuestExpr:   "h.PlayerSendModalForm(playerID, value)",
			RefMethod:   "SendModalForm",
			RefReturn:   "bool",
			RefArg:      &argSpec{Name: "value", Type: "ModalFormData"},
			RefExpr:     "playerSendModalForm(p.id, value)",
			ManagerBody: "return playerValue(m, playerID, false, func(p *player.Player) bool { return sendModalFormData(p, value) })",
		},
		argVoid("PlayerMessage", "playerMessage", "Message", "message", "string", "p.Message(message)"),
	}
	return ops
}

func getter(name, typ, fallback, managerExpr string) opSpec {
	hostMethod := "Player" + name
	guestFunc := "player" + name
	return opSpec{
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

func setter(name, argType, argName, managerStmt string) opSpec {
	hostMethod := "SetPlayer" + name
	guestFunc := "setPlayer" + name
	return opSpec{
		HostMethod:  hostMethod,
		HostReturn:  "bool",
		HostArg:     &argSpec{Name: argName, Type: argType},
		GuestFunc:   guestFunc,
		GuestReturn: "bool",
		GuestArg:    &argSpec{Name: argName, Type: argType},
		GuestExpr:   "h." + hostMethod + "(playerID, " + argName + ")",
		RefMethod:   "Set" + name,
		RefReturn:   "bool",
		RefArg:      &argSpec{Name: argName, Type: argType},
		RefExpr:     guestFunc + "(p.id, " + argName + ")",
		ManagerBody: "return m.playerUpdate(playerID, func(p *player.Player) { " + managerStmt + " })",
	}
}

func toggle(name, onMethod, offMethod string) opSpec {
	op := setter(name, "bool", "value", "")
	op.ManagerBody = "return m.playerToggle(playerID, value, (*player.Player)." + onMethod + ", (*player.Player)." + offMethod + ")"
	return op
}

func action(name string) opSpec {
	hostMethod := "Player" + name
	guestFunc := "player" + name
	return opSpec{
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

func argBool(hostMethod, guestFunc, refMethod, argName, argType, managerStmt string) opSpec {
	return opSpec{
		HostMethod:  hostMethod,
		HostReturn:  "bool",
		HostArg:     &argSpec{Name: argName, Type: argType},
		GuestFunc:   guestFunc,
		GuestReturn: "bool",
		GuestArg:    &argSpec{Name: argName, Type: argType},
		GuestExpr:   "h." + hostMethod + "(playerID, " + argName + ")",
		RefMethod:   refMethod,
		RefReturn:   "bool",
		RefArg:      &argSpec{Name: argName, Type: argType},
		RefExpr:     guestFunc + "(p.id, " + argName + ")",
		ManagerBody: "return m.playerUpdate(playerID, func(p *player.Player) { " + managerStmt + " })",
	}
}

func argVoid(hostMethod, guestFunc, refMethod, argName, argType, managerStmt string) opSpec {
	return opSpec{
		HostMethod: hostMethod,
		HostArg:    &argSpec{Name: argName, Type: argType},
		GuestFunc:  guestFunc,
		GuestArg:   &argSpec{Name: argName, Type: argType},
		GuestExpr:  "h." + hostMethod + "(playerID, " + argName + ")",
		RefMethod:  refMethod,
		RefArg:     &argSpec{Name: argName, Type: argType},
		RefExpr:    guestFunc + "(p.id, " + argName + ")",
		ManagerBody: "_ = m.playerUpdate(playerID, func(p *player.Player) { " +
			managerStmt + " })",
	}
}

func noManager(op opSpec) opSpec {
	op.ManagerBody = ""
	return op
}

func validateOps(ops []opSpec) error {
	if len(ops) == 0 {
		return fmt.Errorf("operation table is empty")
	}
	seenHost := map[string]struct{}{}
	seenGuest := map[string]struct{}{}
	seenRef := map[string]struct{}{}
	seenManager := map[string]struct{}{}

	for i, op := range ops {
		id := fmt.Sprintf("ops[%d]", i)
		if op.HostMethod == "" {
			return fmt.Errorf("%s: HostMethod is required", id)
		}
		if op.GuestFunc == "" {
			return fmt.Errorf("%s: GuestFunc is required", id)
		}
		if op.RefMethod == "" {
			return fmt.Errorf("%s: RefMethod is required", id)
		}
		if op.GuestExpr == "" {
			return fmt.Errorf("%s: GuestExpr is required", id)
		}
		if op.RefExpr == "" {
			return fmt.Errorf("%s: RefExpr is required", id)
		}
		if op.HostArg != nil {
			if op.HostArg.Name == "" || op.HostArg.Type == "" {
				return fmt.Errorf("%s: HostArg requires name and type", id)
			}
		}
		if op.GuestArg != nil {
			if op.GuestArg.Name == "" || op.GuestArg.Type == "" {
				return fmt.Errorf("%s: GuestArg requires name and type", id)
			}
		}
		if op.RefArg != nil {
			if op.RefArg.Name == "" || op.RefArg.Type == "" {
				return fmt.Errorf("%s: RefArg requires name and type", id)
			}
		}
		if op.HostArg == nil && (op.GuestArg != nil || op.RefArg != nil) {
			return fmt.Errorf("%s: GuestArg/RefArg provided without HostArg", id)
		}
		if op.HostArg != nil && (op.GuestArg == nil || op.RefArg == nil) {
			return fmt.Errorf("%s: HostArg requires GuestArg and RefArg", id)
		}
		if op.GuestReturn == "" && op.HostReturn != "" && op.HostReturn != "bool" {
			return fmt.Errorf("%s: GuestReturn missing for value op", id)
		}
		if op.GuestReturn != "" && op.GuestReturn != "bool" && op.GuestFallback == "" {
			return fmt.Errorf("%s: GuestFallback missing for GuestReturn %q", id, op.GuestReturn)
		}

		if _, ok := seenHost[op.HostMethod]; ok {
			return fmt.Errorf("%s: duplicate HostMethod %q", id, op.HostMethod)
		}
		seenHost[op.HostMethod] = struct{}{}

		if _, ok := seenGuest[op.GuestFunc]; ok {
			return fmt.Errorf("%s: duplicate GuestFunc %q", id, op.GuestFunc)
		}
		seenGuest[op.GuestFunc] = struct{}{}

		if _, ok := seenRef[op.RefMethod]; ok {
			return fmt.Errorf("%s: duplicate RefMethod %q", id, op.RefMethod)
		}
		seenRef[op.RefMethod] = struct{}{}

		if op.ManagerBody != "" {
			if _, ok := seenManager[op.HostMethod]; ok {
				return fmt.Errorf("%s: duplicate manager method %q", id, op.HostMethod)
			}
			seenManager[op.HostMethod] = struct{}{}
		}
	}
	return nil
}

func renderGuestHost(ops []opSpec) []byte {
	var b bytes.Buffer
	b.WriteString("// Code generated by go generate; DO NOT EDIT.\n")
	b.WriteString("package guest\n\n")
	b.WriteString("type playerHost interface {\n")
	for _, op := range ops {
		b.WriteString("\t")
		b.WriteString(methodSignature(op.HostMethod, withPlayerID(op.HostArg), op.HostReturn))
		b.WriteString("\n")
	}
	b.WriteString("}\n\n")

	for i, op := range ops {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("func ")
		b.WriteString(methodSignature(op.GuestFunc, withPlayerID(op.GuestArg), op.GuestReturn))
		b.WriteString(" {\n")
		switch op.GuestReturn {
		case "":
			b.WriteString("\thostDo(func(h Host) { ")
			b.WriteString(op.GuestExpr)
			b.WriteString(" })\n")
		case "bool":
			b.WriteString("\treturn hostBool(func(h Host) bool { return ")
			b.WriteString(op.GuestExpr)
			b.WriteString(" })\n")
		default:
			b.WriteString("\treturn hostValue(")
			b.WriteString(op.GuestFallback)
			b.WriteString(", func(h Host) ")
			b.WriteString(op.GuestReturn)
			b.WriteString(" { return ")
			b.WriteString(op.GuestExpr)
			b.WriteString(" })\n")
		}
		b.WriteString("}\n")
	}

	return b.Bytes()
}

func renderPlayerRef(ops []opSpec) []byte {
	var b bytes.Buffer
	b.WriteString("// Code generated by go generate; DO NOT EDIT.\n")
	b.WriteString("package guest\n\n")
	b.WriteString("import \"github.com/sandertv/gophertunnel/minecraft/text\"\n\n")
	for i, op := range ops {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("func (p PlayerRef) ")
		b.WriteString(methodSignature(op.RefMethod, maybeArg(op.RefArg), op.RefReturn))
		b.WriteString(" {\n")
		if op.RefReturn == "" {
			b.WriteString("\t")
			b.WriteString(op.RefExpr)
			b.WriteString("\n")
		} else {
			b.WriteString("\treturn ")
			b.WriteString(op.RefExpr)
			b.WriteString("\n")
		}
		b.WriteString("}\n")
	}

	b.WriteString("\n")
	b.WriteString("func (p PlayerRef) Messagef(format string, a ...any) {\n")
	b.WriteString("\tp.Message(text.Colourf(format, a...))\n")
	b.WriteString("}\n")
	return b.Bytes()
}

func renderManager(ops []opSpec) []byte {
	var b bytes.Buffer
	b.WriteString("// Code generated by go generate; DO NOT EDIT.\n")
	b.WriteString("package plugin\n\n")
	b.WriteString("import \"github.com/df-mc/dragonfly/server/player\"\n\n")

	first := true
	for _, op := range ops {
		if op.ManagerBody == "" {
			continue
		}
		if !first {
			b.WriteString("\n")
		}
		first = false

		b.WriteString("func (m *Manager) ")
		b.WriteString(methodSignature(op.HostMethod, withPlayerID(op.HostArg), op.HostReturn))
		b.WriteString(" {\n")
		for _, line := range strings.Split(op.ManagerBody, "\n") {
			b.WriteString("\t")
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("}\n")
	}
	return b.Bytes()
}

func withPlayerID(extra *argSpec) []argSpec {
	args := []argSpec{{Name: "playerID", Type: "uint64"}}
	if extra != nil {
		args = append(args, *extra)
	}
	return args
}

func maybeArg(arg *argSpec) []argSpec {
	if arg == nil {
		return nil
	}
	return []argSpec{*arg}
}

func methodSignature(name string, args []argSpec, ret string) string {
	var b strings.Builder
	b.WriteString(name)
	b.WriteString("(")
	for i, arg := range args {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(arg.Name)
		b.WriteString(" ")
		b.WriteString(arg.Type)
	}
	b.WriteString(")")
	if ret != "" {
		b.WriteString(" ")
		b.WriteString(ret)
	}
	return b.String()
}

func writeFormatted(path string, src []byte) error {
	formatted, err := format.Source(src)
	if err != nil {
		return fmt.Errorf("format source for %s: %w", path, err)
	}
	if err := os.WriteFile(path, formatted, 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func fail(step string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", step, err)
	os.Exit(1)
}
