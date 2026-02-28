package guest

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/bedrock-gophers/plugin/plugin/abi"
	"github.com/bedrock-gophers/plugin/plugin/internal/ctxkey"
)

// Base exposes shorthand event registration helpers.
var Base baseEvents

type baseEvents struct{}

// HandleEvent registers any supported event handler signature.
//
// For signatures shared by multiple events (for example func(*Event)),
// the target event is inferred from the handler function name (for example onItemUse/onPunchAir).
func (baseEvents) HandleEvent(fn any) {
	switch h := fn.(type) {
	case func(ev *Event, newPos Vec3, newRot Rotation):
		registerMove(h)
	case func(ev *Event, info PlayerIdentity):
		registerByName(fn,
			eventChoice{name: "jump", hints: []string{"jump"}, register: func() { registerJump(h) }},
			eventChoice{name: "quit", hints: []string{"quit"}, register: func() { registerQuit(h) }},
		)
	case func(ev *Event, info PlayerIdentity, cancelMessage MutableArgument[string]):
		registerJoin(h)
	case func(ev *Event, pos Vec3):
		registerTeleport(h)
	case func(ev *Event, before, after WorldData):
		registerChangeWorld(h)
	case func(ev *Event, after bool):
		registerByName(fn,
			eventChoice{name: "toggle sprint", hints: []string{"togglesprint", "sprint"}, register: func() { registerToggleSprint(h) }},
			eventChoice{name: "toggle sneak", hints: []string{"togglesneak", "sneak"}, register: func() { registerToggleSneak(h) }},
		)
	case func(ev *Event, message MutableArgument[string]):
		registerByName(fn,
			eventChoice{name: "chat", hints: []string{"chat"}, register: func() { registerChat(h) }},
			eventChoice{name: "transfer", hints: []string{"transfer"}, register: func() { registerTransfer(h) }},
		)
	case func(ev *Event, from int32, to MutableArgument[int]):
		registerFoodLoss(h)
	case func(ev *Event, amount MutableArgument[float64], src HealingSourceData):
		registerHeal(h)
	case func(ev *Event, damage MutableArgument[float64], immune bool, attackImmunityMillis MutableArgument[int64], src DamageSourceData):
		registerHurt(h)
	case func(ev *Event, src DamageSourceData, keepInventory MutableArgument[bool]):
		registerDeath(h)
	case func(ev *Event, pos MutableVec3, worldName MutableArgument[string]):
		registerRespawn(h)
	case func(ev *Event, skin SkinData, fullID MutableArgument[string]):
		registerSkinChange(h)
	case func(ev *Event, pos Pos):
		registerByName(fn,
			eventChoice{name: "fire extinguish", hints: []string{"fireextinguish"}, register: func() { registerFireExtinguish(h) }},
			eventChoice{name: "start break", hints: []string{"startbreak"}, register: func() { registerStartBreak(h) }},
		)
	case func(ev *Event, pos Pos, drops []ItemStackData, xp MutableArgument[int]):
		registerBlockBreak(h)
	case func(ev *Event, pos Pos, block BlockData):
		registerByName(fn,
			eventChoice{name: "block place", hints: []string{"blockplace"}, register: func() { registerBlockPlace(h) }},
			eventChoice{name: "block pick", hints: []string{"blockpick"}, register: func() { registerBlockPick(h) }},
		)
	case func(ev *Event):
		registerByName(fn,
			eventChoice{name: "item use", hints: []string{"itemuse"}, register: func() { registerItemUse(h) }},
			eventChoice{name: "punch air", hints: []string{"punchair"}, register: func() { registerPunchAir(h) }},
		)
	case func(ev *Event, pos Pos, face uint8, clickPos Vec3):
		registerItemUseOnBlock(h)
	case func(ev *Event, ent EntityData):
		registerItemUseOnEntity(h)
	case func(ev *Event, item ItemStackData, durationMillis int64):
		registerItemRelease(h)
	case func(ev *Event, item ItemStackData):
		registerByName(fn,
			eventChoice{name: "item consume", hints: []string{"itemconsume"}, register: func() { registerItemConsume(h) }},
			eventChoice{name: "item drop", hints: []string{"itemdrop"}, register: func() { registerItemDrop(h) }},
		)
	case func(ev *Event, ent EntityData, force MutableArgument[float64], height MutableArgument[float64], critical MutableArgument[bool]):
		registerAttackEntity(h)
	case func(ev *Event, amount MutableArgument[int]):
		registerExperienceGain(h)
	case func(ev *Event, pos Pos, frontSide bool, oldText, newText string):
		registerSignEdit(h)
	case func(ev *Event, sendReminder MutableArgument[bool]):
		registerSleep(h)
	case func(ev *Event, pos Pos, oldPage int32, newPage MutableArgument[int]):
		registerLecternPageTurn(h)
	case func(ev *Event, item ItemStackData, damage MutableArgument[int]):
		registerByName(fn,
			eventChoice{name: "item damage", hints: []string{"itemdamage"}, register: func() { registerItemDamage(h) }},
			eventChoice{name: "item pickup", hints: []string{"itempickup"}, register: func() { registerItemPickup(h) }},
		)
	case func(ev *Event, from, to int32):
		registerHeldSlotChange(h)
	case func(ev *Event, command CommandData):
		registerCommandExecution(h)
	case func(ev *Event, d DiagnosticsData):
		registerDiagnostics(h)
	default:
		panic(fmt.Sprintf("guest.Base.HandleEvent: unsupported signature %T", fn))
	}
}

