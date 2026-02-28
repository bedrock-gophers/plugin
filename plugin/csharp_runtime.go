//go:build cgo

package plugin

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdint.h>
#include <stdlib.h>

extern int csharp_register_command(uintptr_t ctx, char* plugin_name, char* name, char* description, char* aliases_csv, uint32_t handler_id, uint8_t* overloads, uint32_t overloads_len);
extern uint8_t* csharp_manage_plugins(uintptr_t ctx, uint32_t action, char* target, uint32_t* out_len);
extern uint64_t csharp_resolve_player_by_name(uintptr_t ctx, char* name);
extern uint8_t* csharp_online_player_names(uintptr_t ctx, uint32_t* out_len);
extern void csharp_console_message(uintptr_t ctx, char* plugin_name, char* message);

extern double csharp_player_health(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_health(uintptr_t ctx, uint64_t player_id, double health);
extern int32_t csharp_player_food(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_food(uintptr_t ctx, uint64_t player_id, int32_t food);
extern char* csharp_player_name(uintptr_t ctx, uint64_t player_id);
extern int32_t csharp_player_game_mode(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_game_mode(uintptr_t ctx, uint64_t player_id, int32_t mode);
extern char* csharp_player_xuid(uintptr_t ctx, uint64_t player_id);
extern char* csharp_player_device_id(uintptr_t ctx, uint64_t player_id);
extern char* csharp_player_device_model(uintptr_t ctx, uint64_t player_id);
extern char* csharp_player_self_signed_id(uintptr_t ctx, uint64_t player_id);
extern char* csharp_player_name_tag(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_name_tag(uintptr_t ctx, uint64_t player_id, char* value);
extern char* csharp_player_score_tag(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_score_tag(uintptr_t ctx, uint64_t player_id, char* value);
extern double csharp_player_absorption(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_absorption(uintptr_t ctx, uint64_t player_id, double value);
extern double csharp_player_max_health(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_max_health(uintptr_t ctx, uint64_t player_id, double value);
extern double csharp_player_speed(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_speed(uintptr_t ctx, uint64_t player_id, double value);
extern double csharp_player_flight_speed(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_flight_speed(uintptr_t ctx, uint64_t player_id, double value);
extern double csharp_player_vertical_flight_speed(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_vertical_flight_speed(uintptr_t ctx, uint64_t player_id, double value);
extern int32_t csharp_player_experience(uintptr_t ctx, uint64_t player_id);
extern int32_t csharp_player_experience_level(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_experience_level(uintptr_t ctx, uint64_t player_id, int32_t value);
extern double csharp_player_experience_progress(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_experience_progress(uintptr_t ctx, uint64_t player_id, double value);
extern int csharp_player_on_ground(uintptr_t ctx, uint64_t player_id);
extern int csharp_player_sneaking(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_sneaking(uintptr_t ctx, uint64_t player_id, int value);
extern int csharp_player_sprinting(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_sprinting(uintptr_t ctx, uint64_t player_id, int value);
extern int csharp_player_swimming(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_swimming(uintptr_t ctx, uint64_t player_id, int value);
extern int csharp_player_flying(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_flying(uintptr_t ctx, uint64_t player_id, int value);
extern int csharp_player_gliding(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_gliding(uintptr_t ctx, uint64_t player_id, int value);
extern int csharp_player_crawling(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_crawling(uintptr_t ctx, uint64_t player_id, int value);
extern int csharp_player_using_item(uintptr_t ctx, uint64_t player_id);
extern int csharp_player_invisible(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_invisible(uintptr_t ctx, uint64_t player_id, int value);
extern int csharp_player_immobile(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_immobile(uintptr_t ctx, uint64_t player_id, int value);
extern int csharp_player_dead(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_on_fire_millis(uintptr_t ctx, uint64_t player_id, int64_t millis);
extern int csharp_add_player_food(uintptr_t ctx, uint64_t player_id, int32_t points);
extern int csharp_player_use_item(uintptr_t ctx, uint64_t player_id);
extern int csharp_player_jump(uintptr_t ctx, uint64_t player_id);
extern int csharp_player_swing_arm(uintptr_t ctx, uint64_t player_id);
extern int csharp_player_wake(uintptr_t ctx, uint64_t player_id);
extern int csharp_player_extinguish(uintptr_t ctx, uint64_t player_id);
extern int csharp_set_player_show_coordinates(uintptr_t ctx, uint64_t player_id, int value);
extern void csharp_player_message(uintptr_t ctx, uint64_t player_id, char* message);

extern void csharp_free_string(char* value);
extern void csharp_free_bytes(uint8_t* value);

typedef struct {
	uintptr_t ctx;

	int (*register_command)(uintptr_t ctx, char* plugin_name, char* name, char* description, char* aliases_csv, uint32_t handler_id, uint8_t* overloads, uint32_t overloads_len);
	uint8_t* (*manage_plugins)(uintptr_t ctx, uint32_t action, char* target, uint32_t* out_len);
	uint64_t (*resolve_player_by_name)(uintptr_t ctx, char* name);
	uint8_t* (*online_player_names)(uintptr_t ctx, uint32_t* out_len);
	void (*console_message)(uintptr_t ctx, char* plugin_name, char* message);

	double (*player_health)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_health)(uintptr_t ctx, uint64_t player_id, double health);
	int32_t (*player_food)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_food)(uintptr_t ctx, uint64_t player_id, int32_t food);
	char* (*player_name)(uintptr_t ctx, uint64_t player_id);
	int32_t (*player_game_mode)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_game_mode)(uintptr_t ctx, uint64_t player_id, int32_t mode);
	char* (*player_xuid)(uintptr_t ctx, uint64_t player_id);
	char* (*player_device_id)(uintptr_t ctx, uint64_t player_id);
	char* (*player_device_model)(uintptr_t ctx, uint64_t player_id);
	char* (*player_self_signed_id)(uintptr_t ctx, uint64_t player_id);
	char* (*player_name_tag)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_name_tag)(uintptr_t ctx, uint64_t player_id, char* value);
	char* (*player_score_tag)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_score_tag)(uintptr_t ctx, uint64_t player_id, char* value);
	double (*player_absorption)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_absorption)(uintptr_t ctx, uint64_t player_id, double value);
	double (*player_max_health)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_max_health)(uintptr_t ctx, uint64_t player_id, double value);
	double (*player_speed)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_speed)(uintptr_t ctx, uint64_t player_id, double value);
	double (*player_flight_speed)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_flight_speed)(uintptr_t ctx, uint64_t player_id, double value);
	double (*player_vertical_flight_speed)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_vertical_flight_speed)(uintptr_t ctx, uint64_t player_id, double value);
	int32_t (*player_experience)(uintptr_t ctx, uint64_t player_id);
	int32_t (*player_experience_level)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_experience_level)(uintptr_t ctx, uint64_t player_id, int32_t value);
	double (*player_experience_progress)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_experience_progress)(uintptr_t ctx, uint64_t player_id, double value);
	int (*player_on_ground)(uintptr_t ctx, uint64_t player_id);
	int (*player_sneaking)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_sneaking)(uintptr_t ctx, uint64_t player_id, int value);
	int (*player_sprinting)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_sprinting)(uintptr_t ctx, uint64_t player_id, int value);
	int (*player_swimming)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_swimming)(uintptr_t ctx, uint64_t player_id, int value);
	int (*player_flying)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_flying)(uintptr_t ctx, uint64_t player_id, int value);
	int (*player_gliding)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_gliding)(uintptr_t ctx, uint64_t player_id, int value);
	int (*player_crawling)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_crawling)(uintptr_t ctx, uint64_t player_id, int value);
	int (*player_using_item)(uintptr_t ctx, uint64_t player_id);
	int (*player_invisible)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_invisible)(uintptr_t ctx, uint64_t player_id, int value);
	int (*player_immobile)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_immobile)(uintptr_t ctx, uint64_t player_id, int value);
	int (*player_dead)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_on_fire_millis)(uintptr_t ctx, uint64_t player_id, int64_t millis);
	int (*add_player_food)(uintptr_t ctx, uint64_t player_id, int32_t points);
	int (*player_use_item)(uintptr_t ctx, uint64_t player_id);
	int (*player_jump)(uintptr_t ctx, uint64_t player_id);
	int (*player_swing_arm)(uintptr_t ctx, uint64_t player_id);
	int (*player_wake)(uintptr_t ctx, uint64_t player_id);
	int (*player_extinguish)(uintptr_t ctx, uint64_t player_id);
	int (*set_player_show_coordinates)(uintptr_t ctx, uint64_t player_id, int value);
	void (*player_message)(uintptr_t ctx, uint64_t player_id, char* message);

	void (*free_string)(char* value);
	void (*free_bytes)(uint8_t* value);
} csharp_host_api;

