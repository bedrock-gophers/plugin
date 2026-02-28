package main

import "github.com/bedrock-gophers/plugin/plugins/vanilla/commands"

func PluginLoad() {
	commands.Register()
}

func PluginUnload() {}
