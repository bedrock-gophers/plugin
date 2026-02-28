package guest

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/bedrock-gophers/plugin/plugin/abi"
)

var errHostUnavailable = errors.New("plugin host is not configured")

// MutableState exposes mutable event state for plugin event dispatch.
type MutableState interface {
	Cancel()
	GetI64(slot uint32) (int64, bool)
	SetI64(slot uint32, value int64)
	GetF64(slot uint32) (float64, bool)
	SetF64(slot uint32, value float64)
	GetBool(slot uint32) (bool, bool)
	SetBool(slot uint32, value bool)
	GetString(slot uint32) (string, bool)
	SetString(slot uint32, value string)
}

// Host is implemented by the host manager to back plugin APIs.
type Host interface {
	RegisterCommand(pluginName, name, description string, aliases []string, handlerID uint32, overloads []CommandOverloadSpec) bool
	ManagePlugins(action uint32, target string) ([]string, error)
	ResolvePlayerByName(name string) uint64
	OnlinePlayerNames() []string
	playerHost
	ConsoleMessage(pluginName, message string)
}

var (
	hostMu sync.RWMutex

	host Host

	handlersByPlugin = map[string]map[uint16]func(*Event){}
	globalHandlers   = map[uint16]func(*Event){}

	currentRegistrationPlugin string
)

var (
	sharedMu   sync.RWMutex
	sharedNext uint64 = 1
	shared            = map[uint64][]byte{}
)

// SetHost configures the host callback bridge used by plugins.
func SetHost(h Host) {
	hostMu.Lock()
	host = h
	hostMu.Unlock()
}

// BeginPluginRegistration marks the active plugin registration scope.
func BeginPluginRegistration(name string) {
	hostMu.Lock()
	currentRegistrationPlugin = name
	if _, ok := handlersByPlugin[name]; !ok {
		handlersByPlugin[name] = map[uint16]func(*Event){}
	}
	hostMu.Unlock()
}

// EndPluginRegistration clears the active plugin registration scope.
func EndPluginRegistration() {
	hostMu.Lock()
	currentRegistrationPlugin = ""
	hostMu.Unlock()
}

func currentPluginRegistration() string {
	hostMu.RLock()
	name := currentRegistrationPlugin
	hostMu.RUnlock()
	return name
}

func currentHost() Host {
	hostMu.RLock()
	h := host
	hostMu.RUnlock()
	return h
}

func handle(eventID uint16, fn func(*Event)) {
	if fn == nil {
		return
	}
	hostMu.Lock()
	pluginName := currentRegistrationPlugin
	if pluginName == "" {
		globalHandlers[eventID] = fn
		hostMu.Unlock()
		return
	}
	byEvent := handlersByPlugin[pluginName]
	if byEvent == nil {
		byEvent = map[uint16]func(*Event){}
		handlersByPlugin[pluginName] = byEvent
	}
	byEvent[eventID] = fn
	hostMu.Unlock()
}

// Event is the current event being dispatched to the plugin.
type Event struct {
	Descriptor abi.EventDescriptor
	Player     PlayerRef

	pluginName string
	payload    []byte
	mutable    MutableState
}

// DispatchEvent dispatches one event to a plugin registered through BeginPluginRegistration.
func DispatchEvent(pluginName string, desc abi.EventDescriptor, payload []byte, mutable MutableState) {
	hostMu.RLock()
	byEvent := handlersByPlugin[pluginName]
	fn := byEvent[desc.EventID]
	if fn == nil {
		fn = globalHandlers[desc.EventID]
	}
	hostMu.RUnlock()
	if fn == nil {
		return
	}

	ev := &Event{
		Descriptor: desc,
		Player:     PlayerRef{id: desc.PlayerID},
		pluginName: pluginName,
		payload:    append([]byte(nil), payload...),
		mutable:    mutable,
	}
	defer func() {
		if r := recover(); r != nil {
			slog.Error("plugin event panicked", "plugin", pluginName, "event", abi.EventName(desc.EventID), "panic", r)
		}
	}()
	fn(ev)
}

