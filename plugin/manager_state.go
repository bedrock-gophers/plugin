package plugin

import (
	"net"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/bedrock-gophers/plugin/plugin/abi"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

type pluginCommandRegistration struct {
	plugin      *pluginRuntime
	name        string
	description string
	aliases     []string
	handlerID   uint32
	overloads   []pluginCommandOverloadSpec
}

type pluginCommandBlueprint struct {
	name        string
	description string
	aliases     []string
	handlerID   uint32
	overloads   []pluginCommandOverloadSpec
}

type pluginKind uint8

const (
	pluginKindGo pluginKind = iota
	pluginKindCSharp
	pluginKindRust
)

type pluginRuntime struct {
	name     string
	path     string
	kind     pluginKind
	loaded   bool
	onUnload func()
	runtime  *csharpRuntime

	mu       sync.RWMutex
	commands map[string]pluginCommandBlueprint
}

func newPluginRuntime(name, path string, kind pluginKind) *pluginRuntime {
	return &pluginRuntime{
		name:     name,
		path:     path,
		kind:     kind,
		loaded:   true,
		commands: map[string]pluginCommandBlueprint{},
	}
}

func (p *pluginRuntime) setCommandBlueprint(blueprint pluginCommandBlueprint) {
	p.mu.Lock()
	if p.commands == nil {
		p.commands = map[string]pluginCommandBlueprint{}
	}
	p.commands[blueprint.name] = cloneCommandBlueprint(blueprint)
	p.mu.Unlock()
}

func (p *pluginRuntime) commandBlueprintsSnapshot() []pluginCommandBlueprint {
	p.mu.RLock()
	out := make([]pluginCommandBlueprint, 0, len(p.commands))
	for _, blueprint := range p.commands {
		out = append(out, cloneCommandBlueprint(blueprint))
	}
	p.mu.RUnlock()
	return out
}

func cloneCommandBlueprint(blueprint pluginCommandBlueprint) pluginCommandBlueprint {
	cloned := pluginCommandBlueprint{
		name:        blueprint.name,
		description: blueprint.description,
		handlerID:   blueprint.handlerID,
		aliases:     append([]string(nil), blueprint.aliases...),
		overloads:   make([]pluginCommandOverloadSpec, 0, len(blueprint.overloads)),
	}
	for _, overload := range blueprint.overloads {
		params := make([]pluginCommandParamSpec, 0, len(overload.parameters))
		for _, param := range overload.parameters {
			params = append(params, pluginCommandParamSpec{
				name:        param.name,
				kind:        param.kind,
				optional:    param.optional,
				enumOptions: append([]string(nil), param.enumOptions...),
			})
		}
		cloned.overloads = append(cloned.overloads, pluginCommandOverloadSpec{parameters: params})
	}
	return cloned
}

type playerStore struct {
	next atomic.Uint64

	mu          sync.RWMutex
	byHandle    map[*world.EntityHandle]uint64
	playersByID map[uint64]*player.Player
}

func newPlayerStore() *playerStore {
	p := &playerStore{
		byHandle:    map[*world.EntityHandle]uint64{},
		playersByID: map[uint64]*player.Player{},
	}
	p.next.Store(1)
	return p
}

func (s *playerStore) ensure(p *player.Player) uint64 {
	h := p.H()
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.byHandle[h]; ok {
		s.playersByID[id] = p
		return id
	}
	id := s.next.Add(1)
	s.byHandle[h] = id
	s.playersByID[id] = p
	return id
}

func (s *playerStore) remove(p *player.Player) {
	h := p.H()
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byHandle[h]
	if !ok {
		return
	}
	delete(s.byHandle, h)
	delete(s.playersByID, id)
}

func (s *playerStore) byID(id uint64) (*player.Player, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.playersByID[id]
	return p, ok
}

func (s *playerStore) byName(name string) (uint64, *player.Player, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, nil, false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for id, p := range s.playersByID {
		if strings.EqualFold(p.Name(), name) {
			return id, p, true
		}
	}
	return 0, nil, false
}

func (s *playerStore) names() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.playersByID))
	for _, p := range s.playersByID {
		names = append(names, p.Name())
	}
	sort.Strings(names)
	return names
}

type mutableState struct {
	mu sync.RWMutex

	cancelRequested bool
	cancelFn        func()

	i64     map[uint32]int64
	i64Sink map[uint32]func(int64)

	f64     map[uint32]float64
	f64Sink map[uint32]func(float64)

	bools     map[uint32]bool
	boolsSink map[uint32]func(bool)

	strings     map[uint32]string
	stringsSink map[uint32]func(string)
}

func newMutableState(cancelFn func()) *mutableState {
	return &mutableState{
		cancelFn:    cancelFn,
		i64:         map[uint32]int64{},
		i64Sink:     map[uint32]func(int64){},
		f64:         map[uint32]float64{},
		f64Sink:     map[uint32]func(float64){},
		bools:       map[uint32]bool{},
		boolsSink:   map[uint32]func(bool){},
		strings:     map[uint32]string{},
		stringsSink: map[uint32]func(string){},
	}
}

