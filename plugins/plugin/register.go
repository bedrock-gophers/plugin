package main

import guest "github.com/bedrock-gophers/plugin/plugin/sdk/go"

func PluginLoad() {
	registerChatHandlers()

	guest.Base.RegisterCommand("plugin", "Manage .so plugins.", []string{"pl"},
		pluginListCommand{},
		pluginLoadCommand{},
		pluginUnloadCommand{},
		pluginReloadCommand{},
	)
}

func PluginUnload() {}
