package guest

import (
	"github.com/bedrock-gophers/plugin/plugin/abi"
)

type Vec3 struct {
	X float64
	Y float64
	Z float64
}

type Rotation struct {
	Yaw   float64
	Pitch float64
}

type Pos struct {
	X int32
	Y int32
	Z int32
}

type WorldData struct {
	Present   bool
	Name      string
	Dimension string
}

type BlockData struct {
	Present    bool
	Name       string
	Properties map[string]string
}

type EntityData struct {
	Present   bool
	HasHandle bool
	UUID      string
	Type      string
	Position  Vec3
	Rotation  Rotation
}

type ItemStackData struct {
	Count      int32
	HasItem    bool
	ItemName   string
	ItemMeta   int32
	CustomName string
}

type DamageSourceData struct {
	Present             bool
	Type                string
	ReducedByArmour     bool
	ReducedByResistance bool
	Fire                bool
	IgnoreTotem         bool
}

type HealingSourceData struct {
	Present bool
	Type    string
}

type SkinData struct {
	Present   bool
	Width     int32
	Height    int32
	Persona   bool
	PlayFabID string
	FullID    string
}

type CommandData struct {
	Name        string
	Description string
	Usage       string
	Args        []string
}

type DiagnosticsData struct {
	AverageFramesPerSecond        float64
	AverageServerSimTickTime      float64
	AverageClientSimTickTime      float64
	AverageBeginFrameTime         float64
	AverageInputTime              float64
	AverageRenderTime             float64
	AverageEndFrameTime           float64
	AverageRemainderTimePercent   float64
	AverageUnaccountedTimePercent float64
}

type PlayerIdentity struct {
	Name string
	UUID string
	XUID string
}

type MutableArgument[T any] struct {
	get func() T
	set func(T)
}

func (m MutableArgument[T]) Get() T {
	if m.get == nil {
		var zero T
		return zero
	}
	return m.get()
}

func (m MutableArgument[T]) Set(v T) {
	if m.set != nil {
		m.set(v)
	}
}

type MutableVec3 struct {
	X MutableArgument[float64]
	Y MutableArgument[float64]
	Z MutableArgument[float64]
}

func (m MutableVec3) Get() Vec3 {
	return Vec3{X: m.X.Get(), Y: m.Y.Get(), Z: m.Z.Get()}
}

func (m MutableVec3) Set(v Vec3) {
	m.X.Set(v.X)
	m.Y.Set(v.Y)
	m.Z.Set(v.Z)
}

type MutableItemStack struct {
	Count      MutableArgument[int]
	HasItem    MutableArgument[bool]
	ItemName   MutableArgument[string]
	ItemMeta   MutableArgument[int]
	CustomName MutableArgument[string]
}

func (m MutableItemStack) Get() ItemStackData {
	return ItemStackData{
		Count:      int32(m.Count.Get()),
		HasItem:    m.HasItem.Get(),
		ItemName:   m.ItemName.Get(),
		ItemMeta:   int32(m.ItemMeta.Get()),
		CustomName: m.CustomName.Get(),
	}
}

func (m MutableItemStack) Set(v ItemStackData) {
	m.Count.Set(int(v.Count))
	m.HasItem.Set(v.HasItem)
	m.ItemName.Set(v.ItemName)
	m.ItemMeta.Set(int(v.ItemMeta))
	m.CustomName.Set(v.CustomName)
}

func mutableString(ev *Event, slot uint32) MutableArgument[string] {
	return MutableArgument[string]{
		get: func() string { return ev.getString(slot) },
		set: func(v string) { ev.setString(slot, v) },
	}
}

func mutableBool(ev *Event, slot uint32) MutableArgument[bool] {
	return MutableArgument[bool]{
		get: func() bool { return ev.getBool(slot) },
		set: func(v bool) { ev.setBool(slot, v) },
	}
}

func mutableInt(ev *Event, slot uint32) MutableArgument[int] {
	return MutableArgument[int]{
		get: func() int { return int(ev.getI64(slot)) },
		set: func(v int) { ev.setI64(slot, int64(v)) },
	}
}

func mutableInt64(ev *Event, slot uint32) MutableArgument[int64] {
	return MutableArgument[int64]{
		get: func() int64 { return ev.getI64(slot) },
		set: func(v int64) { ev.setI64(slot, v) },
	}
}

