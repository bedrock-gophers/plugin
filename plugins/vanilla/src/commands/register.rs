use bedrock_plugin::guest;

use super::gamemode;

pub fn register() {
    if !guest::register_command(
        "gamemode",
        "change your own or other people's gamemode",
        &["gm"],
        gamemode::GamemodeCommand,
    ) {
        guest::console_message("failed to register command \"gamemode\"");
    }
}