typedef int (*plugin_load_fn)(csharp_host_api* host, char* plugin_name);
typedef void (*plugin_unload_fn)(void);
typedef void (*plugin_dispatch_fn)(uint16_t version, uint16_t event_id, uint32_t flags, uint64_t player_id, uint64_t request_key, const uint8_t* payload, uint32_t payload_len);

static csharp_host_api make_host_api(uintptr_t ctx) {
	csharp_host_api api;
	api.ctx = ctx;

	api.register_command = csharp_register_command;
	api.manage_plugins = csharp_manage_plugins;
	api.resolve_player_by_name = csharp_resolve_player_by_name;
	api.online_player_names = csharp_online_player_names;
	api.console_message = csharp_console_message;

	api.player_health = csharp_player_health;
	api.set_player_health = csharp_set_player_health;
	api.player_food = csharp_player_food;
	api.set_player_food = csharp_set_player_food;
	api.player_name = csharp_player_name;
	api.player_game_mode = csharp_player_game_mode;
	api.set_player_game_mode = csharp_set_player_game_mode;
	api.player_xuid = csharp_player_xuid;
	api.player_device_id = csharp_player_device_id;
	api.player_device_model = csharp_player_device_model;
	api.player_self_signed_id = csharp_player_self_signed_id;
	api.player_name_tag = csharp_player_name_tag;
	api.set_player_name_tag = csharp_set_player_name_tag;
	api.player_score_tag = csharp_player_score_tag;
	api.set_player_score_tag = csharp_set_player_score_tag;
	api.player_absorption = csharp_player_absorption;
	api.set_player_absorption = csharp_set_player_absorption;
	api.player_max_health = csharp_player_max_health;
	api.set_player_max_health = csharp_set_player_max_health;
	api.player_speed = csharp_player_speed;
	api.set_player_speed = csharp_set_player_speed;
	api.player_flight_speed = csharp_player_flight_speed;
	api.set_player_flight_speed = csharp_set_player_flight_speed;
	api.player_vertical_flight_speed = csharp_player_vertical_flight_speed;
	api.set_player_vertical_flight_speed = csharp_set_player_vertical_flight_speed;
	api.player_experience = csharp_player_experience;
	api.player_experience_level = csharp_player_experience_level;
	api.set_player_experience_level = csharp_set_player_experience_level;
	api.player_experience_progress = csharp_player_experience_progress;
	api.set_player_experience_progress = csharp_set_player_experience_progress;
	api.player_on_ground = csharp_player_on_ground;
	api.player_sneaking = csharp_player_sneaking;
	api.set_player_sneaking = csharp_set_player_sneaking;
	api.player_sprinting = csharp_player_sprinting;
	api.set_player_sprinting = csharp_set_player_sprinting;
	api.player_swimming = csharp_player_swimming;
	api.set_player_swimming = csharp_set_player_swimming;
	api.player_flying = csharp_player_flying;
	api.set_player_flying = csharp_set_player_flying;
	api.player_gliding = csharp_player_gliding;
	api.set_player_gliding = csharp_set_player_gliding;
	api.player_crawling = csharp_player_crawling;
	api.set_player_crawling = csharp_set_player_crawling;
	api.player_using_item = csharp_player_using_item;
	api.player_invisible = csharp_player_invisible;
	api.set_player_invisible = csharp_set_player_invisible;
	api.player_immobile = csharp_player_immobile;
	api.set_player_immobile = csharp_set_player_immobile;
	api.player_dead = csharp_player_dead;
	api.set_player_on_fire_millis = csharp_set_player_on_fire_millis;
	api.add_player_food = csharp_add_player_food;
	api.player_use_item = csharp_player_use_item;
	api.player_jump = csharp_player_jump;
	api.player_swing_arm = csharp_player_swing_arm;
	api.player_wake = csharp_player_wake;
	api.player_extinguish = csharp_player_extinguish;
	api.set_player_show_coordinates = csharp_set_player_show_coordinates;
	api.player_message = csharp_player_message;

	api.free_string = csharp_free_string;
	api.free_bytes = csharp_free_bytes;
	return api;
}