func (m *mutableState) AddI64(slot uint32, value int64, sink func(int64)) {
	m.mu.Lock()
	m.i64[slot] = value
	m.i64Sink[slot] = sink
	m.mu.Unlock()
}

func (m *mutableState) AddF64(slot uint32, value float64, sink func(float64)) {
	m.mu.Lock()
	m.f64[slot] = value
	m.f64Sink[slot] = sink
	m.mu.Unlock()
}

func (m *mutableState) AddBool(slot uint32, value bool, sink func(bool)) {
	m.mu.Lock()
	m.bools[slot] = value
	m.boolsSink[slot] = sink
	m.mu.Unlock()
}

func (m *mutableState) AddString(slot uint32, value string, sink func(string)) {
	m.mu.Lock()
	m.strings[slot] = value
	m.stringsSink[slot] = sink
	m.mu.Unlock()
}

func (m *mutableState) Cancel() {
	m.mu.Lock()
	m.cancelRequested = true
	m.mu.Unlock()
}

func (m *mutableState) Cancelled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cancelRequested
}

func (m *mutableState) GetI64(slot uint32) (int64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.i64[slot]
	return v, ok
}

func (m *mutableState) SetI64(slot uint32, value int64) {
	m.mu.Lock()
	if _, ok := m.i64[slot]; ok {
		m.i64[slot] = value
	}
	m.mu.Unlock()
}

func (m *mutableState) GetF64(slot uint32) (float64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.f64[slot]
	return v, ok
}

func (m *mutableState) SetF64(slot uint32, value float64) {
	m.mu.Lock()
	if _, ok := m.f64[slot]; ok {
		m.f64[slot] = value
	}
	m.mu.Unlock()
}

func (m *mutableState) GetBool(slot uint32) (bool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.bools[slot]
	return v, ok
}

func (m *mutableState) SetBool(slot uint32, value bool) {
	m.mu.Lock()
	if _, ok := m.bools[slot]; ok {
		m.bools[slot] = value
	}
	m.mu.Unlock()
}

func (m *mutableState) GetString(slot uint32) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.strings[slot]
	return v, ok
}

func (m *mutableState) SetString(slot uint32, value string) {
	m.mu.Lock()
	if _, ok := m.strings[slot]; ok {
		m.strings[slot] = value
	}
	m.mu.Unlock()
}

func (m *mutableState) Apply() {
	m.mu.Lock()
	cancel := m.cancelRequested
	cancelFn := m.cancelFn

	i64Values := make(map[uint32]int64, len(m.i64))
	for k, v := range m.i64 {
		i64Values[k] = v
	}
	i64Sinks := make(map[uint32]func(int64), len(m.i64Sink))
	for k, v := range m.i64Sink {
		i64Sinks[k] = v
	}

	f64Values := make(map[uint32]float64, len(m.f64))
	for k, v := range m.f64 {
		f64Values[k] = v
	}
	f64Sinks := make(map[uint32]func(float64), len(m.f64Sink))
	for k, v := range m.f64Sink {
		f64Sinks[k] = v
	}

	boolValues := make(map[uint32]bool, len(m.bools))
	for k, v := range m.bools {
		boolValues[k] = v
	}
	boolSinks := make(map[uint32]func(bool), len(m.boolsSink))
	for k, v := range m.boolsSink {
		boolSinks[k] = v
	}

	stringValues := make(map[uint32]string, len(m.strings))
	for k, v := range m.strings {
		stringValues[k] = v
	}
	stringSinks := make(map[uint32]func(string), len(m.stringsSink))
	for k, v := range m.stringsSink {
		stringSinks[k] = v
	}
	m.mu.Unlock()

	for slot, sink := range i64Sinks {
		sink(i64Values[slot])
	}
	for slot, sink := range f64Sinks {
		sink(f64Values[slot])
	}
	for slot, sink := range boolSinks {
		sink(boolValues[slot])
	}
	for slot, sink := range stringSinks {
		sink(stringValues[slot])
	}
	if cancel && cancelFn != nil {
		cancelFn()
	}
}

type pluginDamageSource struct{}

func (pluginDamageSource) ReducedByArmour() bool     { return false }
func (pluginDamageSource) ReducedByResistance() bool { return false }
func (pluginDamageSource) Fire() bool                { return false }
func (pluginDamageSource) IgnoreTotem() bool         { return true }

type pluginHealingSource struct{}

func (pluginHealingSource) HealingSource() {}

func encodeAddr(addr *net.UDPAddr) string {
	if addr == nil {
		return ""
	}
	return addr.String()
}

func descriptorForDispatch(eventID uint16, flags uint32, playerID uint64) abi.EventDescriptor {
	return abi.EventDescriptor{
		Version:  abi.Version,
		EventID:  eventID,
		Flags:    flags | abi.FlagSynchronous,
		PlayerID: playerID,
	}
}
