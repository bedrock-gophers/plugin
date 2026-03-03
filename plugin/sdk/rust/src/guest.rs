use std::collections::{HashMap, HashSet};
use std::ffi::{c_char, c_int, CStr, CString};
use std::panic::{catch_unwind, AssertUnwindSafe};
use std::sync::{Arc, Mutex, OnceLock};

use crate::abi::{Decoder, Encoder, EVENT_PLUGIN_COMMAND};

#[repr(C)]
pub struct NativeHostApi {
    pub ctx: usize,

    pub register_command: Option<
        unsafe extern "C" fn(
            ctx: usize,
            plugin_name: *mut c_char,
            name: *mut c_char,
            description: *mut c_char,
            aliases_csv: *mut c_char,
            handler_id: u32,
            overloads: *mut u8,
            overloads_len: u32,
        ) -> c_int,
    >,
    pub manage_plugins: Option<
        unsafe extern "C" fn(
            ctx: usize,
            action: u32,
            target: *mut c_char,
            out_len: *mut u32,
        ) -> *mut u8,
    >,
    pub resolve_player_by_name: Option<unsafe extern "C" fn(ctx: usize, name: *mut c_char) -> u64>,
    pub online_player_names: Option<unsafe extern "C" fn(ctx: usize, out_len: *mut u32) -> *mut u8>,
    pub console_message:
        Option<unsafe extern "C" fn(ctx: usize, plugin_name: *mut c_char, message: *mut c_char)>,
    pub host_call: Option<
        unsafe extern "C" fn(
            ctx: usize,
            op: u32,
            payload: *mut u8,
            payload_len: u32,
            out_len: *mut u32,
        ) -> *mut u8,
    >,
    pub event_cancel: Option<unsafe extern "C" fn(ctx: usize, request_key: u64) -> c_int>,
    pub event_item_drop_set: Option<
        unsafe extern "C" fn(
            ctx: usize,
            request_key: u64,
            payload: *mut u8,
            payload_len: u32,
        ) -> c_int,
    >,

    pub player_health: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> f64>,
    pub set_player_health:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, health: f64) -> c_int>,
    pub player_food: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> i32>,
    pub set_player_food:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, food: i32) -> c_int>,
    pub player_name: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> *mut c_char>,
    pub player_game_mode: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> i32>,
    pub set_player_game_mode:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, mode: i32) -> c_int>,
    pub player_xuid: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> *mut c_char>,
    pub player_device_id: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> *mut c_char>,
    pub player_device_model:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> *mut c_char>,
    pub player_self_signed_id:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> *mut c_char>,
    pub player_name_tag: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> *mut c_char>,
    pub set_player_name_tag:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: *mut c_char) -> c_int>,
    pub player_score_tag: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> *mut c_char>,
    pub set_player_score_tag:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: *mut c_char) -> c_int>,
    pub player_absorption: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> f64>,
    pub set_player_absorption:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: f64) -> c_int>,
    pub player_max_health: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> f64>,
    pub set_player_max_health:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: f64) -> c_int>,
    pub player_speed: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> f64>,
    pub set_player_speed:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: f64) -> c_int>,
    pub player_flight_speed: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> f64>,
    pub set_player_flight_speed:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: f64) -> c_int>,
    pub player_vertical_flight_speed:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> f64>,
    pub set_player_vertical_flight_speed:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: f64) -> c_int>,
    pub player_experience: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> i32>,
    pub player_experience_level: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> i32>,
    pub set_player_experience_level:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: i32) -> c_int>,
    pub player_experience_progress: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> f64>,
    pub set_player_experience_progress:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: f64) -> c_int>,
    pub player_on_ground: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> c_int>,
    pub player_sneaking: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> c_int>,
    pub set_player_sneaking:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: c_int) -> c_int>,
    pub player_sprinting: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> c_int>,
    pub set_player_sprinting:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: c_int) -> c_int>,
    pub player_swimming: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> c_int>,
    pub set_player_swimming:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: c_int) -> c_int>,
    pub player_flying: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> c_int>,
    pub set_player_flying:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: c_int) -> c_int>,
    pub player_gliding: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> c_int>,
    pub set_player_gliding:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: c_int) -> c_int>,
    pub player_crawling: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> c_int>,
    pub set_player_crawling:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: c_int) -> c_int>,
    pub player_using_item: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> c_int>,
    pub player_invisible: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> c_int>,
    pub set_player_invisible:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: c_int) -> c_int>,
    pub player_immobile: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> c_int>,
    pub set_player_immobile:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: c_int) -> c_int>,
    pub player_dead: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> c_int>,
    pub set_player_on_fire_millis:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, millis: i64) -> c_int>,
    pub add_player_food:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, points: i32) -> c_int>,
    pub player_use_item: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> c_int>,
    pub player_jump: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> c_int>,
    pub player_swing_arm: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> c_int>,
    pub player_wake: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> c_int>,
    pub player_extinguish: Option<unsafe extern "C" fn(ctx: usize, player_id: u64) -> c_int>,
    pub set_player_show_coordinates:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, value: c_int) -> c_int>,
    pub player_message:
        Option<unsafe extern "C" fn(ctx: usize, player_id: u64, message: *mut c_char)>,

    pub free_string: Option<unsafe extern "C" fn(value: *mut c_char)>,
    pub free_bytes: Option<unsafe extern "C" fn(value: *mut u8)>,
}