static void* csharp_open_library(char* path) {
	return dlopen(path, RTLD_NOW | RTLD_LOCAL);
}

static char* csharp_last_error() {
	return dlerror();
}

static void* csharp_lookup_symbol(void* handle, char* symbol) {
	return dlsym(handle, symbol);
}

static int csharp_close_library(void* handle) {
	return dlclose(handle);
}

static int csharp_call_plugin_load(void* fn, csharp_host_api* host, char* plugin_name) {
	return ((plugin_load_fn) fn)(host, plugin_name);
}

static void csharp_call_plugin_unload(void* fn) {
	((plugin_unload_fn) fn)();
}

static void csharp_call_plugin_dispatch(void* fn, uint16_t version, uint16_t event_id, uint32_t flags, uint64_t player_id, uint64_t request_key, const uint8_t* payload, uint32_t payload_len) {
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

	"github.com/bedrock-gophers/plugin/plugin/abi"
	guest "github.com/bedrock-gophers/plugin/plugin/sdk/go"
)

type csharpRuntime struct {
	mu sync.Mutex

	handle     unsafe.Pointer
	loadFn     unsafe.Pointer
	unloadFn   unsafe.Pointer
	dispatchFn unsafe.Pointer
	hostAPI    *C.csharp_host_api
	ctxID      uintptr
	closed     bool
}

