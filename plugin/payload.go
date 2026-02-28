package plugin

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"sort"

	"github.com/bedrock-gophers/plugin/plugin/abi"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func newPayload(capHint int) *abi.Encoder {
	return abi.NewEncoder(capHint)
}

func encodeVec3(enc *abi.Encoder, v mgl64.Vec3) {
	enc.F64(v[0])
	enc.F64(v[1])
	enc.F64(v[2])
}

func encodeRotation(enc *abi.Encoder, r cube.Rotation) {
	enc.F64(r.Yaw())
	enc.F64(r.Pitch())
}

func encodePos(enc *abi.Encoder, p cube.Pos) {
	enc.I32(int32(p.X()))
	enc.I32(int32(p.Y()))
	enc.I32(int32(p.Z()))
}

func encodeWorld(enc *abi.Encoder, w *world.World) {
	if w == nil {
		enc.Bool(false)
		return
	}
	enc.Bool(true)
	enc.String(w.Name())
	enc.String(fmt.Sprint(w.Dimension()))
}

func encodeBlock(enc *abi.Encoder, b world.Block) {
	if b == nil {
		enc.Bool(false)
		return
	}
	enc.Bool(true)
	name, props := b.EncodeBlock()
	enc.String(name)
	encodeStringMap(enc, props)
}

func encodeEntity(enc *abi.Encoder, e world.Entity) {
	if e == nil {
		enc.Bool(false)
		return
	}
	enc.Bool(true)
	h := e.H()
	if h == nil {
		enc.Bool(false)
		enc.String("")
		enc.String("")
	} else {
		enc.Bool(true)
		enc.String(h.UUID().String())
		enc.String(h.Type().EncodeEntity())
	}
	encodeVec3(enc, e.Position())
	encodeRotation(enc, e.Rotation())
}

func encodeItemStack(enc *abi.Encoder, s item.Stack) {
	enc.I32(int32(s.Count()))
	it := s.Item()
	if it == nil {
		enc.Bool(false)
		enc.String("")
		enc.I32(0)
	} else {
		enc.Bool(true)
		name, meta := it.EncodeItem()
		enc.String(name)
		enc.I32(int32(meta))
	}
	enc.String(s.CustomName())
}

func encodeItemStacks(enc *abi.Encoder, stacks []item.Stack) {
	enc.U32(uint32(len(stacks)))
	for _, s := range stacks {
		encodeItemStack(enc, s)
	}
}

func encodeDamageSource(enc *abi.Encoder, src world.DamageSource) {
	if src == nil {
		enc.Bool(false)
		return
	}
	enc.Bool(true)
	enc.String(reflect.TypeOf(src).String())
	enc.Bool(src.ReducedByArmour())
	enc.Bool(src.ReducedByResistance())
	enc.Bool(src.Fire())
	enc.Bool(src.IgnoreTotem())
}

func encodeHealingSource(enc *abi.Encoder, src world.HealingSource) {
	if src == nil {
		enc.Bool(false)
		return
	}
	enc.Bool(true)
	enc.String(reflect.TypeOf(src).String())
}

func encodeCommand(enc *abi.Encoder, command cmd.Command, args []string) {
	enc.String(command.Name())
	enc.String(command.Description())
	enc.String(command.Usage())
	enc.U32(uint32(len(args)))
	for _, arg := range args {
		enc.String(arg)
	}
}

func encodeSkin(enc *abi.Encoder, s *skin.Skin) {
	if s == nil {
		enc.Bool(false)
		return
	}
	enc.Bool(true)
	bounds := s.Bounds()
	enc.I32(int32(bounds.Dx()))
	enc.I32(int32(bounds.Dy()))
	enc.Bool(s.Persona)
	enc.String(s.PlayFabID)
	enc.String(s.FullID)
}

func encodeDiagnostics(enc *abi.Encoder, d session.Diagnostics) {
	enc.F64(d.AverageFramesPerSecond)
	enc.F64(d.AverageServerSimTickTime)
	enc.F64(d.AverageClientSimTickTime)
	enc.F64(d.AverageBeginFrameTime)
	enc.F64(d.AverageInputTime)
	enc.F64(d.AverageRenderTime)
	enc.F64(d.AverageEndFrameTime)
	enc.F64(d.AverageRemainderTimePercent)
	enc.F64(d.AverageUnaccountedTimePercent)
}

func encodeStringMap(enc *abi.Encoder, values map[string]any) {
	if len(values) == 0 {
		enc.U32(0)
		return
	}
	keys := slices.Collect(maps.Keys(values))
	sort.Strings(keys)
	enc.U32(uint32(len(keys)))
	for _, key := range keys {
		enc.String(key)
		enc.String(fmt.Sprint(values[key]))
	}
}

func payloadPlayerIdentity(p *player.Player) []byte {
	return payloadPlayerIdentityValues(p.Name(), p.UUID().String(), p.XUID())
}

func payloadPlayerIdentityValues(name, uuid, xuid string) []byte {
	enc := newPayload(96)
	enc.String(name)
	enc.String(uuid)
	enc.String(xuid)
	return enc.Data()
}