#[repr(u8)]
#[derive(Clone, Copy)]
pub enum CommandParameterKind {
    String = 0,
    Text = 1,
    Enum = 2,
    Subcommand = 3,
    PluginAvailable = 4,
    PluginLoaded = 5,
    Target = 6,
}

#[derive(Clone)]
pub struct CommandParameterSpec {
    pub name: String,
    pub kind: CommandParameterKind,
    pub optional: bool,
    pub enum_options: Vec<String>,
}

impl CommandParameterSpec {
    pub fn string(name: &str, optional: bool) -> Self {
        Self {
            name: name.to_string(),
            kind: CommandParameterKind::String,
            optional,
            enum_options: Vec::new(),
        }
    }

    pub fn text(name: &str, optional: bool) -> Self {
        Self {
            name: name.to_string(),
            kind: CommandParameterKind::Text,
            optional,
            enum_options: Vec::new(),
        }
    }

    pub fn target(name: &str, optional: bool) -> Self {
        Self {
            name: name.to_string(),
            kind: CommandParameterKind::Target,
            optional,
            enum_options: Vec::new(),
        }
    }

    pub fn enum_options(name: &str, options: &[&str], optional: bool) -> Self {
        Self {
            name: name.to_string(),
            kind: CommandParameterKind::Enum,
            optional,
            enum_options: options.iter().map(|v| v.to_string()).collect(),
        }
    }
}

#[derive(Clone)]
pub struct CommandOverloadSpec {
    pub parameters: Vec<CommandParameterSpec>,
}

impl CommandOverloadSpec {
    pub fn new(parameters: Vec<CommandParameterSpec>) -> Self {
        Self { parameters }
    }
}

#[derive(Clone, Copy)]
pub enum GameMode {
    Survival = 0,
    Creative = 1,
    Adventure = 2,
    Spectator = 3,
}

#[derive(Clone, Copy)]
pub struct PlayerRef {
    id: u64,
}

impl PlayerRef {
    pub fn id(self) -> u64 {
        self.id
    }

    pub fn name(self) -> String {
        let Some(snapshot) = runtime_snapshot() else {
            return String::new();
        };
        // SAFETY: host_api comes from PluginLoad and is valid while runtime is active.
        unsafe {
            let api = &*(snapshot.host_api as *mut NativeHostApi);
            let Some(player_name) = api.player_name else {
                return String::new();
            };
            let raw = player_name(snapshot.ctx, self.id);
            if raw.is_null() {
                return String::new();
            }
            let value = cstr_to_string(raw.cast_const());
            if let Some(free_string) = api.free_string {
                free_string(raw);
            }
            value
        }
    }

