package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/bedrock-gophers/plugin/plugin"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type shutdownController struct {
	once sync.Once

	srv *server.Server
	mgr *plugin.Manager
}

func newShutdownController(srv *server.Server, mgr *plugin.Manager) *shutdownController {
	return &shutdownController{srv: srv, mgr: mgr}
}

func (c *shutdownController) Shutdown(reason string) {
	if c == nil {
		return
	}
	c.once.Do(func() {
		reason = strings.TrimSpace(reason)
		if reason == "" {
			reason = "unspecified"
		}
		slog.Info("shutdown requested", "reason", reason)
		if c.mgr != nil {
			if err := c.mgr.Close(context.Background()); err != nil {
				slog.Error("close plugin manager", "err", err)
			}
		}
		if c.srv != nil {
			if err := c.srv.Close(); err != nil {
				slog.Error("close server", "err", err)
			}
		}
	})
}

func watchForShutdownSignals(shutdown *shutdownController) {
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signals
		signal.Stop(signals)
		shutdown.Shutdown("signal " + sig.String())
	}()
}

func runConsoleLoop(shutdown *shutdownController, srv *server.Server) {
	scanner := bufio.NewScanner(os.Stdin)
	source := consoleCommandSource{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if handleConsoleCommand(line, shutdown, srv, source) {
			continue
		}
		slog.Warn("unknown console command", "line", line)
	}
	if err := scanner.Err(); err != nil {
		slog.Error("read console input", "err", err)
	}
}

func handleConsoleCommand(line string, shutdown *shutdownController, srv *server.Server, source consoleCommandSource) bool {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "/")
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return true
	}
	switch strings.ToLower(fields[0]) {
	case "help", "?":
		printConsoleHelp()
		return true
	case "stop", "shutdown", "exit", "quit":
		shutdown.Shutdown("console " + strings.ToLower(fields[0]) + " command")
		return true
	default:
		return executeRegisteredCommand(srv, source, line)
	}
}

func printConsoleHelp() {
	slog.Info("console commands", "usage", "help")
	names := availableCommandNames()
	if len(names) > 0 {
		slog.Info("registered commands", "count", len(names), "commands", strings.Join(names, ", "))
	}
	slog.Info("console commands", "usage", "stop")
}

func availableCommandNames() []string {
	byAlias := cmd.Commands()
	set := map[string]struct{}{}
	for _, command := range byAlias {
		set[command.Name()] = struct{}{}
	}
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func executeRegisteredCommand(srv *server.Server, source consoleCommandSource, line string) bool {
	tokens := strings.Split(line, " ")
	if len(tokens) == 0 {
		return true
	}
	name := strings.ToLower(strings.TrimSpace(tokens[0]))
	if name == "" {
		return true
	}
	command, ok := cmd.ByAlias(name)
	if !ok {
		return false
	}

	args := strings.TrimSpace(strings.Join(tokens[1:], " "))
	<-srv.World().Exec(func(tx *world.Tx) {
		command.Execute(args, source, tx)
	})
	return true
}

type consoleCommandSource struct{}

func (consoleCommandSource) Position() mgl64.Vec3 {
	return mgl64.Vec3{}
}

func (consoleCommandSource) SendCommandOutput(o *cmd.Output) {
	if o == nil {
		return
	}
	for _, msg := range o.Messages() {
		if msg == nil {
			continue
		}
		slog.Info("command output", "message", msg.String())
	}
	for _, err := range o.Errors() {
		if err == nil {
			continue
		}
		slog.Error("command output", "err", err.Error())
	}
}

func (consoleCommandSource) Name() string {
	return "Console"
}

func (consoleCommandSource) String() string {
	return "Console"
}

func (consoleCommandSource) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v', 's':
		_, _ = s.Write([]byte("Console"))
	case 'q':
		_, _ = s.Write([]byte(strconv.Quote("Console")))
	default:
		_, _ = s.Write([]byte("Console"))
	}
}
