//go:build cgo

package plugin

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdint.h>
#include <stdlib.h>

typedef struct {
	uintptr_t ctx;
	void* fns[74];
} native_host_api;

extern int host_register_command(uintptr_t ctx, char* plugin_name, char* name, char* description, char* aliases_csv, uint32_t handler_id, uint8_t* overloads, uint32_t overloads_len);
extern uint8_t* host_manage_plugins(uintptr_t ctx, uint32_t action, char* target, uint32_t* out_len);
extern uint64_t host_resolve_player_by_name(uintptr_t ctx, char* name);
extern uint64_t host_player_handle(uintptr_t ctx, uint64_t player_id);
extern uint8_t* host_online_player_names(uintptr_t ctx, uint32_t* out_len);
extern void host_console_message(uintptr_t ctx, char* plugin_name, char* message);
extern uint8_t* host_call(uintptr_t ctx, uint32_t op, uint8_t* payload, uint32_t payload_len, uint32_t* out_len);
extern int host_event_cancel(uintptr_t ctx, uint64_t request_key);
extern int host_event_i64_set(uintptr_t ctx, uint64_t request_key, uint32_t slot, int64_t value);
extern int host_event_f64_set(uintptr_t ctx, uint64_t request_key, uint32_t slot, double value);
extern int host_event_bool_set(uintptr_t ctx, uint64_t request_key, uint32_t slot, int value);
extern int host_event_string_set(uintptr_t ctx, uint64_t request_key, uint32_t slot, char* value);
extern int host_event_item_drop_set(uintptr_t ctx, uint64_t request_key, uint8_t* payload, uint32_t payload_len);
extern char* host_player_name(uintptr_t ctx, uint64_t player_id);
extern int64_t host_player_latency(uintptr_t ctx, uint64_t player_id);
extern void host_player_message(uintptr_t ctx, uint64_t player_id, char* message);
extern void host_free_string(char* value);
extern void host_free_bytes(uint8_t* value);

typedef int (*plugin_load_fn)(native_host_api* host, char* plugin_name);
typedef void (*plugin_dispatch_fn)(uint16_t version, uint16_t event_id, uint32_t flags, uint64_t player_id, uint64_t request_key, const uint8_t* payload, uint32_t payload_len);

static native_host_api make_host_api(uintptr_t ctx) {
	native_host_api api;
	api.ctx = ctx;
	for (int i = 0; i < 74; i++) {
		api.fns[i] = NULL;
	}
	api.fns[0] = (void*)host_register_command;
	api.fns[1] = (void*)host_manage_plugins;
	api.fns[2] = (void*)host_resolve_player_by_name;
	api.fns[3] = (void*)host_player_handle;
	api.fns[4] = (void*)host_online_player_names;
	api.fns[5] = (void*)host_console_message;
	api.fns[6] = (void*)host_call;
	api.fns[7] = (void*)host_event_cancel;
	api.fns[8] = (void*)host_event_i64_set;
	api.fns[9] = (void*)host_event_f64_set;
	api.fns[10] = (void*)host_event_bool_set;
	api.fns[11] = (void*)host_event_string_set;
	api.fns[12] = (void*)host_event_item_drop_set;
	api.fns[17] = (void*)host_player_name;
	api.fns[62] = (void*)host_player_latency;
	api.fns[71] = (void*)host_player_message;
	api.fns[72] = (void*)host_free_string;
	api.fns[73] = (void*)host_free_bytes;
	return api;
}

static void* native_open_library(char* path) { return dlopen(path, RTLD_NOW | RTLD_LOCAL); }
static char* native_last_error() { return dlerror(); }
static void* native_lookup_symbol(void* handle, char* symbol) { return dlsym(handle, symbol); }
static int native_call_plugin_load(void* fn, native_host_api* host, char* plugin_name) { return ((plugin_load_fn) fn)(host, plugin_name); }
static void native_call_plugin_dispatch(void* fn, uint16_t version, uint16_t event_id, uint32_t flags, uint64_t player_id, uint64_t request_key, const uint8_t* payload, uint32_t payload_len) {
	((plugin_dispatch_fn) fn)(version, event_id, flags, player_id, request_key, payload, payload_len);
}
*/
import "C"

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"

	"github.com/bedrock-gophers/plugin/internal/generator/output/ctxkey"
	"github.com/bedrock-gophers/plugin/plugin/abi"
)