type csharpHostContext struct {
	manager *Manager
	plugin  *pluginRuntime
}

var (
	csharpHostCtxMu   sync.RWMutex
	csharpHostCtxNext uintptr = 1
	csharpHostCtx             = map[uintptr]csharpHostContext{}
)

func registerCSharpHostContext(m *Manager, plug *pluginRuntime) uintptr {
	csharpHostCtxMu.Lock()
	id := csharpHostCtxNext
	csharpHostCtxNext++
	csharpHostCtx[id] = csharpHostContext{manager: m, plugin: plug}
	csharpHostCtxMu.Unlock()
	return id
}

func csharpHostContextByID(id uintptr) (csharpHostContext, bool) {
	csharpHostCtxMu.RLock()
	v, ok := csharpHostCtx[id]
	csharpHostCtxMu.RUnlock()
	return v, ok
}

func unregisterCSharpHostContext(id uintptr) {
	csharpHostCtxMu.Lock()
	delete(csharpHostCtx, id)
	csharpHostCtxMu.Unlock()
}

func (m *Manager) startCSharpPlugin(plug *pluginRuntime) error {
	if plug == nil {
		return fmt.Errorf("csharp plugin runtime is nil")
	}
	if !strings.EqualFold(filepath.Ext(plug.path), ".so") {
		return fmt.Errorf("csharp plugin must be a native shared library (.so), got %q", filepath.Ext(plug.path))
	}

	rt, err := loadCSharpRuntime(m, plug)
	if err != nil {
		return err
	}
	plug.csharp = rt
	plug.onUnload = func() {
		rt.close()
	}
	return nil
}

func loadCSharpRuntime(m *Manager, plug *pluginRuntime) (*csharpRuntime, error) {
	cPath := C.CString(plug.path)
	defer C.free(unsafe.Pointer(cPath))

	handle := C.csharp_open_library(cPath)
	if handle == nil {
		return nil, fmt.Errorf("open csharp plugin %q: %s", plug.path, csharpLastError())
	}

	loadFn, err := lookupCSharpSymbol(handle, "PluginLoad")
	if err != nil {
		C.csharp_close_library(handle)
		return nil, err
	}
	unloadFn, err := lookupCSharpSymbol(handle, "PluginUnload")
	if err != nil {
		C.csharp_close_library(handle)
		return nil, err
	}
	dispatchFn, err := lookupCSharpSymbol(handle, "PluginDispatchEvent")
	if err != nil {
		C.csharp_close_library(handle)
		return nil, err
	}

	ctxID := registerCSharpHostContext(m, plug)
	hostAPI := (*C.csharp_host_api)(C.malloc(C.size_t(C.sizeof_csharp_host_api)))
	*hostAPI = C.make_host_api(C.uintptr_t(ctxID))

	cName := C.CString(plug.name)
	defer C.free(unsafe.Pointer(cName))

	ok := C.csharp_call_plugin_load(loadFn, hostAPI, cName)
	if ok == 0 {
		unregisterCSharpHostContext(ctxID)
		C.free(unsafe.Pointer(hostAPI))
		C.csharp_close_library(handle)
		return nil, fmt.Errorf("PluginLoad returned false for csharp plugin %q", plug.path)
	}

	return &csharpRuntime{
		handle:     handle,
		loadFn:     loadFn,
		unloadFn:   unloadFn,
		dispatchFn: dispatchFn,
		hostAPI:    hostAPI,
		ctxID:      ctxID,
	}, nil
}

func lookupCSharpSymbol(handle unsafe.Pointer, symbol string) (unsafe.Pointer, error) {
	cSymbol := C.CString(symbol)
	defer C.free(unsafe.Pointer(cSymbol))
	ptr := C.csharp_lookup_symbol(handle, cSymbol)
	if ptr == nil {
		return nil, fmt.Errorf("lookup csharp plugin symbol %q failed: %s", symbol, csharpLastError())
	}
	return ptr, nil
}

func csharpLastError() string {
	err := C.csharp_last_error()
	if err == nil {
		return "unknown error"
	}
	return C.GoString(err)
}

