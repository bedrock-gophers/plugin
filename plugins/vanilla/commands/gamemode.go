package commands

import (
	"github.com/bedrock-gophers/plugin/plugin/sdk/go"
)

const (
	gamemodeSurvival  = "survival"
	gamemodeCreative  = "creative"
	gamemodeAdventure = "adventure"
	gamemodeSpectator = "spectator"
)

type Gamemode struct {
	Gamemode enumGamemodes                `cmd:"gamemode"`
	Target   guest.Optional[guest.Target] `cmd:"player"`
}

func (c Gamemode) Run(ctx guest.Context) {
	var targetPlayer guest.PlayerRef
	target, ok := c.Target.Load()
	if !ok {
		sourcePlayer, isPlayer := ctx.Player()
		if !isPlayer {
			ctx.Message("console must provide a player target")
			return
		}
		targetPlayer = sourcePlayer
	} else {
		resolved, found := target.Player()
		if !found {
			ctx.Messagef("cannot resolve online player %q", target)
			return
		}
		targetPlayer = resolved
	}

	mode, ok := toGuestGamemode(c.Gamemode)
	if !ok {
		ctx.Messagef("invalid gamemode %q", c.Gamemode)
		return
	}
	if !targetPlayer.SetGameMode(mode) {
		ctx.Message("failed to set gamemode")
		return
	}
	ctx.Messagef("set gamemode of %s to %s", targetPlayer.Name(), c.Gamemode)

	sourcePlayer, isPlayer := ctx.Player()
	if !isPlayer || sourcePlayer.ID() != targetPlayer.ID() {
		targetPlayer.Messagef("your gamemode was set to %s", c.Gamemode)
	}
}

type enumGamemodes string

func (enumGamemodes) Type() string {
	return "gamemode"
}

func (enumGamemodes) Options(_ guest.CommandSource) []string {
	return []string{gamemodeSurvival, gamemodeCreative, gamemodeAdventure, gamemodeSpectator}
}

func toGuestGamemode(mode enumGamemodes) (guest.GameMode, bool) {
	switch mode {
	case gamemodeSurvival:
		return guest.GameModeSurvival, true
	case gamemodeCreative:
		return guest.GameModeCreative, true
	case gamemodeAdventure:
		return guest.GameModeAdventure, true
	case gamemodeSpectator:
		return guest.GameModeSpectator, true
	default:
		return 0, false
	}
}