type nativeRuntime struct {
	mu sync.Mutex

	handle     unsafe.Pointer
	dispatchFn unsafe.Pointer
	hostAPI    *C.native_host_api
	ctxID      uintptr
	closed     bool
}

type nativeHostContext struct {
	manager *Manager
	plugin  *pluginRuntime

	mutableMu   sync.RWMutex
	mutableNext uint64
	mutableByID map[uint64]*mutableState
}

var (
	nativeHostCtxMu   sync.RWMutex
	nativeHostCtxNext uintptr = 1
	nativeHostCtx             = map[uintptr]*nativeHostContext{}
)

func registerNativeHostContext(m *Manager, plug *pluginRuntime) uintptr {
	nativeHostCtxMu.Lock()
	id := nativeHostCtxNext
	nativeHostCtxNext++
	nativeHostCtx[id] = &nativeHostContext{manager: m, plugin: plug, mutableNext: 1, mutableByID: map[uint64]*mutableState{}}
	nativeHostCtxMu.Unlock()
	return id
}

func nativeHostContextByID(id uintptr) (*nativeHostContext, bool) {
	nativeHostCtxMu.RLock()
	ctx, ok := nativeHostCtx[id]
	nativeHostCtxMu.RUnlock()
	return ctx, ok
}

func unregisterNativeHostContext(id uintptr) {
	nativeHostCtxMu.Lock()
	delete(nativeHostCtx, id)
	nativeHostCtxMu.Unlock()
}

func registerNativeMutableState(ctxID uintptr, mutable *mutableState) uint64 {
	if mutable == nil {
		return 0
	}
	hostCtx, ok := nativeHostContextByID(ctxID)
	if !ok || hostCtx == nil {
		return 0
	}
	hostCtx.mutableMu.Lock()
	id := hostCtx.mutableNext
	hostCtx.mutableNext++
	hostCtx.mutableByID[id] = mutable
	hostCtx.mutableMu.Unlock()
	return id
}

func unregisterNativeMutableState(ctxID uintptr, requestKey uint64) {
	if requestKey == 0 {
		return
	}
	hostCtx, ok := nativeHostContextByID(ctxID)
	if !ok || hostCtx == nil {
		return
	}
	hostCtx.mutableMu.Lock()
	delete(hostCtx.mutableByID, requestKey)
	hostCtx.mutableMu.Unlock()
}

func nativeMutableState(ctxID uintptr, requestKey uint64) (*mutableState, bool) {
	if requestKey == 0 {
		return nil, false
	}
	hostCtx, ok := nativeHostContextByID(ctxID)
	if !ok || hostCtx == nil {
		return nil, false
	}
	hostCtx.mutableMu.RLock()
	mutable, ok := hostCtx.mutableByID[requestKey]
	hostCtx.mutableMu.RUnlock()
	return mutable, ok
}

func (m *Manager) startNativePlugin(plug *pluginRuntime, kind string) error {
	kind = strings.TrimSpace(strings.ToLower(kind))
	if kind == "" {
		kind = "native"
	}
	if plug == nil {
		return fmt.Errorf("%s plugin runtime is nil", kind)
	}
	if !strings.EqualFold(filepath.Ext(plug.path), ".so") {
		return fmt.Errorf("%s plugin must be a native shared library (.so), got %q", kind, filepath.Ext(plug.path))
	}

	rt, err := loadNativeRuntime(m, plug, kind)
	if err != nil {
		return err
	}
	plug.runtime = rt
	plug.onUnload = rt.close
	return nil
}