func (rt *csharpRuntime) dispatch(_ *Manager, _ *pluginRuntime, desc abi.EventDescriptor, payload []byte) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.closed {
		return fmt.Errorf("csharp runtime is closed")
	}
	var payloadPtr *C.uint8_t
	if len(payload) > 0 {
		payloadPtr = (*C.uint8_t)(unsafe.Pointer(&payload[0]))
	}
	C.csharp_call_plugin_dispatch(
		rt.dispatchFn,
		C.uint16_t(desc.Version),
		C.uint16_t(desc.EventID),
		C.uint32_t(desc.Flags),
		C.uint64_t(desc.PlayerID),
		C.uint64_t(desc.RequestKey),
		payloadPtr,
		C.uint32_t(len(payload)),
	)
	return nil
}

func (rt *csharpRuntime) close() {
	if rt == nil {
		return
	}

	rt.mu.Lock()
	if rt.closed {
		rt.mu.Unlock()
		return
	}
	rt.closed = true
	unloadFn := rt.unloadFn
	handle := rt.handle
	hostAPI := rt.hostAPI
	ctxID := rt.ctxID
	rt.mu.Unlock()

	if unloadFn != nil {
		C.csharp_call_plugin_unload(unloadFn)
	}
	if handle != nil {
		C.csharp_close_library(handle)
	}
	if hostAPI != nil {
		C.free(unsafe.Pointer(hostAPI))
	}
	if ctxID != 0 {
		unregisterCSharpHostContext(ctxID)
	}
}

func csvAliases(v string) []string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		out = append(out, token)
	}
	return out
}

func goCString(v *C.char) string {
	if v == nil {
		return ""
	}
	return C.GoString(v)
}

func boolInt(v bool) C.int {
	if v {
		return 1
	}
	return 0
}

func cString(v string) *C.char {
	return C.CString(v)
}

func cBytes(v []byte) (*C.uint8_t, C.uint32_t) {
	if len(v) == 0 {
		return nil, 0
	}
	p := C.CBytes(v)
	if p == nil {
		return nil, 0
	}
	return (*C.uint8_t)(p), C.uint32_t(len(v))
}

func encodeStringList(values []string) []byte {
	enc := abi.NewEncoder(max(32, 8+len(values)*8))
	enc.U32(uint32(len(values)))
	for _, v := range values {
		enc.String(v)
	}
	return enc.Data()
}

func decodeCommandOverloads(data []byte) ([]guest.CommandOverloadSpec, bool) {
	if len(data) == 0 {
		return nil, true
	}
	d := abi.NewDecoder(data)
	overloadCount := int(d.U32())
	if overloadCount < 0 || overloadCount > 1024 {
		return nil, false
	}
	out := make([]guest.CommandOverloadSpec, 0, overloadCount)
	for i := 0; i < overloadCount; i++ {
		paramCount := int(d.U32())
		if paramCount < 0 || paramCount > 1024 {
			return nil, false
		}
		params := make([]guest.CommandParameterSpec, 0, paramCount)
		for j := 0; j < paramCount; j++ {
			name := d.String()
			kind := guest.CommandParameterKind(d.U8())
			optional := d.Bool()
			enumCount := int(d.U32())
			if enumCount < 0 || enumCount > 4096 {
				return nil, false
			}
			enumOptions := make([]string, 0, enumCount)
			for k := 0; k < enumCount; k++ {
				enumOptions = append(enumOptions, d.String())
			}
			params = append(params, guest.CommandParameterSpec{
				Name:        name,
				Kind:        kind,
				Optional:    optional,
				EnumOptions: enumOptions,
			})
		}
		out = append(out, guest.CommandOverloadSpec{Parameters: params})
	}
	if !d.Ok() {
		return nil, false
	}
	return out, true
}

//export csharp_register_command
func csharp_register_command(ctx C.uintptr_t, pluginName *C.char, name *C.char, description *C.char, aliasesCSV *C.char, handlerID C.uint32_t, overloads *C.uint8_t, overloadsLen C.uint32_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil || hostCtx.plugin == nil {
		return 0
	}

	var overloadSpecs []guest.CommandOverloadSpec
	if overloadsLen > 0 {
		if overloads == nil {
			return 0
		}
		overloadBytes := C.GoBytes(unsafe.Pointer(overloads), C.int(overloadsLen))
		decoded, okDecode := decodeCommandOverloads(overloadBytes)
		if !okDecode {
			return 0
		}
		overloadSpecs = decoded
	}

	okRegister := hostCtx.manager.RegisterCommand(
		goCString(pluginName),
		goCString(name),
		goCString(description),
		csvAliases(goCString(aliasesCSV)),
		uint32(handlerID),
		overloadSpecs,
	)
	if !okRegister {
		return 0
	}
	return 1
}