    pub fn set_game_mode(self, mode: GameMode) -> bool {
        let Some(snapshot) = runtime_snapshot() else {
            return false;
        };
        // SAFETY: host_api comes from PluginLoad and is valid while runtime is active.
        unsafe {
            let api = &*(snapshot.host_api as *mut NativeHostApi);
            let Some(set_player_game_mode) = api.set_player_game_mode else {
                return false;
            };
            set_player_game_mode(snapshot.ctx, self.id, mode as i32) != 0
        }
    }

    pub fn message(self, message: &str) {
        let Some(snapshot) = runtime_snapshot() else {
            return;
        };
        send_player_message(&snapshot, self.id, message);
    }
}

#[derive(Clone, Copy)]
pub struct CommandSource {
    source_player_id: u64,
}

impl CommandSource {
    pub fn is_console(self) -> bool {
        self.source_player_id == 0
    }

    pub fn is_player(self) -> bool {
        self.source_player_id != 0
    }

    pub fn player(self) -> Option<PlayerRef> {
        if self.source_player_id == 0 {
            return None;
        }
        Some(PlayerRef {
            id: self.source_player_id,
        })
    }

    pub fn message(self, message: &str) {
        let Some(snapshot) = runtime_snapshot() else {
            return;
        };
        if self.source_player_id == 0 {
            send_console_message(&snapshot, &snapshot.plugin_name, message);
            return;
        }
        send_player_message(&snapshot, self.source_player_id, message);
    }
}

#[derive(Clone, Copy)]
pub struct CommandContext {
    source: CommandSource,
}

impl CommandContext {
    pub fn source(self) -> CommandSource {
        self.source
    }

    pub fn player(self) -> Option<PlayerRef> {
        self.source.player()
    }

    pub fn message(self, message: &str) {
        self.source.message(message);
    }
}

pub trait Command: Send + Sync {
    fn overloads(&self) -> Vec<CommandOverloadSpec>;
    fn run(&self, ctx: CommandContext, args: &[String]);
    fn allow(&self, _source: CommandSource) -> bool {
        true
    }
}

pub type RawCommandHandler = fn(CommandContext, &[String]);

struct RawCommand {
    overloads: Vec<CommandOverloadSpec>,
    handler: RawCommandHandler,
}

impl Command for RawCommand {
    fn overloads(&self) -> Vec<CommandOverloadSpec> {
        self.overloads.clone()
    }

    fn run(&self, ctx: CommandContext, args: &[String]) {
        (self.handler)(ctx, args);
    }
}

struct Runtime {
    host_api: usize,
    ctx: usize,
    plugin_name: String,
    handlers: HashMap<u32, Arc<dyn Command>>,
    next_handler_id: u32,
}

#[derive(Clone)]
struct RuntimeSnapshot {
    host_api: usize,
    ctx: usize,
    plugin_name: String,
}

fn runtime_slot() -> &'static Mutex<Option<Runtime>> {
    static SLOT: OnceLock<Mutex<Option<Runtime>>> = OnceLock::new();
    SLOT.get_or_init(|| Mutex::new(None))
}

fn runtime_snapshot() -> Option<RuntimeSnapshot> {
    let guard = runtime_slot().lock().ok()?;
    let runtime = guard.as_ref()?;
    Some(RuntimeSnapshot {
        host_api: runtime.host_api,
        ctx: runtime.ctx,
        plugin_name: runtime.plugin_name.clone(),
    })
}

pub fn plugin_load(host_api: *mut NativeHostApi, plugin_name: *const c_char, init: fn()) -> c_int {
    if host_api.is_null() {
        return 0;
    }

    let mut name = cstr_to_string(plugin_name);
    if name.trim().is_empty() {
        name = "rust".to_string();
    }

    {
        let mut guard = match runtime_slot().lock() {
            Ok(value) => value,
            Err(_) => return 0,
        };
        // SAFETY: host_api was checked for null above.
        let ctx = unsafe { (*host_api).ctx };
        *guard = Some(Runtime {
            host_api: host_api as usize,
            ctx,
            plugin_name: name,
            handlers: HashMap::new(),
            next_handler_id: 1,
        });
    }

    if catch_unwind(AssertUnwindSafe(init)).is_ok() {
        return 1;
    }

    console_message("PluginLoad failed: init panicked");
    if let Ok(mut guard) = runtime_slot().lock() {
        *guard = None;
    }
    0
}

