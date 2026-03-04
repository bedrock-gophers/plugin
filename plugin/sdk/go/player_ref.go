package guest

import (
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/text"
)

type GameMode int32

const (
	GameModeSurvival  GameMode = 0
	GameModeCreative  GameMode = 1
	GameModeAdventure GameMode = 2
	GameModeSpectator GameMode = 3
)

type PlayerRef struct {
	id uint64
}

func (p PlayerRef) ID() uint64 {
	return p.id
}

func PlayerByName(name string) (PlayerRef, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return PlayerRef{}, false
	}
	id := hostValue(uint64(0), func(h Host) uint64 { return h.ResolvePlayerByName(name) })
	if id == 0 {
		return PlayerRef{}, false
	}
	return PlayerRef{id: id}, true
}

// Target is a built-in command enum for online player names.
// It can be used as a command field type directly or inside Optional[T].
type Target string

func (Target) Type() string {
	return "target"
}

func (Target) Options(_ CommandSource) []string {
	return hostValue([]string(nil), func(h Host) []string {
		names := h.OnlinePlayerNames()
		return append([]string(nil), names...)
	})
}

func (t Target) Player() (PlayerRef, bool) {
	return PlayerByName(string(t))
}

func (p PlayerRef) Messagef(format string, a ...any) {
	p.Message(text.Colourf(format, a...))
}