//export csharp_manage_plugins
func csharp_manage_plugins(ctx C.uintptr_t, action C.uint32_t, target *C.char, outLen *C.uint32_t) *C.uint8_t {
	if outLen != nil {
		*outLen = 0
	}
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return nil
	}
	names, err := hostCtx.manager.ManagePlugins(uint32(action), goCString(target))
	if err != nil {
		pluginName := "csharp"
		if hostCtx.plugin != nil && hostCtx.plugin.name != "" {
			pluginName = hostCtx.plugin.name
		}
		hostCtx.manager.ConsoleMessage(pluginName, fmt.Sprintf("plugin manage action failed: %v", err))
		return nil
	}
	payload := encodeStringList(names)
	ptr, size := cBytes(payload)
	if outLen != nil {
		*outLen = size
	}
	return ptr
}

//export csharp_resolve_player_by_name
func csharp_resolve_player_by_name(ctx C.uintptr_t, name *C.char) C.uint64_t {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return C.uint64_t(hostCtx.manager.ResolvePlayerByName(goCString(name)))
}

//export csharp_online_player_names
func csharp_online_player_names(ctx C.uintptr_t, outLen *C.uint32_t) *C.uint8_t {
	if outLen != nil {
		*outLen = 0
	}
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return nil
	}
	payload := encodeStringList(hostCtx.manager.OnlinePlayerNames())
	ptr, size := cBytes(payload)
	if outLen != nil {
		*outLen = size
	}
	return ptr
}

//export csharp_console_message
func csharp_console_message(ctx C.uintptr_t, pluginName *C.char, message *C.char) {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return
	}
	name := goCString(pluginName)
	if name == "" && hostCtx.plugin != nil {
		name = hostCtx.plugin.name
	}
	hostCtx.manager.ConsoleMessage(name, goCString(message))
}

//export csharp_player_health
func csharp_player_health(ctx C.uintptr_t, playerID C.uint64_t) C.double {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return C.double(hostCtx.manager.PlayerHealth(uint64(playerID)))
}

//export csharp_set_player_health
func csharp_set_player_health(ctx C.uintptr_t, playerID C.uint64_t, health C.double) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerHealth(uint64(playerID), float64(health)))
}

//export csharp_player_food
func csharp_player_food(ctx C.uintptr_t, playerID C.uint64_t) C.int32_t {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return C.int32_t(hostCtx.manager.PlayerFood(uint64(playerID)))
}

//export csharp_set_player_food
func csharp_set_player_food(ctx C.uintptr_t, playerID C.uint64_t, food C.int32_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerFood(uint64(playerID), int32(food)))
}

//export csharp_player_name
func csharp_player_name(ctx C.uintptr_t, playerID C.uint64_t) *C.char {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return cString("")
	}
	return cString(hostCtx.manager.PlayerName(uint64(playerID)))
}

//export csharp_player_game_mode
func csharp_player_game_mode(ctx C.uintptr_t, playerID C.uint64_t) C.int32_t {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return C.int32_t(hostCtx.manager.PlayerGameMode(uint64(playerID)))
}

//export csharp_set_player_game_mode
func csharp_set_player_game_mode(ctx C.uintptr_t, playerID C.uint64_t, mode C.int32_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerGameMode(uint64(playerID), int32(mode)))
}

//export csharp_player_xuid
func csharp_player_xuid(ctx C.uintptr_t, playerID C.uint64_t) *C.char {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return cString("")
	}
	return cString(hostCtx.manager.PlayerXUID(uint64(playerID)))
}

//export csharp_player_device_id
func csharp_player_device_id(ctx C.uintptr_t, playerID C.uint64_t) *C.char {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return cString("")
	}
	return cString(hostCtx.manager.PlayerDeviceID(uint64(playerID)))
}

//export csharp_player_device_model
func csharp_player_device_model(ctx C.uintptr_t, playerID C.uint64_t) *C.char {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return cString("")
	}
	return cString(hostCtx.manager.PlayerDeviceModel(uint64(playerID)))
}

//export csharp_player_self_signed_id
func csharp_player_self_signed_id(ctx C.uintptr_t, playerID C.uint64_t) *C.char {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return cString("")
	}
	return cString(hostCtx.manager.PlayerSelfSignedID(uint64(playerID)))
}

//export csharp_player_name_tag
func csharp_player_name_tag(ctx C.uintptr_t, playerID C.uint64_t) *C.char {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return cString("")
	}
	return cString(hostCtx.manager.PlayerNameTag(uint64(playerID)))
}

//export csharp_set_player_name_tag
func csharp_set_player_name_tag(ctx C.uintptr_t, playerID C.uint64_t, value *C.char) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerNameTag(uint64(playerID), goCString(value)))
}

