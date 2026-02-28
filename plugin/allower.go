package plugin

import (
	"net"
	"sync"

	"github.com/df-mc/dragonfly/server"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

// ServerAllower bridges Dragonfly's server-level allower to plugin join events.
type ServerAllower struct {
	base server.Allower

	mu  sync.RWMutex
	mgr *Manager
}

func NewServerAllower(base server.Allower) *ServerAllower {
	return &ServerAllower{base: base}
}

func (a *ServerAllower) SetManager(mgr *Manager) {
	a.mu.Lock()
	a.mgr = mgr
	a.mu.Unlock()
}

func (a *ServerAllower) Allow(addr net.Addr, d login.IdentityData, c login.ClientData) (string, bool) {
	if a.base != nil {
		msg, ok := a.base.Allow(addr, d, c)
		if !ok {
			return msg, false
		}
	}

	a.mu.RLock()
	mgr := a.mgr
	a.mu.RUnlock()
	if mgr == nil {
		return "", true
	}
	return mgr.AllowJoin(addr, d, c)
}