func loadNativeRuntime(m *Manager, plug *pluginRuntime, kind string) (*nativeRuntime, error) {
	cPath := C.CString(plug.path)
	defer C.free(unsafe.Pointer(cPath))
	handle := C.native_open_library(cPath)
	if handle == nil {
		return nil, fmt.Errorf("open %s plugin %q: %s", kind, plug.path, nativeLastError())
	}

	loadFn, err := lookupNativeSymbol(handle, "PluginLoad")
	if err != nil {
		return nil, err
	}
	dispatchFn, err := lookupNativeSymbol(handle, "PluginDispatchEvent")
	if err != nil {
		return nil, err
	}

	ctxID := registerNativeHostContext(m, plug)
	hostAPI := (*C.native_host_api)(C.malloc(C.size_t(C.sizeof_native_host_api)))
	*hostAPI = C.make_host_api(C.uintptr_t(ctxID))

	cName := C.CString(plug.name)
	defer C.free(unsafe.Pointer(cName))
	if C.native_call_plugin_load(loadFn, hostAPI, cName) == 0 {
		unregisterNativeHostContext(ctxID)
		C.free(unsafe.Pointer(hostAPI))
		return nil, fmt.Errorf("PluginLoad returned false for %s plugin %q", kind, plug.path)
	}

	return &nativeRuntime{handle: handle, dispatchFn: dispatchFn, hostAPI: hostAPI, ctxID: ctxID}, nil
}

func lookupNativeSymbol(handle unsafe.Pointer, name string) (unsafe.Pointer, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	ptr := C.native_lookup_symbol(handle, cName)
	if ptr == nil {
		return nil, fmt.Errorf("lookup native plugin symbol %q failed: %s", name, nativeLastError())
	}
	return ptr, nil
}

func nativeLastError() string {
	err := C.native_last_error()
	if err == nil {
		return "unknown error"
	}
	return C.GoString(err)
}

func (rt *nativeRuntime) dispatch(_ *Manager, _ *pluginRuntime, desc abi.EventDescriptor, payload []byte) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.closed {
		return fmt.Errorf("native runtime is closed")
	}
	var payloadPtr *C.uint8_t
	if len(payload) > 0 {
		payloadPtr = (*C.uint8_t)(unsafe.Pointer(&payload[0]))
	}
	C.native_call_plugin_dispatch(rt.dispatchFn, C.uint16_t(desc.Version), C.uint16_t(desc.EventID), C.uint32_t(desc.Flags), C.uint64_t(desc.PlayerID), C.uint64_t(desc.RequestKey), payloadPtr, C.uint32_t(len(payload)))
	return nil
}

func (rt *nativeRuntime) close() {
	if rt == nil {
		return
	}
	rt.mu.Lock()
	if rt.closed {
		rt.mu.Unlock()
		return
	}
	rt.closed = true
	hostAPI := rt.hostAPI
	ctxID := rt.ctxID
	rt.mu.Unlock()
	if hostAPI != nil {
		C.free(unsafe.Pointer(hostAPI))
	}
	if ctxID != 0 {
		unregisterNativeHostContext(ctxID)
	}
}

func nativeCString(v string) *C.char { return C.CString(v) }

func nativeGoString(v *C.char) string {
	if v == nil {
		return ""
	}
	return C.GoString(v)
}

func nativeBoolInt(v bool) C.int {
	if v {
		return 1
	}
	return 0
}

func nativeCBytes(v []byte) (*C.uint8_t, C.uint32_t) {
	if len(v) == 0 {
		return nil, 0
	}
	ptr := C.CBytes(v)
	if ptr == nil {
		return nil, 0
	}
	return (*C.uint8_t)(ptr), C.uint32_t(len(v))
}

func csvAliases(v string) []string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		alias := strings.TrimSpace(part)
		if alias != "" {
			out = append(out, alias)
		}
	}
	return out
}

func decodeCommandOverloads(data []byte) ([]CommandOverloadSpec, bool) {
	d := abi.NewDecoder(data)
	count := int(d.U32())
	if !d.Ok() || count < 0 || count > 1024 {
		return nil, false
	}
	out := make([]CommandOverloadSpec, 0, count)
	for i := 0; i < count; i++ {
		paramCount := int(d.U32())
		if !d.Ok() || paramCount < 0 || paramCount > 1024 {
			return nil, false
		}
		params := make([]CommandParameterSpec, 0, paramCount)
		for j := 0; j < paramCount; j++ {
			name := d.String()
			kind := CommandParameterKind(d.U8())
			optional := d.Bool()
			enumCount := int(d.U32())
			if !d.Ok() || enumCount < 0 || enumCount > 4096 {
				return nil, false
			}
			enumOptions := make([]string, 0, enumCount)
			for k := 0; k < enumCount; k++ {
				enumOptions = append(enumOptions, d.String())
			}
			params = append(params, CommandParameterSpec{Name: name, Kind: kind, Optional: optional, EnumOptions: enumOptions})
		}
		out = append(out, CommandOverloadSpec{Parameters: params})
	}
	return out, d.Ok()
}

