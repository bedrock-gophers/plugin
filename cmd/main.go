package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/bedrock-gophers/plugin/plugin"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player/chat"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelInfo)
	chat.Global.Subscribe(chat.StdoutSubscriber{})
	conf, err := config(slog.Default())
	if err != nil {
		panic(err)
	}

	allower := plugin.NewServerAllower(conf.Allower)
	conf.Allower = allower

	srv := conf.New()

	pluginDir := strings.TrimSpace(os.Getenv("PLUGIN_DIR"))
	if pluginDir == "" {
		pluginDir = "plugins"
	}

	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		panic(fmt.Errorf("create plugins dir: %w", err))
	}
	mgr, err := plugin.Load(context.Background(), srv, pluginDir)
	if err != nil {
		panic(err)
	}
	allower.SetManager(mgr)
	shutdown := newShutdownController(srv, mgr)
	defer shutdown.Shutdown("main exit")

	watchForShutdownSignals(shutdown)
	go runConsoleLoop(shutdown, srv)

	srv.Listen()
	for p := range srv.Accept() {
		mgr.Attach(p)
	}
	shutdown.Shutdown("server accept loop ended")
}

func config(log *slog.Logger) (server.Config, error) {
	c := server.DefaultConfig()
	c.Network.Address = ":19132"

	c.Players.Folder = ".data/players"
	c.Players.MaxCount = 0
	c.Players.MaximumChunkRadius = 32
	c.Players.SaveData = true

	c.Resources.AutoBuildPack = true
	c.Resources.Folder = ".data/resources"
	c.Resources.Required = false

	c.Server.AuthEnabled = true
	c.Server.DisableJoinQuitMessages = false
	c.Server.MuteEmoteChat = false
	c.Server.Name = "Dragonfly Server"

	c.World.Folder = ".data/world"
	c.World.SaveData = true

	return c.Config(log)
}
