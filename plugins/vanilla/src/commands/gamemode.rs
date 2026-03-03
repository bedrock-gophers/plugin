use bedrock_plugin::guest::{
    self, Command, CommandContext, CommandOverloadSpec, CommandParameterSpec, GameMode, PlayerRef,
};

pub const GAMEMODE_SURVIVAL: &str = "survival";
pub const GAMEMODE_CREATIVE: &str = "creative";
pub const GAMEMODE_ADVENTURE: &str = "adventure";
pub const GAMEMODE_SPECTATOR: &str = "spectator";

pub struct GamemodeCommand;

impl Command for GamemodeCommand {
    fn overloads(&self) -> Vec<CommandOverloadSpec> {
        vec![CommandOverloadSpec::new(vec![
            CommandParameterSpec::enum_options(
                "gamemode",
                &[
                    GAMEMODE_SURVIVAL,
                    GAMEMODE_CREATIVE,
                    GAMEMODE_ADVENTURE,
                    GAMEMODE_SPECTATOR,
                ],
                false,
            ),
            CommandParameterSpec::target("player", true),
        ])]
    }

    fn run(&self, ctx: CommandContext, args: &[String]) {
        if args.is_empty() || args.len() > 2 {
            ctx.message("usage: /gamemode <gamemode> [player]");
            return;
        }

        let mode_token = args[0].trim().to_ascii_lowercase();
        let Some(mode) = to_guest_gamemode(&mode_token) else {
            ctx.message(&format!("invalid gamemode \"{}\"", mode_token));
            return;
        };

        let Some(target_player) = resolve_target_player(ctx, args) else {
            return;
        };

        if !target_player.set_game_mode(mode) {
            ctx.message("failed to set gamemode");
            return;
        }

        let target_name = {
            let candidate = target_player.name();
            if candidate.trim().is_empty() {
                "target".to_string()
            } else {
                candidate
            }
        };
        ctx.message(&format!(
            "set gamemode of {} to {}",
            target_name, mode_token
        ));

        let source_player_id = ctx.player().map(PlayerRef::id).unwrap_or(0);
        if source_player_id == 0 || source_player_id != target_player.id() {
            target_player.message(&format!("your gamemode was set to {}", mode_token));
        }
    }
}

fn resolve_target_player(ctx: CommandContext, args: &[String]) -> Option<PlayerRef> {
    if args.len() < 2 {
        let Some(source_player) = ctx.player() else {
            ctx.message("console must provide a player target");
            return None;
        };
        return Some(source_player);
    }

    let target = args[1].trim();
    if target.is_empty() {
        ctx.message("target must be non-empty");
        return None;
    }

    let Some(player) = guest::player_by_name(target) else {
        ctx.message(&format!("cannot resolve online player \"{}\"", target));
        return None;
    };
    Some(player)
}

fn to_guest_gamemode(mode: &str) -> Option<GameMode> {
    match mode {
        GAMEMODE_SURVIVAL => Some(GameMode::Survival),
        GAMEMODE_CREATIVE => Some(GameMode::Creative),
        GAMEMODE_ADVENTURE => Some(GameMode::Adventure),
        GAMEMODE_SPECTATOR => Some(GameMode::Spectator),
        _ => None,
    }
}
