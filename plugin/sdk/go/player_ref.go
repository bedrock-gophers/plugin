package guest

import (
	"strings"
	"time"
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

// Latency returns the rolling player latency as a duration.
// If unavailable, it returns 0.
func (p PlayerRef) Latency() time.Duration {
	type latencyHost interface {
		PlayerLatencyMillis(playerID uint64) int64
	}
	return hostValue(time.Duration(0), func(h Host) time.Duration {
		lh, ok := any(h).(latencyHost)
		if !ok {
			return 0
		}
		ms := lh.PlayerLatencyMillis(p.id)
		if ms < 0 {
			return 0
		}
		return time.Duration(ms) * time.Millisecond
	})
}