func mutableFloat64(ev *Event, slot uint32) MutableArgument[float64] {
	return MutableArgument[float64]{
		get: func() float64 { return ev.getF64(slot) },
		set: func(v float64) { ev.setF64(slot, v) },
	}
}

func decodeVec3(d *abi.Decoder) Vec3 {
	return Vec3{X: d.F64(), Y: d.F64(), Z: d.F64()}
}

func decodeRotation(d *abi.Decoder) Rotation {
	return Rotation{Yaw: d.F64(), Pitch: d.F64()}
}

func decodePos(d *abi.Decoder) Pos {
	return Pos{X: d.I32(), Y: d.I32(), Z: d.I32()}
}

func decodeWorld(d *abi.Decoder) WorldData {
	if !d.Bool() {
		return WorldData{}
	}
	return WorldData{Present: true, Name: d.String(), Dimension: d.String()}
}

func decodeStringMap(d *abi.Decoder) map[string]string {
	n := int(d.U32())
	if n <= 0 {
		return map[string]string{}
	}
	m := make(map[string]string, n)
	for range n {
		k := d.String()
		v := d.String()
		m[k] = v
	}
	return m
}

func decodeBlock(d *abi.Decoder) BlockData {
	if !d.Bool() {
		return BlockData{}
	}
	return BlockData{Present: true, Name: d.String(), Properties: decodeStringMap(d)}
}

func decodeEntity(d *abi.Decoder) EntityData {
	if !d.Bool() {
		return EntityData{}
	}
	hasHandle := d.Bool()
	ent := EntityData{Present: true, HasHandle: hasHandle}
	if hasHandle {
		ent.UUID = d.String()
		ent.Type = d.String()
	}
	ent.Position = decodeVec3(d)
	ent.Rotation = decodeRotation(d)
	return ent
}

func decodeItemStack(d *abi.Decoder) ItemStackData {
	count := d.I32()
	hasItem := d.Bool()
	name := d.String()
	meta := d.I32()
	custom := d.String()
	return ItemStackData{Count: count, HasItem: hasItem, ItemName: name, ItemMeta: meta, CustomName: custom}
}

func decodeItemStacks(d *abi.Decoder) []ItemStackData {
	n := int(d.U32())
	if n <= 0 {
		return nil
	}
	items := make([]ItemStackData, 0, n)
	for range n {
		items = append(items, decodeItemStack(d))
	}
	return items
}

func decodeDamageSource(d *abi.Decoder) DamageSourceData {
	if !d.Bool() {
		return DamageSourceData{}
	}
	return DamageSourceData{
		Present:             true,
		Type:                d.String(),
		ReducedByArmour:     d.Bool(),
		ReducedByResistance: d.Bool(),
		Fire:                d.Bool(),
		IgnoreTotem:         d.Bool(),
	}
}

func decodeHealingSource(d *abi.Decoder) HealingSourceData {
	if !d.Bool() {
		return HealingSourceData{}
	}
	return HealingSourceData{Present: true, Type: d.String()}
}

func decodeSkin(d *abi.Decoder) SkinData {
	if !d.Bool() {
		return SkinData{}
	}
	return SkinData{
		Present:   true,
		Width:     d.I32(),
		Height:    d.I32(),
		Persona:   d.Bool(),
		PlayFabID: d.String(),
		FullID:    d.String(),
	}
}

func decodeCommand(d *abi.Decoder) CommandData {
	cmd := CommandData{Name: d.String(), Description: d.String(), Usage: d.String()}
	n := int(d.U32())
	if n > 0 {
		cmd.Args = make([]string, 0, n)
		for range n {
			cmd.Args = append(cmd.Args, d.String())
		}
	}
	return cmd
}

func decodePlayerIdentity(d *abi.Decoder) PlayerIdentity {
	return PlayerIdentity{Name: d.String(), UUID: d.String(), XUID: d.String()}
}

func decodeDiagnostics(d *abi.Decoder) DiagnosticsData {
	return DiagnosticsData{
		AverageFramesPerSecond:        d.F64(),
		AverageServerSimTickTime:      d.F64(),
		AverageClientSimTickTime:      d.F64(),
		AverageBeginFrameTime:         d.F64(),
		AverageInputTime:              d.F64(),
		AverageRenderTime:             d.F64(),
		AverageEndFrameTime:           d.F64(),
		AverageRemainderTimePercent:   d.F64(),
		AverageUnaccountedTimePercent: d.F64(),
	}
}
