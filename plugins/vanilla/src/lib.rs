use std::os::raw::c_char;

use bedrock_plugin::guest::{self, NativeHostApi};

mod commands;

#[no_mangle]
pub extern "C" fn PluginLoad(host_api: *mut NativeHostApi, plugin_name: *const c_char) -> i32 {
    guest::plugin_load(host_api, plugin_name, commands::register)
}

#[no_mangle]
pub extern "C" fn PluginUnload() {
    guest::plugin_unload(|| {});
}

#[no_mangle]
pub extern "C" fn PluginDispatchEvent(
    version: u16,
    event_id: u16,
    flags: u32,
    player_id: u64,
    request_key: u64,
    payload: *const u8,
    payload_len: u32,
) {
    guest::plugin_dispatch_event(
        version,
        event_id,
        flags,
        player_id,
        request_key,
        payload,
        payload_len,
    );
}