func registerMove(fn func(ev *Event, newPos Vec3, newRot Rotation)) {
	handle(abi.EventMove, func(ev *Event) {
		d := ev.Decoder()
		fn(ev, decodeVec3(d), decodeRotation(d))
	})
}

func registerJump(fn func(ev *Event, info PlayerIdentity)) {
	handle(abi.EventJump, func(ev *Event) {
		fn(ev, decodePlayerIdentity(ev.Decoder()))
	})
}

func registerJoin(fn func(ev *Event, info PlayerIdentity, cancelMessage MutableArgument[string])) {
	handle(abi.EventJoin, func(ev *Event) {
		fn(ev, decodePlayerIdentity(ev.Decoder()), mutableString(ev, ctxkey.JoinCancelMessage))
	})
}

func registerTeleport(fn func(ev *Event, pos Vec3)) {
	handle(abi.EventTeleport, func(ev *Event) {
		fn(ev, decodeVec3(ev.Decoder()))
	})
}

func registerChangeWorld(fn func(ev *Event, before, after WorldData)) {
	handle(abi.EventChangeWorld, func(ev *Event) {
		d := ev.Decoder()
		fn(ev, decodeWorld(d), decodeWorld(d))
	})
}

func registerToggleSprint(fn func(ev *Event, after bool)) {
	handle(abi.EventToggleSprint, func(ev *Event) {
		fn(ev, ev.Decoder().Bool())
	})
}

func registerToggleSneak(fn func(ev *Event, after bool)) {
	handle(abi.EventToggleSneak, func(ev *Event) {
		fn(ev, ev.Decoder().Bool())
	})
}

func registerChat(fn func(ev *Event, message MutableArgument[string])) {
	handle(abi.EventChat, func(ev *Event) {
		_ = ev.Decoder().String()
		fn(ev, mutableString(ev, ctxkey.ChatMessage))
	})
}

func registerFoodLoss(fn func(ev *Event, from int32, to MutableArgument[int])) {
	handle(abi.EventFoodLoss, func(ev *Event) {
		d := ev.Decoder()
		from := d.I32()
		_ = d.I32()
		fn(ev, from, mutableInt(ev, ctxkey.FoodTo))
	})
}

func registerHeal(fn func(ev *Event, amount MutableArgument[float64], src HealingSourceData)) {
	handle(abi.EventHeal, func(ev *Event) {
		d := ev.Decoder()
		_ = d.F64()
		fn(ev, mutableFloat64(ev, ctxkey.HealAmount), decodeHealingSource(d))
	})
}

func registerHurt(fn func(ev *Event, damage MutableArgument[float64], immune bool, attackImmunityMillis MutableArgument[int64], src DamageSourceData)) {
	handle(abi.EventHurt, func(ev *Event) {
		d := ev.Decoder()
		_ = d.F64()
		immune := d.Bool()
		_ = d.I64()
		fn(ev, mutableFloat64(ev, ctxkey.HurtDamage), immune, mutableInt64(ev, ctxkey.HurtAttackImmunityMillis), decodeDamageSource(d))
	})
}

func registerDeath(fn func(ev *Event, src DamageSourceData, keepInventory MutableArgument[bool])) {
	handle(abi.EventDeath, func(ev *Event) {
		d := ev.Decoder()
		src := decodeDamageSource(d)
		_ = d.Bool()
		fn(ev, src, mutableBool(ev, ctxkey.DeathKeepInventory))
	})
}