// InvokeCommand dispatches a plugin-command event directly.
func InvokeCommand(pluginName string, playerID uint64, handlerID uint32, args []string) {
	enc := abi.NewEncoder(128)
	enc.U32(handlerID)
	enc.U32(uint32(len(args)))
	for _, arg := range args {
		enc.String(arg)
	}
	DispatchEvent(pluginName, abi.EventDescriptor{
		Version:  abi.Version,
		EventID:  abi.EventPluginCommand,
		Flags:    abi.FlagSynchronous,
		PlayerID: playerID,
	}, enc.Data(), nil)
}

func (e *Event) EventID() uint16 {
	return e.Descriptor.EventID
}

func (e *Event) PlayerID() uint64 {
	return e.Descriptor.PlayerID
}

func (e *Event) Flags() uint32 {
	return e.Descriptor.Flags
}

func (e *Event) PluginName() string {
	return e.pluginName
}

func (e *Event) Payload() []byte {
	return append([]byte(nil), e.payload...)
}

func (e *Event) Decoder() *abi.Decoder {
	return abi.NewDecoder(e.payload)
}

func (e *Event) Cancel() {
	if e.mutable != nil {
		e.mutable.Cancel()
	}
}

func (e *Event) getI64(slot uint32) int64 {
	if e.mutable == nil {
		return 0
	}
	v, _ := e.mutable.GetI64(slot)
	return v
}

func (e *Event) setI64(slot uint32, value int64) {
	if e.mutable == nil {
		return
	}
	e.mutable.SetI64(slot, value)
}

func (e *Event) getF64(slot uint32) float64 {
	if e.mutable == nil {
		return 0
	}
	v, _ := e.mutable.GetF64(slot)
	return v
}

func (e *Event) setF64(slot uint32, value float64) {
	if e.mutable == nil {
		return
	}
	e.mutable.SetF64(slot, value)
}

func (e *Event) getBool(slot uint32) bool {
	if e.mutable == nil {
		return false
	}
	v, _ := e.mutable.GetBool(slot)
	return v
}

func (e *Event) setBool(slot uint32, value bool) {
	if e.mutable == nil {
		return
	}
	e.mutable.SetBool(slot, value)
}

func (e *Event) getString(slot uint32) string {
	if e.mutable == nil {
		return ""
	}
	v, _ := e.mutable.GetString(slot)
	return v
}

func (e *Event) setString(slot uint32, value string) {
	if e.mutable == nil {
		return
	}
	e.mutable.SetString(slot, value)
}

func AllocShared() uint64 {
	sharedMu.Lock()
	sharedNext++
	key := sharedNext
	shared[key] = nil
	sharedMu.Unlock()
	return key
}

func ReadShared(key uint64) []byte {
	sharedMu.RLock()
	v := append([]byte(nil), shared[key]...)
	sharedMu.RUnlock()
	return v
}

func WriteShared(key uint64, value []byte) uint32 {
	sharedMu.Lock()
	shared[key] = append([]byte(nil), value...)
	sharedMu.Unlock()
	return uint32(len(value))
}

func DeleteShared(key uint64) {
	sharedMu.Lock()
	delete(shared, key)
	sharedMu.Unlock()
}

func hostValue[T any](fallback T, fn func(Host) T) T {
	h := currentHost()
	if h == nil {
		return fallback
	}
	return fn(h)
}

func hostBool(fn func(Host) bool) bool {
	return hostValue(false, fn)
}

func hostDo(fn func(Host)) {
	h := currentHost()
	if h == nil {
		return
	}
	fn(h)
}

func consoleMessageForPlugin(pluginName, message string) {
	h := currentHost()
	if h == nil {
		slog.Info(message)
		return
	}
	h.ConsoleMessage(pluginName, message)
}
