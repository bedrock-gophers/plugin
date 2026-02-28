package guest

import (
	"fmt"
	"strings"
	"sync"

	"github.com/bedrock-gophers/plugin/plugin/abi"
)

var (
	commandHandlersMu    sync.RWMutex
	commandHandlers             = map[uint32]func(source CommandSource, args []string){}
	nextCommandHandlerID uint32 = 1
)

type commandParameterKind uint8

const (
	commandParameterString commandParameterKind = iota
	commandParameterText
	commandParameterEnum
	commandParameterSubcommand
	commandParameterPluginAvailable
	commandParameterPluginLoaded
)

type commandParameterSpec struct {
	name        string
	kind        commandParameterKind
	optional    bool
	enumOptions []string
}

type commandOverloadSpec struct {
	parameters []commandParameterSpec
}

func init() {
	handle(abi.EventPluginCommand, dispatchPluginCommand)
}

// CommandSource is the source of a command execution.
// It may either be a player or the console.
type CommandSource struct {
	playerID   uint64
	pluginName string
}

func (s CommandSource) IsConsole() bool {
	return s.playerID == 0
}

func (s CommandSource) IsPlayer() bool {
	return s.playerID != 0
}

func (s CommandSource) Player() (PlayerRef, bool) {
	if s.playerID == 0 {
		return PlayerRef{}, false
	}
	return PlayerRef{id: s.playerID}, true
}

func (s CommandSource) Message(message string) {
	if s.playerID == 0 {
		consoleMessageForPlugin(s.pluginName, message)
		return
	}
	playerMessage(s.playerID, message)
}

// HandleCommand registers a low-level command handler owned by the plugin.
func (baseEvents) HandleCommand(name, description string, aliases []string, fn func(source CommandSource, args []string)) {
	registerCommandHandler(name, description, aliases, nil, fn)
}

func registerCommandHandler(name, description string, aliases []string, overloads []commandOverloadSpec, fn func(source CommandSource, args []string)) {
	if fn == nil {
		panic("guest.Base.HandleCommand: handler cannot be nil")
	}
	name = normalizeCommandToken(name)
	if name == "" {
		panic("guest.Base.HandleCommand: command name must be non-empty and contain no spaces")
	}

	description = strings.TrimSpace(description)
	aliases = normalizeCommandAliases(aliases, name)

	commandHandlersMu.Lock()
	handlerID := nextCommandHandlerID
	nextCommandHandlerID++
	commandHandlers[handlerID] = fn
	commandHandlersMu.Unlock()

	if !registerCommand(name, description, aliases, handlerID, overloads) {
		commandHandlersMu.Lock()
		delete(commandHandlers, handlerID)
		commandHandlersMu.Unlock()
		panic(fmt.Sprintf("guest.Base.HandleCommand: failed to register command %q", name))
	}
}

func dispatchPluginCommand(ev *Event) {
	d := ev.Decoder()
	handlerID := d.U32()
	argCount := int(d.U32())
	args := make([]string, 0, argCount)
	for i := 0; i < argCount; i++ {
		args = append(args, d.String())
	}
	if !d.Ok() {
		return
	}
	source := CommandSource{playerID: ev.PlayerID(), pluginName: ev.PluginName()}
	commandHandlersMu.RLock()
	fn, ok := commandHandlers[handlerID]
	commandHandlersMu.RUnlock()
	if ok {
		fn(source, args)
	}
}

func normalizeCommandAliases(aliases []string, commandName string) []string {
	out := make([]string, 0, len(aliases))
	seen := map[string]struct{}{commandName: {}}
	for _, alias := range aliases {
		token := normalizeCommandToken(alias)
		if token == "" {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		out = append(out, token)
	}
	return out
}

func normalizeCommandToken(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" || strings.ContainsAny(v, " \t\r\n") {
		return ""
	}
	return v
}
