package commands

import "github.com/bedrock-gophers/plugin/plugin/sdk/go"

func Register() {
	guest.Base.RegisterCommand("gamemode", "change your own or other people's gamemode", []string{"gm"}, Gamemode{})
}