//export csharp_player_score_tag
func csharp_player_score_tag(ctx C.uintptr_t, playerID C.uint64_t) *C.char {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return cString("")
	}
	return cString(hostCtx.manager.PlayerScoreTag(uint64(playerID)))
}

//export csharp_set_player_score_tag
func csharp_set_player_score_tag(ctx C.uintptr_t, playerID C.uint64_t, value *C.char) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerScoreTag(uint64(playerID), goCString(value)))
}

//export csharp_player_absorption
func csharp_player_absorption(ctx C.uintptr_t, playerID C.uint64_t) C.double {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return C.double(hostCtx.manager.PlayerAbsorption(uint64(playerID)))
}

//export csharp_set_player_absorption
func csharp_set_player_absorption(ctx C.uintptr_t, playerID C.uint64_t, value C.double) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerAbsorption(uint64(playerID), float64(value)))
}

//export csharp_player_max_health
func csharp_player_max_health(ctx C.uintptr_t, playerID C.uint64_t) C.double {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return C.double(hostCtx.manager.PlayerMaxHealth(uint64(playerID)))
}

//export csharp_set_player_max_health
func csharp_set_player_max_health(ctx C.uintptr_t, playerID C.uint64_t, value C.double) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerMaxHealth(uint64(playerID), float64(value)))
}

//export csharp_player_speed
func csharp_player_speed(ctx C.uintptr_t, playerID C.uint64_t) C.double {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return C.double(hostCtx.manager.PlayerSpeed(uint64(playerID)))
}

//export csharp_set_player_speed
func csharp_set_player_speed(ctx C.uintptr_t, playerID C.uint64_t, value C.double) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerSpeed(uint64(playerID), float64(value)))
}

//export csharp_player_flight_speed
func csharp_player_flight_speed(ctx C.uintptr_t, playerID C.uint64_t) C.double {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return C.double(hostCtx.manager.PlayerFlightSpeed(uint64(playerID)))
}

//export csharp_set_player_flight_speed
func csharp_set_player_flight_speed(ctx C.uintptr_t, playerID C.uint64_t, value C.double) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerFlightSpeed(uint64(playerID), float64(value)))
}

//export csharp_player_vertical_flight_speed
func csharp_player_vertical_flight_speed(ctx C.uintptr_t, playerID C.uint64_t) C.double {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return C.double(hostCtx.manager.PlayerVerticalFlightSpeed(uint64(playerID)))
}

//export csharp_set_player_vertical_flight_speed
func csharp_set_player_vertical_flight_speed(ctx C.uintptr_t, playerID C.uint64_t, value C.double) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerVerticalFlightSpeed(uint64(playerID), float64(value)))
}

//export csharp_player_experience
func csharp_player_experience(ctx C.uintptr_t, playerID C.uint64_t) C.int32_t {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return C.int32_t(hostCtx.manager.PlayerExperience(uint64(playerID)))
}

//export csharp_player_experience_level
func csharp_player_experience_level(ctx C.uintptr_t, playerID C.uint64_t) C.int32_t {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return C.int32_t(hostCtx.manager.PlayerExperienceLevel(uint64(playerID)))
}

//export csharp_set_player_experience_level
func csharp_set_player_experience_level(ctx C.uintptr_t, playerID C.uint64_t, value C.int32_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerExperienceLevel(uint64(playerID), int32(value)))
}

//export csharp_player_experience_progress
func csharp_player_experience_progress(ctx C.uintptr_t, playerID C.uint64_t) C.double {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return C.double(hostCtx.manager.PlayerExperienceProgress(uint64(playerID)))
}

//export csharp_set_player_experience_progress
func csharp_set_player_experience_progress(ctx C.uintptr_t, playerID C.uint64_t, value C.double) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerExperienceProgress(uint64(playerID), float64(value)))
}

//export csharp_player_on_ground
func csharp_player_on_ground(ctx C.uintptr_t, playerID C.uint64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.PlayerOnGround(uint64(playerID)))
}

//export csharp_player_sneaking
func csharp_player_sneaking(ctx C.uintptr_t, playerID C.uint64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.PlayerSneaking(uint64(playerID)))
}

//export csharp_set_player_sneaking
func csharp_set_player_sneaking(ctx C.uintptr_t, playerID C.uint64_t, value C.int) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerSneaking(uint64(playerID), value != 0))
}

//export csharp_player_sprinting
func csharp_player_sprinting(ctx C.uintptr_t, playerID C.uint64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.PlayerSprinting(uint64(playerID)))
}

//export csharp_set_player_sprinting
func csharp_set_player_sprinting(ctx C.uintptr_t, playerID C.uint64_t, value C.int) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerSprinting(uint64(playerID), value != 0))
}