func registerRespawn(fn func(ev *Event, pos MutableVec3, worldName MutableArgument[string])) {
	handle(abi.EventRespawn, func(ev *Event) {
		d := ev.Decoder()
		_ = decodeVec3(d)
		_ = d.String()
		fn(ev, MutableVec3{
			X: mutableFloat64(ev, ctxkey.RespawnPosX),
			Y: mutableFloat64(ev, ctxkey.RespawnPosY),
			Z: mutableFloat64(ev, ctxkey.RespawnPosZ),
		}, mutableString(ev, ctxkey.RespawnWorldName))
	})
}

func registerSkinChange(fn func(ev *Event, skin SkinData, fullID MutableArgument[string])) {
	handle(abi.EventSkinChange, func(ev *Event) {
		d := ev.Decoder()
		skin := decodeSkin(d)
		fn(ev, skin, mutableString(ev, ctxkey.ChatMessage))
	})
}

func registerFireExtinguish(fn func(ev *Event, pos Pos)) {
	handle(abi.EventFireExtinguish, func(ev *Event) {
		fn(ev, decodePos(ev.Decoder()))
	})
}

func registerStartBreak(fn func(ev *Event, pos Pos)) {
	handle(abi.EventStartBreak, func(ev *Event) {
		fn(ev, decodePos(ev.Decoder()))
	})
}

func registerBlockBreak(fn func(ev *Event, pos Pos, drops []ItemStackData, xp MutableArgument[int])) {
	handle(abi.EventBlockBreak, func(ev *Event) {
		d := ev.Decoder()
		pos := decodePos(d)
		drops := decodeItemStacks(d)
		_ = d.I32()
		fn(ev, pos, drops, mutableInt(ev, ctxkey.BlockBreakXP))
	})
}

func registerBlockPlace(fn func(ev *Event, pos Pos, block BlockData)) {
	handle(abi.EventBlockPlace, func(ev *Event) {
		d := ev.Decoder()
		fn(ev, decodePos(d), decodeBlock(d))
	})
}

func registerBlockPick(fn func(ev *Event, pos Pos, block BlockData)) {
	handle(abi.EventBlockPick, func(ev *Event) {
		d := ev.Decoder()
		fn(ev, decodePos(d), decodeBlock(d))
	})
}

func registerItemUse(fn func(ev *Event)) {
	handle(abi.EventItemUse, fn)
}

func registerItemUseOnBlock(fn func(ev *Event, pos Pos, face uint8, clickPos Vec3)) {
	handle(abi.EventItemUseOnBlock, func(ev *Event) {
		d := ev.Decoder()
		pos := decodePos(d)
		face := d.U8()
		fn(ev, pos, face, decodeVec3(d))
	})
}

func registerItemUseOnEntity(fn func(ev *Event, ent EntityData)) {
	handle(abi.EventItemUseOnEntity, func(ev *Event) {
		fn(ev, decodeEntity(ev.Decoder()))
	})
}

func registerItemRelease(fn func(ev *Event, item ItemStackData, durationMillis int64)) {
	handle(abi.EventItemRelease, func(ev *Event) {
		d := ev.Decoder()
		fn(ev, decodeItemStack(d), d.I64())
	})
}

func registerItemConsume(fn func(ev *Event, item ItemStackData)) {
	handle(abi.EventItemConsume, func(ev *Event) {
		fn(ev, decodeItemStack(ev.Decoder()))
	})
}

func registerAttackEntity(fn func(ev *Event, ent EntityData, force MutableArgument[float64], height MutableArgument[float64], critical MutableArgument[bool])) {
	handle(abi.EventAttackEntity, func(ev *Event) {
		d := ev.Decoder()
		ent := decodeEntity(d)
		_ = d.F64()
		_ = d.F64()
		_ = d.Bool()
		fn(ev, ent, mutableFloat64(ev, ctxkey.AttackForce), mutableFloat64(ev, ctxkey.AttackHeight), mutableBool(ev, ctxkey.AttackCritical))
	})
}

func registerExperienceGain(fn func(ev *Event, amount MutableArgument[int])) {
	handle(abi.EventExperienceGain, func(ev *Event) {
		_ = ev.Decoder().I32()
		fn(ev, mutableInt(ev, ctxkey.ExperienceAmount))
	})
}

