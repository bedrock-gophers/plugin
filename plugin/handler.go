package plugin

import (
	"net"
	"time"

	"github.com/bedrock-gophers/plugin/plugin/abi"
	"github.com/bedrock-gophers/plugin/plugin/internal/ctxkey"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type Handler struct {
	player.NopHandler
	m *Manager
}

func (h *Handler) HandleMove(ctx *player.Context, newPos mgl64.Vec3, newRot cube.Rotation) {
	enc := newPayload(64)
	encodeVec3(enc, newPos)
	encodeRotation(enc, newRot)
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventMove, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleJump(p *player.Player) {
	h.m.dispatch(p, abi.EventJump, 0, payloadPlayerIdentity(p), nil)
}

func (h *Handler) HandleTeleport(ctx *player.Context, pos mgl64.Vec3) {
	enc := newPayload(48)
	encodeVec3(enc, pos)
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventTeleport, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleChangeWorld(p *player.Player, before, after *world.World) {
	enc := newPayload(128)
	encodeWorld(enc, before)
	encodeWorld(enc, after)
	h.m.dispatch(p, abi.EventChangeWorld, 0, enc.Data(), nil)
}

func (h *Handler) HandleToggleSprint(ctx *player.Context, after bool) {
	enc := newPayload(8)
	enc.Bool(after)
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventToggleSprint, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleToggleSneak(ctx *player.Context, after bool) {
	enc := newPayload(8)
	enc.Bool(after)
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventToggleSneak, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleChat(ctx *player.Context, message *string) {
	original := *message
	enc := newPayload(64)
	enc.String(original)
	mutable := newMutableState(ctx.Cancel)
	mutable.AddString(ctxkey.ChatMessage, original, func(v string) { *message = v })
	h.m.dispatch(ctx.Val(), abi.EventChat, abi.FlagCancellable, enc.Data(), mutable)
	if *message != original {
		ctx.Cancel()
		_, _ = chat.Global.WriteString(*message)
	}
}

func (h *Handler) HandleFoodLoss(ctx *player.Context, from int, to *int) {
	enc := newPayload(16)
	enc.I32(int32(from))
	enc.I32(int32(*to))
	mutable := newMutableState(ctx.Cancel)
	mutable.AddI64(ctxkey.FoodTo, int64(*to), func(v int64) { *to = int(v) })
	h.m.dispatch(ctx.Val(), abi.EventFoodLoss, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleHeal(ctx *player.Context, health *float64, src world.HealingSource) {
	if _, ok := src.(pluginHealingSource); ok {
		return
	}
	enc := newPayload(64)
	enc.F64(*health)
	encodeHealingSource(enc, src)
	mutable := newMutableState(ctx.Cancel)
	mutable.AddF64(ctxkey.HealAmount, *health, func(v float64) { *health = v })
	h.m.dispatch(ctx.Val(), abi.EventHeal, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleHurt(ctx *player.Context, damage *float64, immune bool, attackImmunity *time.Duration, src world.DamageSource) {
	if _, ok := src.(pluginDamageSource); ok {
		return
	}
	enc := newPayload(96)
	enc.F64(*damage)
	enc.Bool(immune)
	enc.I64(int64(*attackImmunity / time.Millisecond))
	encodeDamageSource(enc, src)
	mutable := newMutableState(ctx.Cancel)
	mutable.AddF64(ctxkey.HurtDamage, *damage, func(v float64) { *damage = v })
	mutable.AddI64(ctxkey.HurtAttackImmunityMillis, int64(*attackImmunity/time.Millisecond), func(v int64) {
		*attackImmunity = time.Duration(v) * time.Millisecond
	})
	h.m.dispatch(ctx.Val(), abi.EventHurt, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleDeath(p *player.Player, src world.DamageSource, keepInv *bool) {
	enc := newPayload(64)
	encodeDamageSource(enc, src)
	enc.Bool(*keepInv)
	mutable := newMutableState(nil)
	mutable.AddBool(ctxkey.DeathKeepInventory, *keepInv, func(v bool) { *keepInv = v })
	h.m.dispatch(p, abi.EventDeath, 0, enc.Data(), mutable)
}

func (h *Handler) HandleRespawn(p *player.Player, pos *mgl64.Vec3, w **world.World) {
	enc := newPayload(96)
	encodeVec3(enc, *pos)
	if *w == nil {
		enc.String("")
	} else {
		enc.String((*w).Name())
	}
	mutable := newMutableState(nil)
	mutable.AddF64(ctxkey.RespawnPosX, (*pos)[0], func(v float64) { (*pos)[0] = v })
	mutable.AddF64(ctxkey.RespawnPosY, (*pos)[1], func(v float64) { (*pos)[1] = v })
	mutable.AddF64(ctxkey.RespawnPosZ, (*pos)[2], func(v float64) { (*pos)[2] = v })
	worldName := ""
	if *w != nil {
		worldName = (*w).Name()
	}
	mutable.AddString(ctxkey.RespawnWorldName, worldName, func(v string) {
		if resolved := h.m.resolveWorld(v); resolved != nil {
			*w = resolved
		}
	})
	h.m.dispatch(p, abi.EventRespawn, 0, enc.Data(), mutable)
}

func (h *Handler) HandleSkinChange(ctx *player.Context, s *skin.Skin) {
	enc := newPayload(128)
	encodeSkin(enc, s)
	mutable := newMutableState(ctx.Cancel)
	if s != nil {
		mutable.AddString(ctxkey.ChatMessage, s.FullID, func(v string) { s.FullID = v })
	}
	h.m.dispatch(ctx.Val(), abi.EventSkinChange, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleFireExtinguish(ctx *player.Context, pos cube.Pos) {
	enc := newPayload(24)
	encodePos(enc, pos)
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventFireExtinguish, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleStartBreak(ctx *player.Context, pos cube.Pos) {
	enc := newPayload(24)
	encodePos(enc, pos)
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventStartBreak, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleBlockBreak(ctx *player.Context, pos cube.Pos, drops *[]item.Stack, xp *int) {
	enc := newPayload(128)
	encodePos(enc, pos)
	encodeItemStacks(enc, *drops)
	enc.I32(int32(*xp))
	mutable := newMutableState(ctx.Cancel)
	mutable.AddI64(ctxkey.BlockBreakXP, int64(*xp), func(v int64) { *xp = int(v) })
	h.m.dispatch(ctx.Val(), abi.EventBlockBreak, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleBlockPlace(ctx *player.Context, pos cube.Pos, b world.Block) {
	enc := newPayload(128)
	encodePos(enc, pos)
	encodeBlock(enc, b)
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventBlockPlace, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleBlockPick(ctx *player.Context, pos cube.Pos, b world.Block) {
	enc := newPayload(128)
	encodePos(enc, pos)
	encodeBlock(enc, b)
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventBlockPick, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleItemUse(ctx *player.Context) {
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventItemUse, abi.FlagCancellable, nil, mutable)
}

func (h *Handler) HandleItemUseOnBlock(ctx *player.Context, pos cube.Pos, face cube.Face, clickPos mgl64.Vec3) {
	enc := newPayload(72)
	encodePos(enc, pos)
	enc.U8(uint8(face))
	encodeVec3(enc, clickPos)
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventItemUseOnBlock, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleItemUseOnEntity(ctx *player.Context, e world.Entity) {
	enc := newPayload(128)
	encodeEntity(enc, e)
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventItemUseOnEntity, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleItemRelease(ctx *player.Context, i item.Stack, dur time.Duration) {
	enc := newPayload(96)
	encodeItemStack(enc, i)
	enc.I64(int64(dur / time.Millisecond))
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventItemRelease, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleItemConsume(ctx *player.Context, i item.Stack) {
	enc := newPayload(64)
	encodeItemStack(enc, i)
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventItemConsume, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleAttackEntity(ctx *player.Context, e world.Entity, force, height *float64, critical *bool) {
	enc := newPayload(128)
	encodeEntity(enc, e)
	enc.F64(*force)
	enc.F64(*height)
	enc.Bool(*critical)
	mutable := newMutableState(ctx.Cancel)
	mutable.AddF64(ctxkey.AttackForce, *force, func(v float64) { *force = v })
	mutable.AddF64(ctxkey.AttackHeight, *height, func(v float64) { *height = v })
	mutable.AddBool(ctxkey.AttackCritical, *critical, func(v bool) { *critical = v })
	h.m.dispatch(ctx.Val(), abi.EventAttackEntity, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleExperienceGain(ctx *player.Context, amount *int) {
	enc := newPayload(16)
	enc.I32(int32(*amount))
	mutable := newMutableState(ctx.Cancel)
	mutable.AddI64(ctxkey.ExperienceAmount, int64(*amount), func(v int64) { *amount = int(v) })
	h.m.dispatch(ctx.Val(), abi.EventExperienceGain, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandlePunchAir(ctx *player.Context) {
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventPunchAir, abi.FlagCancellable, nil, mutable)
}

func (h *Handler) HandleSignEdit(ctx *player.Context, pos cube.Pos, frontSide bool, oldText, newText string) {
	enc := newPayload(128)
	encodePos(enc, pos)
	enc.Bool(frontSide)
	enc.String(oldText)
	enc.String(newText)
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventSignEdit, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleSleep(ctx *player.Context, sendReminder *bool) {
	enc := newPayload(8)
	enc.Bool(*sendReminder)
	mutable := newMutableState(ctx.Cancel)
	mutable.AddBool(ctxkey.SleepSendReminder, *sendReminder, func(v bool) { *sendReminder = v })
	h.m.dispatch(ctx.Val(), abi.EventSleep, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleLecternPageTurn(ctx *player.Context, pos cube.Pos, oldPage int, newPage *int) {
	enc := newPayload(32)
	encodePos(enc, pos)
	enc.I32(int32(oldPage))
	enc.I32(int32(*newPage))
	mutable := newMutableState(ctx.Cancel)
	mutable.AddI64(ctxkey.LecternNewPage, int64(*newPage), func(v int64) { *newPage = int(v) })
	h.m.dispatch(ctx.Val(), abi.EventLecternPageTurn, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleItemDamage(ctx *player.Context, i item.Stack, damage *int) {
	enc := newPayload(96)
	encodeItemStack(enc, i)
	enc.I32(int32(*damage))
	mutable := newMutableState(ctx.Cancel)
	mutable.AddI64(ctxkey.ItemDamageAmount, int64(*damage), func(v int64) { *damage = int(v) })
	h.m.dispatch(ctx.Val(), abi.EventItemDamage, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleItemPickup(ctx *player.Context, i *item.Stack) {
	enc := newPayload(96)
	encodeItemStack(enc, *i)
	mutable := newMutableState(ctx.Cancel)
	mutable.AddI64(ctxkey.ItemPickupCount, int64(i.Count()), func(v int64) {
		*i = i.Grow(int(v) - i.Count())
	})
	h.m.dispatch(ctx.Val(), abi.EventItemPickup, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleHeldSlotChange(ctx *player.Context, from, to int) {
	enc := newPayload(16)
	enc.I32(int32(from))
	enc.I32(int32(to))
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventHeldSlotChange, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleItemDrop(ctx *player.Context, s item.Stack) {
	enc := newPayload(96)
	encodeItemStack(enc, s)
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventItemDrop, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleTransfer(ctx *player.Context, addr *net.UDPAddr) {
	enc := newPayload(64)
	encoded := encodeAddr(addr)
	enc.String(encoded)
	mutable := newMutableState(ctx.Cancel)
	mutable.AddString(ctxkey.TransferAddr, encoded, func(v string) {
		a, err := net.ResolveUDPAddr("udp", v)
		if err == nil {
			*addr = *a
		}
	})
	h.m.dispatch(ctx.Val(), abi.EventTransfer, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleCommandExecution(ctx *player.Context, command cmd.Command, args []string) {
	enc := newPayload(256)
	encodeCommand(enc, command, args)
	mutable := newMutableState(ctx.Cancel)
	h.m.dispatch(ctx.Val(), abi.EventCommandExecution, abi.FlagCancellable, enc.Data(), mutable)
}

func (h *Handler) HandleQuit(p *player.Player) {
	h.m.dispatch(p, abi.EventQuit, 0, payloadPlayerIdentity(p), nil)
	h.m.players.remove(p)
}

func (h *Handler) HandleDiagnostics(p *player.Player, d session.Diagnostics) {
	enc := newPayload(128)
	encodeDiagnostics(enc, d)
	h.m.dispatch(p, abi.EventDiagnostics, 0, enc.Data(), nil)
}
