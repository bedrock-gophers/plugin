package main

import (
	"strings"

	guest "github.com/bedrock-gophers/plugin/plugin/sdk/go"
)

func registerChatHandlers() {
	guest.Base.HandleEvent(handleChatMessage)
}

func handleChatMessage(ev *guest.Event, message guest.MutableArgument[string]) {
	text := strings.TrimSpace(message.Get())
	if text == "" {
		return
	}

	switch {
	case strings.HasPrefix(text, "!upper "):
		upper := strings.ToUpper(strings.TrimSpace(strings.TrimPrefix(text, "!upper ")))
		if upper == "" {
			ev.Player.Message(pluginPrefix + " usage: !upper <text>")
			return
		}
		message.Set(upper)
	case strings.EqualFold(text, "!pluginping"):
		ev.Player.Message(pluginPrefix + " pong from chat handler")
	}
}