//export host_register_command
func host_register_command(ctx C.uintptr_t, pluginName *C.char, name *C.char, description *C.char, aliasesCSV *C.char, handlerID C.uint32_t, overloads *C.uint8_t, overloadsLen C.uint32_t) C.int {
	hostCtx, ok := nativeHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil || hostCtx.plugin == nil {
		return 0
	}
	var overloadSpecs []CommandOverloadSpec
	if overloadsLen > 0 {
		if overloads == nil {
			return 0
		}
		decoded, ok := decodeCommandOverloads(C.GoBytes(unsafe.Pointer(overloads), C.int(overloadsLen)))
		if !ok {
			return 0
		}
		overloadSpecs = decoded
	}
	if hostCtx.manager.RegisterCommand(nativeGoString(pluginName), nativeGoString(name), nativeGoString(description), csvAliases(nativeGoString(aliasesCSV)), uint32(handlerID), overloadSpecs) {
		return 1
	}
	return 0
}

//export host_manage_plugins
func host_manage_plugins(ctx C.uintptr_t, action C.uint32_t, target *C.char, outLen *C.uint32_t) *C.uint8_t {
	if outLen != nil {
		*outLen = 0
	}
	hostCtx, ok := nativeHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return nil
	}
	names, err := hostCtx.manager.ManagePlugins(uint32(action), nativeGoString(target))
	if err != nil {
		return nil
	}
	ptr, size := nativeCBytes(encodeStringListPayload(names))
	if outLen != nil {
		*outLen = size
	}
	return ptr
}

//export host_resolve_player_by_name
func host_resolve_player_by_name(ctx C.uintptr_t, name *C.char) C.uint64_t {
	hostCtx, ok := nativeHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return C.uint64_t(hostCtx.manager.ResolvePlayerByName(nativeGoString(name)))
}

//export host_player_handle
func host_player_handle(ctx C.uintptr_t, playerID C.uint64_t) C.uint64_t {
	hostCtx, ok := nativeHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return C.uint64_t(hostCtx.manager.PlayerHandle(uint64(playerID)))
}

//export host_online_player_names
func host_online_player_names(ctx C.uintptr_t, outLen *C.uint32_t) *C.uint8_t {
	if outLen != nil {
		*outLen = 0
	}
	hostCtx, ok := nativeHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return nil
	}
	ptr, size := nativeCBytes(encodeStringListPayload(hostCtx.manager.OnlinePlayerNames()))
	if outLen != nil {
		*outLen = size
	}
	return ptr
}

//export host_console_message
func host_console_message(ctx C.uintptr_t, pluginName *C.char, message *C.char) {
	hostCtx, ok := nativeHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return
	}
	name := nativeGoString(pluginName)
	if name == "" && hostCtx.plugin != nil {
		name = hostCtx.plugin.name
	}
	hostCtx.manager.ConsoleMessage(name, nativeGoString(message))
}

//export host_call
func host_call(ctx C.uintptr_t, op C.uint32_t, payload *C.uint8_t, payloadLen C.uint32_t, outLen *C.uint32_t) *C.uint8_t {
	if outLen != nil {
		*outLen = 0
	}
	hostCtx, ok := nativeHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return nil
	}
	var in []byte
	if payloadLen > 0 {
		if payload == nil {
			return nil
		}
		in = C.GoBytes(unsafe.Pointer(payload), C.int(payloadLen))
	}
	ptr, size := nativeCBytes(hostCtx.manager.HostCall(uint32(op), in))
	if outLen != nil {
		*outLen = size
	}
	return ptr
}

