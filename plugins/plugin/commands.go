package main

import (
	"strings"

	guest "github.com/bedrock-gophers/plugin/plugin/sdk/go"
)

type pluginListCommand struct {
	List guest.SubCommand `cmd:"list"`
}

type pluginLoadCommand struct {
	Load   guest.SubCommand       `cmd:"load"`
	Target guest.Optional[string] `cmd:"target"`
}

type pluginUnloadCommand struct {
	Unload guest.SubCommand       `cmd:"unload"`
	Target guest.Optional[string] `cmd:"target"`
}

type pluginReloadCommand struct {
	Reload guest.SubCommand       `cmd:"reload"`
	Target guest.Optional[string] `cmd:"target"`
}

func (pluginListCommand) Allow(source guest.CommandSource) bool {
	return source.IsConsole()
}

func (pluginLoadCommand) Allow(source guest.CommandSource) bool {
	return source.IsConsole()
}

func (pluginUnloadCommand) Allow(source guest.CommandSource) bool {
	return source.IsConsole()
}

func (pluginReloadCommand) Allow(source guest.CommandSource) bool {
	return source.IsConsole()
}

func (pluginListCommand) Run(ctx guest.Context) {
	names, err := guest.ListPlugins()
	if err != nil {
		ctx.Messagef("%s list failed: %s", pluginPrefix, err)
		return
	}
	if len(names) == 0 {
		ctx.Messagef("%s no plugins loaded", pluginPrefix)
		return
	}
	ctx.Messagef("%s loaded %d plugin(s): %s", pluginPrefix, len(names), strings.Join(names, ", "))
}

func (c pluginLoadCommand) Run(ctx guest.Context) {
	target := commandTarget(c.Target)
	names, err := guest.LoadPlugins(target)
	if err != nil {
		ctx.Messagef("%s load failed: %s", pluginPrefix, err)
		return
	}
	if len(names) == 0 {
		ctx.Messagef("%s no plugins loaded", pluginPrefix)
		return
	}
	ctx.Messagef("%s loaded %d plugin(s): %s", pluginPrefix, len(names), strings.Join(names, ", "))
}

func (c pluginUnloadCommand) Run(ctx guest.Context) {
	target := commandTarget(c.Target)
	names, err := guest.UnloadPlugins(target)
	if err != nil {
		ctx.Messagef("%s unload failed: %s", pluginPrefix, err)
		return
	}
	if len(names) == 0 {
		ctx.Messagef("%s no plugins unloaded", pluginPrefix)
		return
	}
	ctx.Messagef("%s unloaded %d plugin(s): %s", pluginPrefix, len(names), strings.Join(names, ", "))
}

func (c pluginReloadCommand) Run(ctx guest.Context) {
	target := commandTarget(c.Target)
	names, err := guest.ReloadPlugins(target)
	if err != nil {
		ctx.Messagef("%s reload failed: %s", pluginPrefix, err)
		return
	}
	if len(names) == 0 {
		ctx.Messagef("%s no plugins reloaded", pluginPrefix)
		return
	}
	ctx.Messagef("%s reloaded %d plugin(s): %s", pluginPrefix, len(names), strings.Join(names, ", "))
}

func commandTarget(targetArg guest.Optional[string]) string {
	target := normalizeToken(targetArg.LoadOr(""))
	if target == "" {
		return allTarget
	}
	return target
}

func normalizeToken(v string) string {
	return strings.TrimSpace(strings.ToLower(v))
}