pub fn plugin_unload(unload: fn()) {
    let _ = catch_unwind(AssertUnwindSafe(unload));
    if let Ok(mut guard) = runtime_slot().lock() {
        *guard = None;
    }
}

#[allow(clippy::too_many_arguments)]
pub fn plugin_dispatch_event(
    _version: u16,
    event_id: u16,
    _flags: u32,
    player_id: u64,
    _request_key: u64,
    payload: *const u8,
    payload_len: u32,
) {
    if event_id != EVENT_PLUGIN_COMMAND {
        return;
    }

    let payload_bytes: &[u8] = if payload.is_null() || payload_len == 0 {
        &[]
    } else {
        // SAFETY: payload pointer/length are provided by host for this call.
        unsafe { std::slice::from_raw_parts(payload, payload_len as usize) }
    };

    let mut dec = Decoder::new(payload_bytes);
    let handler_id = dec.u32();
    let arg_count = dec.u32() as usize;
    if arg_count > 1024 {
        return;
    }
    let mut args = Vec::with_capacity(arg_count);
    for _ in 0..arg_count {
        args.push(dec.string());
    }
    if !dec.ok() {
        return;
    }

    let command = {
        let guard = match runtime_slot().lock() {
            Ok(value) => value,
            Err(_) => return,
        };
        let Some(runtime) = guard.as_ref() else {
            return;
        };
        runtime.handlers.get(&handler_id).cloned()
    };

    let Some(command) = command else {
        return;
    };

    let source = CommandSource {
        source_player_id: player_id,
    };
    if !command.allow(source) {
        source.message("you are not allowed to use this command");
        return;
    }

    let ctx = CommandContext { source };
    if catch_unwind(AssertUnwindSafe(|| command.run(ctx, &args))).is_err() {
        console_message("command handler panicked");
    }
}

pub fn register_command<C>(name: &str, description: &str, aliases: &[&str], command: C) -> bool
where
    C: Command + 'static,
{
    register_command_arc(name, description, aliases, Arc::new(command))
}

pub fn register_raw_command(
    name: &str,
    description: &str,
    aliases: &[&str],
    overloads: &[CommandOverloadSpec],
    handler: RawCommandHandler,
) -> bool {
    register_command_arc(
        name,
        description,
        aliases,
        Arc::new(RawCommand {
            overloads: overloads.to_vec(),
            handler,
        }),
    )
}

fn register_command_arc(
    name: &str,
    description: &str,
    aliases: &[&str],
    command: Arc<dyn Command>,
) -> bool {
    let command_name = normalize_command_token(name);
    if command_name.is_empty() {
        return false;
    }
    let normalized_aliases = normalize_aliases(aliases, &command_name);
    let aliases_csv = normalized_aliases.join(",");
    let overloads = command.overloads();
    let mut overload_bytes = encode_overloads(&overloads);

    let (host_api, ctx, plugin_name, handler_id) = {
        let mut guard = match runtime_slot().lock() {
            Ok(value) => value,
            Err(_) => return false,
        };
        let Some(runtime) = guard.as_mut() else {
            return false;
        };

        let id = runtime.next_handler_id;
        runtime.next_handler_id = runtime.next_handler_id.saturating_add(1);
        runtime.handlers.insert(id, command.clone());
        (
            runtime.host_api,
            runtime.ctx,
            runtime.plugin_name.clone(),
            id,
        )
    };

    let ok = unsafe {
        let api = &*(host_api as *mut NativeHostApi);
        let Some(register) = api.register_command else {
            return false;
        };
        let plugin_name_c = cstring(&plugin_name);
        let command_name_c = cstring(&command_name);
        let description_c = cstring(description.trim());
        let aliases_csv_c = cstring(&aliases_csv);
        let overload_ptr = if overload_bytes.is_empty() {
            std::ptr::null_mut()
        } else {
            overload_bytes.as_mut_ptr()
        };
        register(
            ctx,
            plugin_name_c.as_ptr().cast_mut(),
            command_name_c.as_ptr().cast_mut(),
            description_c.as_ptr().cast_mut(),
            aliases_csv_c.as_ptr().cast_mut(),
            handler_id,
            overload_ptr,
            overload_bytes.len() as u32,
        ) != 0
    };

    if ok {
        return true;
    }

    if let Ok(mut guard) = runtime_slot().lock() {
        if let Some(runtime) = guard.as_mut() {
            runtime.handlers.remove(&handler_id);
        }
    }
    false
}