//export host_event_cancel
func host_event_cancel(ctx C.uintptr_t, requestKey C.uint64_t) C.int {
	mutable, ok := nativeMutableState(uintptr(ctx), uint64(requestKey))
	if !ok || mutable == nil {
		return 0
	}
	mutable.Cancel()
	return 1
}

//export host_event_i64_set
func host_event_i64_set(ctx C.uintptr_t, requestKey C.uint64_t, slot C.uint32_t, value C.int64_t) C.int {
	mutable, ok := nativeMutableState(uintptr(ctx), uint64(requestKey))
	if !ok || mutable == nil {
		return 0
	}
	mutable.SetI64(uint32(slot), int64(value))
	return 1
}

//export host_event_f64_set
func host_event_f64_set(ctx C.uintptr_t, requestKey C.uint64_t, slot C.uint32_t, value C.double) C.int {
	mutable, ok := nativeMutableState(uintptr(ctx), uint64(requestKey))
	if !ok || mutable == nil {
		return 0
	}
	mutable.SetF64(uint32(slot), float64(value))
	return 1
}

//export host_event_bool_set
func host_event_bool_set(ctx C.uintptr_t, requestKey C.uint64_t, slot C.uint32_t, value C.int) C.int {
	mutable, ok := nativeMutableState(uintptr(ctx), uint64(requestKey))
	if !ok || mutable == nil {
		return 0
	}
	mutable.SetBool(uint32(slot), value != 0)
	return 1
}

//export host_event_string_set
func host_event_string_set(ctx C.uintptr_t, requestKey C.uint64_t, slot C.uint32_t, value *C.char) C.int {
	mutable, ok := nativeMutableState(uintptr(ctx), uint64(requestKey))
	if !ok || mutable == nil {
		return 0
	}
	mutable.SetString(uint32(slot), nativeGoString(value))
	return 1
}

func setMutableItemDrop(mutable *mutableState, value ItemStackData) {
	if mutable == nil {
		return
	}
	mutable.SetI64(ctxkey.ItemDropCount, int64(value.Count))
	mutable.SetBool(ctxkey.ItemDropHasItem, value.HasItem)
	mutable.SetString(ctxkey.ItemDropName, value.ItemName)
	mutable.SetI64(ctxkey.ItemDropMeta, int64(value.ItemMeta))
	mutable.SetString(ctxkey.ItemDropCustomName, value.CustomName)
}

//export host_event_item_drop_set
func host_event_item_drop_set(ctx C.uintptr_t, requestKey C.uint64_t, payload *C.uint8_t, payloadLen C.uint32_t) C.int {
	mutable, ok := nativeMutableState(uintptr(ctx), uint64(requestKey))
	if !ok || mutable == nil || payload == nil || payloadLen == 0 {
		return 0
	}
	d := abi.NewDecoder(C.GoBytes(unsafe.Pointer(payload), C.int(payloadLen)))
	value, ok := decodeItemStackPayload(d)
	if !ok {
		return 0
	}
	setMutableItemDrop(mutable, value)
	return 1
}

//export host_player_name
func host_player_name(ctx C.uintptr_t, playerID C.uint64_t) *C.char {
	hostCtx, ok := nativeHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return nativeCString("")
	}
	return nativeCString(hostCtx.manager.PlayerName(uint64(playerID)))
}

//export host_player_latency
func host_player_latency(ctx C.uintptr_t, playerID C.uint64_t) C.int64_t {
	hostCtx, ok := nativeHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return C.int64_t(hostCtx.manager.PlayerLatency(uint64(playerID)).Milliseconds())
}

//export host_player_message
func host_player_message(ctx C.uintptr_t, playerID C.uint64_t, message *C.char) {
	hostCtx, ok := nativeHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return
	}
	hostCtx.manager.PlayerMessage(uint64(playerID), nativeGoString(message))
}

//export host_free_string
func host_free_string(value *C.char) {
	if value != nil {
		C.free(unsafe.Pointer(value))
	}
}

//export host_free_bytes
func host_free_bytes(value *C.uint8_t) {
	if value != nil {
		C.free(unsafe.Pointer(value))
	}
}