func registerPunchAir(fn func(ev *Event)) {
	handle(abi.EventPunchAir, fn)
}

func registerSignEdit(fn func(ev *Event, pos Pos, frontSide bool, oldText, newText string)) {
	handle(abi.EventSignEdit, func(ev *Event) {
		d := ev.Decoder()
		pos := decodePos(d)
		frontSide := d.Bool()
		oldText := d.String()
		newText := d.String()
		fn(ev, pos, frontSide, oldText, newText)
	})
}

func registerSleep(fn func(ev *Event, sendReminder MutableArgument[bool])) {
	handle(abi.EventSleep, func(ev *Event) {
		_ = ev.Decoder().Bool()
		fn(ev, mutableBool(ev, ctxkey.SleepSendReminder))
	})
}

func registerLecternPageTurn(fn func(ev *Event, pos Pos, oldPage int32, newPage MutableArgument[int])) {
	handle(abi.EventLecternPageTurn, func(ev *Event) {
		d := ev.Decoder()
		pos := decodePos(d)
		oldPage := d.I32()
		_ = d.I32()
		fn(ev, pos, oldPage, mutableInt(ev, ctxkey.LecternNewPage))
	})
}

func registerItemDamage(fn func(ev *Event, item ItemStackData, damage MutableArgument[int])) {
	handle(abi.EventItemDamage, func(ev *Event) {
		d := ev.Decoder()
		item := decodeItemStack(d)
		_ = d.I32()
		fn(ev, item, mutableInt(ev, ctxkey.ItemDamageAmount))
	})
}

func registerItemPickup(fn func(ev *Event, item ItemStackData, count MutableArgument[int])) {
	handle(abi.EventItemPickup, func(ev *Event) {
		item := decodeItemStack(ev.Decoder())
		fn(ev, item, mutableInt(ev, ctxkey.ItemPickupCount))
	})
}

func registerHeldSlotChange(fn func(ev *Event, from, to int32)) {
	handle(abi.EventHeldSlotChange, func(ev *Event) {
		d := ev.Decoder()
		fn(ev, d.I32(), d.I32())
	})
}

func registerItemDrop(fn func(ev *Event, item ItemStackData)) {
	handle(abi.EventItemDrop, func(ev *Event) {
		fn(ev, decodeItemStack(ev.Decoder()))
	})
}

func registerTransfer(fn func(ev *Event, addr MutableArgument[string])) {
	handle(abi.EventTransfer, func(ev *Event) {
		_ = ev.Decoder().String()
		fn(ev, mutableString(ev, ctxkey.TransferAddr))
	})
}

func registerCommandExecution(fn func(ev *Event, command CommandData)) {
	handle(abi.EventCommandExecution, func(ev *Event) {
		fn(ev, decodeCommand(ev.Decoder()))
	})
}

func registerQuit(fn func(ev *Event, info PlayerIdentity)) {
	handle(abi.EventQuit, func(ev *Event) {
		fn(ev, decodePlayerIdentity(ev.Decoder()))
	})
}

func registerDiagnostics(fn func(ev *Event, d DiagnosticsData)) {
	handle(abi.EventDiagnostics, func(ev *Event) {
		fn(ev, decodeDiagnostics(ev.Decoder()))
	})
}

type eventChoice struct {
	name     string
	hints    []string
	register func()
}

func registerByName(fn any, choices ...eventChoice) {
	name := normalizedHandlerName(fn)
	for _, choice := range choices {
		for _, hint := range choice.hints {
			if strings.Contains(name, hint) {
				choice.register()
				return
			}
		}
	}
	options := make([]string, len(choices))
	for i, choice := range choices {
		options[i] = choice.name
	}
	panic(fmt.Sprintf("guest.Base.HandleEvent: ambiguous signature %T for handler %q; expected one of %s", fn, shortHandlerName(fn), strings.Join(options, ", ")))
}

func normalizedHandlerName(fn any) string {
	name := shortHandlerName(fn)
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func shortHandlerName(fn any) string {
	v := reflect.ValueOf(fn)
	if !v.IsValid() || v.Kind() != reflect.Func {
		return "<non-func>"
	}
	f := runtime.FuncForPC(v.Pointer())
	if f == nil {
		return "<unknown>"
	}
	name := f.Name()
	if i := strings.LastIndex(name, "."); i >= 0 && i+1 < len(name) {
		return name[i+1:]
	}
	return name
}
