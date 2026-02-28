package plugin

import (
	"reflect"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type pluginCommandParamKind uint8

const (
	pluginCommandParamString pluginCommandParamKind = iota
	pluginCommandParamText
	pluginCommandParamEnum
	pluginCommandParamSubcommand
)

type pluginCommandParamSpec struct {
	name        string
	kind        pluginCommandParamKind
	optional    bool
	enumOptions []string
}

type pluginCommandOverloadSpec struct {
	parameters []pluginCommandParamSpec
}

type pluginCommandArgs []string

func (pluginCommandArgs) Parse(line *cmd.Line, v reflect.Value) error {
	v.Set(reflect.ValueOf(pluginCommandArgs(line.Leftover())))
	return nil
}

func (pluginCommandArgs) Type() string {
	return "text"
}

type pluginCommandRunnable struct {
	Args pluginCommandArgs `cmd:"args"`

	manager       *Manager                   `cmd:"-"`
	entry         *pluginCommandRegistration `cmd:"-"`
	overload      pluginCommandOverloadSpec  `cmd:"-"`
	overloadIndex int                        `cmd:"-"`
}

func (r pluginCommandRunnable) Run(src cmd.Source, o *cmd.Output, _ *world.Tx) {
	if r.manager == nil || r.entry == nil {
		o.Error("plugin command is unavailable")
		return
	}
	var playerID uint64
	if p, ok := src.(*player.Player); ok {
		playerID = r.manager.players.ensure(p)
	}
	if err := r.manager.executePluginCommand(r.entry, playerID, []string(r.Args)); err != nil {
		o.Error(err)
	}
}

func (r pluginCommandRunnable) DescribeParams(cmd.Source) []cmd.ParamInfo {
	if len(r.overload.parameters) == 0 {
		return []cmd.ParamInfo{{Name: "args", Value: cmd.Varargs(""), Optional: true}}
	}
	out := make([]cmd.ParamInfo, 0, len(r.overload.parameters))
	for _, param := range r.overload.parameters {
		value := any("")
		switch param.kind {
		case pluginCommandParamText:
			value = cmd.Varargs("")
		case pluginCommandParamEnum:
			value = pluginCommandEnum{
				typ:     pluginEnumTypeName(r.entry.name, param.name),
				options: append([]string(nil), param.enumOptions...),
			}
		case pluginCommandParamSubcommand:
			value = cmd.SubCommand{}
		}
		out = append(out, cmd.ParamInfo{
			Name:     param.name,
			Value:    value,
			Optional: param.optional,
		})
	}
	return out
}

type pluginCommandEnum struct {
	typ     string
	options []string
}

func (e pluginCommandEnum) Type() string {
	return e.typ
}

func (e pluginCommandEnum) Options(cmd.Source) []string {
	return append([]string(nil), e.options...)
}

func pluginEnumTypeName(commandName, paramName string) string {
	commandToken := normalizePluginCommandToken(commandName)
	paramToken := normalizePluginCommandToken(paramName)
	if paramToken == "" {
		paramToken = "value"
	}
	if commandToken == "" {
		return paramToken
	}
	return commandToken + "_" + paramToken
}

func isPluginRuntimeCommand(command cmd.Command) bool {
	for _, runnable := range command.Runnables(pluginCommandInspectSource{}) {
		if _, ok := runnable.(pluginCommandRunnable); ok {
			return true
		}
	}
	return false
}

type pluginCommandInspectSource struct{}

func (pluginCommandInspectSource) Position() mgl64.Vec3 { return mgl64.Vec3{} }

func (pluginCommandInspectSource) SendCommandOutput(*cmd.Output) {}