pub fn player_by_name(name: &str) -> Option<PlayerRef> {
    let candidate = name.trim();
    if candidate.is_empty() {
        return None;
    }
    let snapshot = runtime_snapshot()?;
    // SAFETY: host_api comes from PluginLoad and is valid while runtime is active.
    let id = unsafe {
        let api = &*(snapshot.host_api as *mut NativeHostApi);
        let Some(resolve_player_by_name) = api.resolve_player_by_name else {
            return None;
        };
        let candidate_c = cstring(candidate);
        resolve_player_by_name(snapshot.ctx, candidate_c.as_ptr().cast_mut())
    };
    if id == 0 {
        return None;
    }
    Some(PlayerRef { id })
}

pub fn console_message(message: &str) {
    let Some(snapshot) = runtime_snapshot() else {
        return;
    };
    send_console_message(&snapshot, &snapshot.plugin_name, message);
}

fn encode_overloads(overloads: &[CommandOverloadSpec]) -> Vec<u8> {
    let mut enc = Encoder::with_capacity(64 + overloads.len() * 24);
    enc.u32(overloads.len() as u32);
    for overload in overloads {
        enc.u32(overload.parameters.len() as u32);
        for parameter in &overload.parameters {
            enc.string(&parameter.name);
            enc.u8(parameter.kind as u8);
            enc.bool(parameter.optional);
            enc.u32(parameter.enum_options.len() as u32);
            for option in &parameter.enum_options {
                enc.string(option);
            }
        }
    }
    enc.into_bytes()
}

fn normalize_command_token(value: &str) -> String {
    let token = value.trim().to_ascii_lowercase();
    if token.is_empty() || token.chars().any(char::is_whitespace) {
        return String::new();
    }
    token
}

fn normalize_aliases(aliases: &[&str], command_name: &str) -> Vec<String> {
    let mut seen = HashSet::new();
    seen.insert(command_name.to_string());

    let mut out = Vec::new();
    for alias in aliases {
        let normalized = normalize_command_token(alias);
        if normalized.is_empty() {
            continue;
        }
        if seen.insert(normalized.clone()) {
            out.push(normalized);
        }
    }
    out
}

fn cstring(value: &str) -> CString {
    match CString::new(value) {
        Ok(v) => v,
        Err(_) => CString::new(value.replace('\0', "")).unwrap_or_default(),
    }
}

fn cstr_to_string(ptr: *const c_char) -> String {
    if ptr.is_null() {
        return String::new();
    }
    // SAFETY: caller provides a NUL-terminated string pointer or null.
    unsafe { CStr::from_ptr(ptr).to_string_lossy().into_owned() }
}

fn send_console_message(snapshot: &RuntimeSnapshot, plugin_name: &str, message: &str) {
    // SAFETY: host_api comes from PluginLoad and is valid while runtime is active.
    unsafe {
        let api = &*(snapshot.host_api as *mut NativeHostApi);
        let Some(console_message) = api.console_message else {
            return;
        };
        let plugin_name_c = cstring(plugin_name);
        let message_c = cstring(message);
        console_message(
            snapshot.ctx,
            plugin_name_c.as_ptr().cast_mut(),
            message_c.as_ptr().cast_mut(),
        );
    }
}

fn send_player_message(snapshot: &RuntimeSnapshot, player_id: u64, message: &str) {
    // SAFETY: host_api comes from PluginLoad and is valid while runtime is active.
    unsafe {
        let api = &*(snapshot.host_api as *mut NativeHostApi);
        let Some(player_message) = api.player_message else {
            return;
        };
        let message_c = cstring(message);
        player_message(snapshot.ctx, player_id, message_c.as_ptr().cast_mut());
    }
}
