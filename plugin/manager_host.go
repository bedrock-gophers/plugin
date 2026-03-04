package plugin

import (
	"log/slog"
	"math"
	"strings"
	"time"

	genoutput "github.com/bedrock-gophers/plugin/internal/generator/output"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

func playerValue[T any](m *Manager, playerID uint64, fallback T, fn func(*player.Player) T) T {
	p, ok := m.players.byID(playerID)
	if !ok {
		return fallback
	}
	return fn(p)
}

func (m *Manager) playerUpdate(playerID uint64, fn func(*player.Player)) bool {
	p, ok := m.players.byID(playerID)
	if !ok {
		return false
	}
	fn(p)
	return true
}

func (m *Manager) playerToggle(playerID uint64, enabled bool, on func(*player.Player), off func(*player.Player)) bool {
	return m.playerUpdate(playerID, func(p *player.Player) {
		if enabled {
			on(p)
			return
		}
		off(p)
	})
}

func (m *Manager) SetPlayerHealth(playerID uint64, target float64) bool {
	p, ok := m.players.byID(playerID)
	if !ok {
		return false
	}
	target = maxF64(target, 0)
	current := p.Health()
	if math.Abs(current-target) <= 0.00001 {
		return true
	}
	if target > current {
		p.Heal(target-current, pluginHealingSource{})
		return true
	}
	p.Hurt(current-target, pluginDamageSource{})
	return true
}

func (m *Manager) PlayerGameMode(playerID uint64) int32 {
	p, ok := m.players.byID(playerID)
	if !ok {
		return 0
	}
	id, ok := world.GameModeID(p.GameMode())
	if !ok {
		return 0
	}
	return int32(id)
}

func (m *Manager) SetPlayerGameMode(playerID uint64, modeID int32) bool {
	p, ok := m.players.byID(playerID)
	if !ok {
		return false
	}
	mode, ok := world.GameModeByID(int(modeID))
	if !ok {
		return false
	}
	p.SetGameMode(mode)
	return true
}

func (m *Manager) SetPlayerOnFireMillis(playerID uint64, millis int64) bool {
	if millis < 0 {
		millis = 0
	}
	return m.playerUpdate(playerID, func(p *player.Player) {
		p.SetOnFire(time.Duration(millis) * time.Millisecond)
	})
}

func (m *Manager) PlayerLatency(playerID uint64) time.Duration {
	return playerValue(m, playerID, time.Duration(0), func(p *player.Player) time.Duration {
		return p.Latency()
	})
}

func (m *Manager) ConsoleMessage(pluginName, message string) {
	if pluginName != "" {
		slog.Info(message, "plugin", pluginName)
		return
	}
	slog.Info(message)
}

func (m *Manager) ResolvePlayerByName(name string) uint64 {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0
	}
	if id := m.commandTargetID(name); id != 0 {
		return id
	}
	id, _, ok := m.players.byName(name)
	if !ok {
		return 0
	}
	return id
}

func (m *Manager) PlayerHandle(playerID uint64) uint64 {
	p, ok := m.players.byID(playerID)
	if !ok || p == nil {
		return 0
	}
	return genoutput.RegisterExternalObject(p)
}

func (m *Manager) OnlinePlayerNames() []string {
	return m.commandTargetNames()
}

func (m *Manager) EventCancel(_ uint64) bool {
	// Go plugins mutate cancellable events through MutableState directly.
	return false
}

func (m *Manager) EventSetItemDrop(_ uint64, _ ItemStackData) bool {
	// Go plugins mutate item drops through MutableState directly.
	return false
}

func maxF64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