//export csharp_player_swimming
func csharp_player_swimming(ctx C.uintptr_t, playerID C.uint64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.PlayerSwimming(uint64(playerID)))
}

//export csharp_set_player_swimming
func csharp_set_player_swimming(ctx C.uintptr_t, playerID C.uint64_t, value C.int) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerSwimming(uint64(playerID), value != 0))
}

//export csharp_player_flying
func csharp_player_flying(ctx C.uintptr_t, playerID C.uint64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.PlayerFlying(uint64(playerID)))
}

//export csharp_set_player_flying
func csharp_set_player_flying(ctx C.uintptr_t, playerID C.uint64_t, value C.int) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerFlying(uint64(playerID), value != 0))
}

//export csharp_player_gliding
func csharp_player_gliding(ctx C.uintptr_t, playerID C.uint64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.PlayerGliding(uint64(playerID)))
}

//export csharp_set_player_gliding
func csharp_set_player_gliding(ctx C.uintptr_t, playerID C.uint64_t, value C.int) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerGliding(uint64(playerID), value != 0))
}

//export csharp_player_crawling
func csharp_player_crawling(ctx C.uintptr_t, playerID C.uint64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.PlayerCrawling(uint64(playerID)))
}

//export csharp_set_player_crawling
func csharp_set_player_crawling(ctx C.uintptr_t, playerID C.uint64_t, value C.int) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerCrawling(uint64(playerID), value != 0))
}

//export csharp_player_using_item
func csharp_player_using_item(ctx C.uintptr_t, playerID C.uint64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.PlayerUsingItem(uint64(playerID)))
}

//export csharp_player_invisible
func csharp_player_invisible(ctx C.uintptr_t, playerID C.uint64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.PlayerInvisible(uint64(playerID)))
}

//export csharp_set_player_invisible
func csharp_set_player_invisible(ctx C.uintptr_t, playerID C.uint64_t, value C.int) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerInvisible(uint64(playerID), value != 0))
}

//export csharp_player_immobile
func csharp_player_immobile(ctx C.uintptr_t, playerID C.uint64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.PlayerImmobile(uint64(playerID)))
}

//export csharp_set_player_immobile
func csharp_set_player_immobile(ctx C.uintptr_t, playerID C.uint64_t, value C.int) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerImmobile(uint64(playerID), value != 0))
}

//export csharp_player_dead
func csharp_player_dead(ctx C.uintptr_t, playerID C.uint64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.PlayerDead(uint64(playerID)))
}

//export csharp_set_player_on_fire_millis
func csharp_set_player_on_fire_millis(ctx C.uintptr_t, playerID C.uint64_t, millis C.int64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerOnFireMillis(uint64(playerID), int64(millis)))
}

//export csharp_add_player_food
func csharp_add_player_food(ctx C.uintptr_t, playerID C.uint64_t, points C.int32_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.AddPlayerFood(uint64(playerID), int32(points)))
}

//export csharp_player_use_item
func csharp_player_use_item(ctx C.uintptr_t, playerID C.uint64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.PlayerUseItem(uint64(playerID)))
}

//export csharp_player_jump
func csharp_player_jump(ctx C.uintptr_t, playerID C.uint64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.PlayerJump(uint64(playerID)))
}

//export csharp_player_swing_arm
func csharp_player_swing_arm(ctx C.uintptr_t, playerID C.uint64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.PlayerSwingArm(uint64(playerID)))
}

//export csharp_player_wake
func csharp_player_wake(ctx C.uintptr_t, playerID C.uint64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.PlayerWake(uint64(playerID)))
}

//export csharp_player_extinguish
func csharp_player_extinguish(ctx C.uintptr_t, playerID C.uint64_t) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.PlayerExtinguish(uint64(playerID)))
}

//export csharp_set_player_show_coordinates
func csharp_set_player_show_coordinates(ctx C.uintptr_t, playerID C.uint64_t, value C.int) C.int {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return 0
	}
	return boolInt(hostCtx.manager.SetPlayerShowCoordinates(uint64(playerID), value != 0))
}

//export csharp_player_message
func csharp_player_message(ctx C.uintptr_t, playerID C.uint64_t, message *C.char) {
	hostCtx, ok := csharpHostContextByID(uintptr(ctx))
	if !ok || hostCtx.manager == nil {
		return
	}
	hostCtx.manager.PlayerMessage(uint64(playerID), goCString(message))
}

//export csharp_free_string
func csharp_free_string(value *C.char) {
	if value == nil {
		return
	}
	C.free(unsafe.Pointer(value))
}

//export csharp_free_bytes
func csharp_free_bytes(value *C.uint8_t) {
	if value == nil {
		return
	}
	C.free(unsafe.Pointer(value))
}
