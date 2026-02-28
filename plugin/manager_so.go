package plugin

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	goplugin "plugin"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/bedrock-gophers/plugin/plugin/abi"
	"github.com/bedrock-gophers/plugin/plugin/internal/ctxkey"
	guest "github.com/bedrock-gophers/plugin/plugin/sdk/go"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

type Manager struct {
	ctx    context.Context
	cancel context.CancelFunc

	srv *server.Server
	dir string

	handler *Handler

	pluginsMu     sync.RWMutex
	plugins       []*pluginRuntime
	pluginsByName map[string]*pluginRuntime
	pluginsByPath map[string]*pluginRuntime

	pluginCommandsMu       sync.Mutex
	pluginCommandsByPlugin map[*pluginRuntime]map[string]*pluginCommandRegistration
	pluginCommandsByAlias  map[string]*pluginCommandRegistration
	releasedCommandAliases map[string]struct{}

	registrationMu    sync.RWMutex
	registeringPlugin *pluginRuntime

	players *playerStore

	closeOnce sync.Once
}

func Load(ctx context.Context, srv *server.Server, dirPath string) (*Manager, error) {
	if srv == nil {
		return nil, fmt.Errorf("load plugins: server is nil")
	}
	ctx, cancel := context.WithCancel(ctx)
	m := &Manager{
		ctx:                    ctx,
		cancel:                 cancel,
		srv:                    srv,
		dir:                    dirPath,
		pluginsByName:          map[string]*pluginRuntime{},
		pluginsByPath:          map[string]*pluginRuntime{},
		pluginCommandsByPlugin: map[*pluginRuntime]map[string]*pluginCommandRegistration{},
		pluginCommandsByAlias:  map[string]*pluginCommandRegistration{},
		releasedCommandAliases: map[string]struct{}{},
		players:                newPlayerStore(),
	}
	m.handler = &Handler{m: m}
	guest.SetHost(m)

	if err := m.loadPlugins(dirPath); err != nil {
		guest.SetHost(nil)
		cancel()
		return nil, err
	}
	return m, nil
}

func (m *Manager) Close(context.Context) error {
	m.closeOnce.Do(func() {
		m.cancel()
		plugins := m.loadedPluginsSnapshot()
		for i := len(plugins) - 1; i >= 0; i-- {
			m.deactivatePlugin(plugins[i])
		}
		guest.SetHost(nil)
	})
	return nil
}

func (m *Manager) Handler() player.Handler {
	return m.handler
}

func (m *Manager) List() []string {
	plugins := m.loadedPluginsSnapshot()
	names := make([]string, 0, len(plugins))
	for _, plug := range plugins {
		names = append(names, plug.name)
	}
	sort.Strings(names)
	return names
}

func (m *Manager) Load(target string) ([]string, error) {
	target = normalizePluginTarget(target)
	if target == "" {
		return nil, fmt.Errorf("load target cannot be empty")
	}
	if target == "all" {
		return m.loadAllMissing()
	}
	return m.loadOne(target)
}

func (m *Manager) Unload(target string) ([]string, error) {
	target = normalizePluginTarget(target)
	if target == "" {
		return nil, fmt.Errorf("unload target cannot be empty")
	}
	if target == "all" {
		return m.unloadAll()
	}
	return m.unloadOne(target)
}

func (m *Manager) Reload(target string) ([]string, error) {
	target = normalizePluginTarget(target)
	if target == "" {
		return nil, fmt.Errorf("reload target cannot be empty")
	}
	if target == "all" {
		return m.reloadAll()
	}
	return m.reloadOne(target)
}

func (m *Manager) Attach(p *player.Player) {
	if p == nil {
		return
	}
	m.players.ensure(p)
	p.Handle(m.handler)
}

func (m *Manager) AllowJoin(_ net.Addr, d login.IdentityData, _ login.ClientData) (string, bool) {
	allowed := true
	denyMessage := "join denied by plugin"
	mutable := newMutableState(func() {
		allowed = false
	})
	mutable.AddString(ctxkey.JoinCancelMessage, denyMessage, func(v string) {
		denyMessage = v
	})
	m.dispatchByPlayerID(0, abi.EventJoin, abi.FlagCancellable, payloadPlayerIdentityValues(d.DisplayName, d.Identity, d.XUID), mutable)
	if !allowed {
		return denyMessage, false
	}
	return "", true
}

func (m *Manager) loadPlugins(dirPath string) error {
	paths, err := collectPluginPaths(dirPath)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		slog.Info("no plugins found", "dir", dirPath)
		return nil
	}
	slog.Info("discovered plugins", "dir", dirPath, "count", len(paths))

	for _, path := range paths {
		slog.Info("loading plugin", "path", path)
		if _, err := m.loadPath(path, ""); err != nil {
			return err
		}
	}
	return nil
}

func collectGoSOPaths(dirPath string) ([]string, error) {
	if _, err := os.Stat(dirPath); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat plugins dir: %w", err)
	}

	csDir := filepath.Join(dirPath, "csharp")

	var soPaths []string
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if strings.EqualFold(path, csDir) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(d.Name()), ".so") {
			return nil
		}
		normalized, err := normalizePluginPath(path)
		if err != nil {
			return err
		}
		soPaths = append(soPaths, normalized)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk plugins: %w", err)
	}
	sort.Strings(soPaths)
	return soPaths, nil
}

func collectPluginPaths(dirPath string) ([]string, error) {
	soPaths, err := collectGoSOPaths(dirPath)
	if err != nil {
		return nil, err
	}
	csPaths, err := collectCSharpPluginSOs(dirPath)
	if err != nil {
		return nil, err
	}
	all := append(soPaths, csPaths...)
	sort.Strings(all)
	return all, nil
}

func collectCSharpPluginSOs(dirPath string) ([]string, error) {
	csDir := filepath.Join(dirPath, "csharp")
	if _, err := os.Stat(csDir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat csharp plugins dir: %w", err)
	}

	var soPaths []string
	err := filepath.WalkDir(csDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if strings.EqualFold(d.Name(), "runtime") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(d.Name()), ".so") {
			return nil
		}
		base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if !strings.EqualFold(base, filepath.Base(filepath.Dir(path))) {
			return nil
		}
		normalized, err := normalizePluginPath(path)
		if err != nil {
			return err
		}
		soPaths = append(soPaths, normalized)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk csharp plugins: %w", err)
	}
	sort.Strings(soPaths)
	return soPaths, nil
}

func (m *Manager) loadAllMissing() ([]string, error) {
	paths, err := collectPluginPaths(m.dir)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, nil
	}

	loaded := make([]string, 0, len(paths))
	for _, path := range paths {
		m.pluginsMu.RLock()
		existing := m.pluginsByPath[path]
		m.pluginsMu.RUnlock()

		if existing != nil {
			if existing.loaded {
				continue
			}
			if err := m.activatePlugin(existing); err != nil {
				return nil, err
			}
			loaded = append(loaded, existing.name)
			continue
		}

		name, err := m.loadPath(path, "")
		if err != nil {
			return nil, err
		}
		loaded = append(loaded, name)
	}
	sort.Strings(loaded)
	return loaded, nil
}

func (m *Manager) loadOne(target string) ([]string, error) {
	m.pluginsMu.RLock()
	if plug, ok := m.resolveKnownPluginLocked(target); ok {
		m.pluginsMu.RUnlock()
		if plug.loaded {
			return nil, fmt.Errorf("plugin %q is already loaded", plug.name)
		}
		if err := m.activatePlugin(plug); err != nil {
			return nil, err
		}
		return []string{plug.name}, nil
	}
	m.pluginsMu.RUnlock()

	path, err := m.resolvePluginPath(target)
	if err != nil {
		return nil, err
	}

	m.pluginsMu.RLock()
	if existing := m.pluginsByPath[path]; existing != nil {
		m.pluginsMu.RUnlock()
		if existing.loaded {
			return nil, fmt.Errorf("plugin %q is already loaded", existing.name)
		}
		if err := m.activatePlugin(existing); err != nil {
			return nil, err
		}
		return []string{existing.name}, nil
	}
	m.pluginsMu.RUnlock()

	name, err := m.loadPath(path, "")
	if err != nil {
		return nil, err
	}
	return []string{name}, nil
}

func (m *Manager) unloadAll() ([]string, error) {
	plugins := m.loadedPluginsSnapshot()
	names := make([]string, 0, len(plugins))
	for _, plug := range plugins {
		names = append(names, plug.name)
	}
	for i := len(plugins) - 1; i >= 0; i-- {
		m.deactivatePlugin(plugins[i])
	}
	sort.Strings(names)
	return names, nil
}

func (m *Manager) unloadOne(target string) ([]string, error) {
	m.pluginsMu.RLock()
	plug, ok := m.resolveLoadedPluginLocked(target)
	m.pluginsMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("plugin %q not loaded", target)
	}
	m.deactivatePlugin(plug)
	return []string{plug.name}, nil
}

func (m *Manager) reloadAll() ([]string, error) {
	if _, err := m.unloadAll(); err != nil {
		return nil, err
	}
	if _, err := m.loadAllMissing(); err != nil {
		return nil, err
	}
	return m.List(), nil
}

func (m *Manager) reloadOne(target string) ([]string, error) {
	m.pluginsMu.RLock()
	plug, ok := m.resolveLoadedPluginLocked(target)
	m.pluginsMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("plugin %q not loaded", target)
	}
	m.deactivatePlugin(plug)
	if err := m.activatePlugin(plug); err != nil {
		return nil, err
	}
	return []string{plug.name}, nil
}

func (m *Manager) resolveKnownPluginLocked(target string) (*pluginRuntime, bool) {
	if target == "" {
		return nil, false
	}
	if plug, ok := m.pluginsByName[target]; ok {
		return plug, true
	}

	targetWithExt := target
	targetNoExt := strings.TrimSuffix(target, filepath.Ext(target))
	targetPath := strings.ToLower(filepath.Clean(target))
	targetPathFromDir := targetPath
	if !filepath.IsAbs(target) {
		targetPathFromDir = strings.ToLower(filepath.Clean(filepath.Join(m.dir, target)))
	}

	for _, plug := range m.pluginsByPath {
		candidatePath := strings.ToLower(filepath.Clean(plug.path))
		if candidatePath == targetPath || candidatePath == targetPathFromDir {
			return plug, true
		}
		baseWithExt := strings.ToLower(filepath.Base(plug.path))
		base := strings.TrimSuffix(baseWithExt, filepath.Ext(baseWithExt))
		if baseWithExt == targetWithExt || base == targetNoExt || sanitizeModuleName(base) == targetNoExt {
			return plug, true
		}
	}
	return nil, false
}

func (m *Manager) resolveLoadedPluginLocked(target string) (*pluginRuntime, bool) {
	plug, ok := m.resolveKnownPluginLocked(target)
	if !ok || !plug.loaded {
		return nil, false
	}
	return plug, true
}

func (m *Manager) resolvePluginPath(target string) (string, error) {
	target = normalizePluginTarget(target)
	if target == "" {
		return "", fmt.Errorf("plugin target cannot be empty")
	}

	paths, err := collectPluginPaths(m.dir)
	if err != nil {
		return "", err
	}
	if len(paths) == 0 {
		return "", fmt.Errorf("no plugin files found in %q", m.dir)
	}

	targetWithExt := target
	targetNoExt := strings.TrimSuffix(target, filepath.Ext(target))
	targetPath := strings.ToLower(filepath.Clean(target))
	targetPathFromDir := targetPath
	if !filepath.IsAbs(target) {
		targetPathFromDir = strings.ToLower(filepath.Clean(filepath.Join(m.dir, target)))
	}

	exactMatches := make([]string, 0, 2)
	nameMatches := make([]string, 0, 4)
	for _, path := range paths {
		pathNormalized := strings.ToLower(filepath.Clean(path))
		if pathNormalized == targetPath || pathNormalized == targetPathFromDir {
			exactMatches = append(exactMatches, path)
			continue
		}
		baseWithExt := strings.ToLower(filepath.Base(path))
		base := strings.TrimSuffix(baseWithExt, filepath.Ext(baseWithExt))
		if baseWithExt == targetWithExt || base == targetNoExt || sanitizeModuleName(base) == targetNoExt {
			nameMatches = append(nameMatches, path)
		}
	}

	matches := nameMatches
	if len(exactMatches) > 0 {
		matches = exactMatches
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("plugin %q not found in %q", target, m.dir)
	}
	if len(matches) > 1 {
		labels := make([]string, 0, len(matches))
		for _, path := range matches {
			labels = append(labels, filepath.Base(path))
		}
		sort.Strings(labels)
		return "", fmt.Errorf("plugin %q is ambiguous; matches: %s", target, strings.Join(labels, ", "))
	}
	return matches[0], nil
}

func (m *Manager) loadPath(path, forcedName string) (string, error) {
	normalizedPath, err := normalizePluginPath(path)
	if err != nil {
		return "", err
	}

	m.pluginsMu.RLock()
	if existing := m.pluginsByPath[normalizedPath]; existing != nil {
		m.pluginsMu.RUnlock()
		if existing.loaded {
			return "", fmt.Errorf("plugin %q is already loaded", existing.name)
		}
		if err := m.activatePlugin(existing); err != nil {
			return "", err
		}
		return existing.name, nil
	}
	m.pluginsMu.RUnlock()

	base := strings.TrimSuffix(filepath.Base(normalizedPath), filepath.Ext(normalizedPath))
	name := forcedName
	if name == "" {
		name = m.nextAvailablePluginName(base)
	}
	var kind pluginKind
	switch strings.ToLower(filepath.Ext(normalizedPath)) {
	case ".so":
		kind = m.pluginKindForSOPath(normalizedPath)
	default:
		return "", fmt.Errorf("unsupported plugin extension %q", filepath.Ext(normalizedPath))
	}
	plug := newPluginRuntime(name, normalizedPath, kind)

	var onLoad, onUnload func()
	switch plug.kind {
	case pluginKindGo:
		m.setRegisteringPlugin(plug)
		guest.BeginPluginRegistration(plug.name)
		opened, openErr := goplugin.Open(normalizedPath)
		if openErr == nil {
			onLoad, onUnload, openErr = resolvePluginHooks(opened)
		}
		if openErr == nil {
			openErr = callPluginLoad(onLoad)
		}
		guest.EndPluginRegistration()
		m.clearRegisteringPlugin(plug)
		err = openErr
	case pluginKindCSharp:
		m.setRegisteringPlugin(plug)
		err = m.startCSharpPlugin(plug)
		m.clearRegisteringPlugin(plug)
	default:
		err = fmt.Errorf("unsupported plugin kind %d", plug.kind)
	}
	if err != nil {
		m.unregisterPluginCommandsFor(plug)
		return "", fmt.Errorf("load plugin %s: %w", normalizedPath, err)
	}
	if plug.kind == pluginKindGo {
		plug.onUnload = onUnload
	}

	m.pluginsMu.Lock()
	plug.loaded = true
	m.plugins = append(m.plugins, plug)
	m.pluginsByName[plug.name] = plug
	m.pluginsByPath[plug.path] = plug
	m.pluginsMu.Unlock()

	slog.Info("plugin loaded", "name", plug.name, "path", plug.path)
	return plug.name, nil
}

func resolvePluginHooks(p *goplugin.Plugin) (onLoad, onUnload func(), err error) {
	if p == nil {
		return nil, nil, nil
	}
	onLoad, _, err = lookupOptionalPluginFunc(p, "PluginLoad", "PluginInit", "DFPluginInit")
	if err != nil {
		return nil, nil, err
	}
	onUnload, _, err = lookupOptionalPluginFunc(p, "PluginUnload")
	if err != nil {
		return nil, nil, err
	}
	return onLoad, onUnload, nil
}

func lookupOptionalPluginFunc(p *goplugin.Plugin, symbolNames ...string) (fn func(), found bool, err error) {
	if p == nil || len(symbolNames) == 0 {
		return nil, false, nil
	}
	for _, symbolName := range symbolNames {
		symbol, lookupErr := p.Lookup(symbolName)
		if lookupErr != nil {
			continue
		}
		switch typed := symbol.(type) {
		case func():
			return typed, true, nil
		case *func():
			return *typed, true, nil
		default:
			return nil, true, fmt.Errorf("symbol %q has type %T, expected func()", symbolName, symbol)
		}
	}
	return nil, false, nil
}

func callPluginLoad(fn func()) (err error) {
	if fn == nil {
		return nil
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("plugin load panic: %v", r)
		}
	}()
	fn()
	return nil
}

func callPluginUnload(plug *pluginRuntime) {
	if plug == nil || plug.onUnload == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			slog.Error("plugin unload panic", "name", plug.name, "path", plug.path, "panic", r)
		}
	}()
	plug.onUnload()
}

func (m *Manager) activatePlugin(plug *pluginRuntime) error {
	if plug == nil {
		return fmt.Errorf("plugin is nil")
	}

	m.pluginsMu.Lock()
	if plug.loaded {
		m.pluginsMu.Unlock()
		return fmt.Errorf("plugin %q is already loaded", plug.name)
	}
	if _, ok := m.pluginsByName[plug.name]; ok {
		m.pluginsMu.Unlock()
		return fmt.Errorf("plugin name %q is already in use", plug.name)
	}
	plug.loaded = true
	m.plugins = append(m.plugins, plug)
	m.pluginsByName[plug.name] = plug
	m.pluginsByPath[plug.path] = plug
	m.pluginsMu.Unlock()

	var err error
	switch plug.kind {
	case pluginKindGo:
		err = m.activatePluginCommands(plug)
	case pluginKindCSharp:
		m.setRegisteringPlugin(plug)
		err = m.startCSharpPlugin(plug)
		m.clearRegisteringPlugin(plug)
	default:
		err = fmt.Errorf("unsupported plugin kind %d", plug.kind)
	}
	if err != nil {
		m.pluginsMu.Lock()
		delete(m.pluginsByName, plug.name)
		for i := range m.plugins {
			if m.plugins[i] == plug {
				m.plugins = append(m.plugins[:i], m.plugins[i+1:]...)
				break
			}
		}
		plug.loaded = false
		m.pluginsMu.Unlock()
		m.unregisterPluginCommandsFor(plug)
		return err
	}
	return nil
}

func (m *Manager) deactivatePlugin(plug *pluginRuntime) {
	if plug == nil {
		return
	}
	m.pluginsMu.Lock()
	if !plug.loaded {
		m.pluginsMu.Unlock()
		return
	}
	plug.loaded = false
	delete(m.pluginsByName, plug.name)
	for i := range m.plugins {
		if m.plugins[i] == plug {
			m.plugins = append(m.plugins[:i], m.plugins[i+1:]...)
			break
		}
	}
	m.pluginsMu.Unlock()
	m.unregisterPluginCommandsFor(plug)
	callPluginUnload(plug)
	slog.Info("plugin unloaded", "name", plug.name, "path", plug.path)
}

func (m *Manager) activatePluginCommands(plug *pluginRuntime) error {
	blueprints := plug.commandBlueprintsSnapshot()
	sort.Slice(blueprints, func(i, j int) bool {
		return blueprints[i].name < blueprints[j].name
	})
	for _, blueprint := range blueprints {
		if err := m.registerPluginCommandBlueprint(plug, blueprint, false); err != nil {
			return fmt.Errorf("activate command %q for plugin %q: %w", blueprint.name, plug.name, err)
		}
	}
	return nil
}

func normalizePluginTarget(v string) string {
	return strings.TrimSpace(strings.ToLower(v))
}

func normalizePluginPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("plugin path cannot be empty")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve plugin path %q: %w", path, err)
	}
	return filepath.Clean(abs), nil
}

func (m *Manager) pluginKindForSOPath(path string) pluginKind {
	csRoot, err := normalizePluginPath(filepath.Join(m.dir, "csharp"))
	if err == nil && pathWithin(path, csRoot) {
		return pluginKindCSharp
	}
	return pluginKindGo
}

func pathWithin(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func (m *Manager) loadedPluginsSnapshot() []*pluginRuntime {
	m.pluginsMu.RLock()
	plugins := append([]*pluginRuntime(nil), m.plugins...)
	m.pluginsMu.RUnlock()
	return plugins
}

func (m *Manager) nextAvailablePluginName(base string) string {
	name := sanitizeModuleName(base)
	used := map[string]struct{}{}
	m.pluginsMu.RLock()
	for _, plug := range m.pluginsByPath {
		used[plug.name] = struct{}{}
	}
	m.pluginsMu.RUnlock()
	if _, exists := used[name]; !exists {
		return name
	}
	for i := 1; ; i++ {
		candidate := name + "_" + strconv.Itoa(i)
		if _, exists := used[candidate]; !exists {
			return candidate
		}
	}
}

func sanitizeModuleName(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return "plugin"
	}
	var b strings.Builder
	for _, r := range v {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

func normalizePluginCommandToken(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" || strings.ContainsAny(v, " \t\r\n") {
		return ""
	}
	return v
}

func (m *Manager) registerPluginCommand(plug *pluginRuntime, name, description string, aliases []string, handlerID uint32, overloads []pluginCommandOverloadSpec) error {
	blueprint := pluginCommandBlueprint{
		name:        name,
		description: description,
		aliases:     append([]string(nil), aliases...),
		handlerID:   handlerID,
		overloads:   append([]pluginCommandOverloadSpec(nil), overloads...),
	}
	return m.registerPluginCommandBlueprint(plug, blueprint, true)
}

func (m *Manager) registerPluginCommandBlueprint(plug *pluginRuntime, blueprint pluginCommandBlueprint, persist bool) error {
	allAliases := make([]string, 0, len(blueprint.aliases)+1)
	allAliases = append(allAliases, blueprint.name)
	allAliases = append(allAliases, blueprint.aliases...)

	m.pluginCommandsMu.Lock()
	defer m.pluginCommandsMu.Unlock()

	byName := m.pluginCommandsByPlugin[plug]
	if byName == nil {
		byName = map[string]*pluginCommandRegistration{}
		m.pluginCommandsByPlugin[plug] = byName
	}
	if _, exists := byName[blueprint.name]; exists {
		return fmt.Errorf("command %q already registered by plugin", blueprint.name)
	}

	for _, alias := range allAliases {
		if existing, ok := m.pluginCommandsByAlias[alias]; ok {
			if existing.plugin == plug {
				return fmt.Errorf("command alias %q already used by this plugin", alias)
			}
			return fmt.Errorf("command alias %q already registered by plugin %q", alias, existing.plugin.name)
		}
		if existing, ok := cmd.ByAlias(alias); ok {
			if _, reusable := m.releasedCommandAliases[alias]; !reusable || !isPluginRuntimeCommand(existing) {
				return fmt.Errorf("command alias %q is already registered", alias)
			}
		}
	}

	entry := &pluginCommandRegistration{
		plugin:      plug,
		name:        blueprint.name,
		description: blueprint.description,
		aliases:     append([]string(nil), blueprint.aliases...),
		handlerID:   blueprint.handlerID,
		overloads:   append([]pluginCommandOverloadSpec(nil), blueprint.overloads...),
	}

	runnables := make([]cmd.Runnable, 0, max(1, len(blueprint.overloads)))
	if len(blueprint.overloads) == 0 {
		runnables = append(runnables, pluginCommandRunnable{manager: m, entry: entry})
	} else {
		for i, overload := range blueprint.overloads {
			runnables = append(runnables, pluginCommandRunnable{
				manager:       m,
				entry:         entry,
				overload:      overload,
				overloadIndex: i,
			})
		}
	}
	command := cmd.New(blueprint.name, blueprint.description, append([]string(nil), blueprint.aliases...), runnables...)
	cmd.Register(command)

	byName[blueprint.name] = entry
	for _, alias := range allAliases {
		m.pluginCommandsByAlias[alias] = entry
		delete(m.releasedCommandAliases, alias)
	}
	if persist {
		plug.setCommandBlueprint(blueprint)
	}
	return nil
}

func (m *Manager) unregisterPluginCommandsFor(plug *pluginRuntime) {
	m.pluginCommandsMu.Lock()
	byName, ok := m.pluginCommandsByPlugin[plug]
	if !ok {
		m.pluginCommandsMu.Unlock()
		return
	}

	for name, entry := range byName {
		delete(byName, name)
		delete(m.pluginCommandsByAlias, entry.name)
		m.releasedCommandAliases[entry.name] = struct{}{}
		for _, alias := range entry.aliases {
			delete(m.pluginCommandsByAlias, alias)
			m.releasedCommandAliases[alias] = struct{}{}
		}
	}
	delete(m.pluginCommandsByPlugin, plug)
	m.pluginCommandsMu.Unlock()
}

func (m *Manager) executePluginCommand(entry *pluginCommandRegistration, playerID uint64, args []string) error {
	if entry == nil || entry.plugin == nil {
		return fmt.Errorf("plugin command is unavailable")
	}
	m.pluginCommandsMu.Lock()
	byName, ok := m.pluginCommandsByPlugin[entry.plugin]
	if !ok {
		m.pluginCommandsMu.Unlock()
		return fmt.Errorf("plugin command is unavailable")
	}
	active, ok := byName[entry.name]
	m.pluginCommandsMu.Unlock()
	if !ok || active != entry {
		return fmt.Errorf("plugin command is unavailable")
	}

	enc := abi.NewEncoder(128)
	enc.U32(entry.handlerID)
	enc.U32(uint32(len(args)))
	for _, arg := range args {
		enc.String(arg)
	}
	m.dispatchToPlugin(entry.plugin, playerID, abi.EventPluginCommand, 0, enc.Data(), nil)
	return nil
}

func (m *Manager) dispatch(p *player.Player, eventID uint16, flags uint32, payload []byte, mutable *mutableState) {
	if p == nil {
		return
	}
	playerID := m.players.ensure(p)
	m.dispatchByPlayerID(playerID, eventID, flags, payload, mutable)
}

func (m *Manager) dispatchByPlayerID(playerID uint64, eventID uint16, flags uint32, payload []byte, mutable *mutableState) {
	plugins := m.loadedPluginsSnapshot()
	if len(plugins) == 0 {
		return
	}
	for _, plug := range plugins {
		m.dispatchToPlugin(plug, playerID, eventID, flags, payload, mutable)
		if mutable != nil {
			mutable.Apply()
			if mutable.Cancelled() {
				return
			}
		}
	}
}

func (m *Manager) dispatchToPlugin(plug *pluginRuntime, playerID uint64, eventID uint16, flags uint32, payload []byte, mutable *mutableState) {
	if plug == nil || !plug.loaded {
		return
	}
	desc := descriptorForDispatch(eventID, flags, playerID)
	switch plug.kind {
	case pluginKindGo:
		guest.DispatchEvent(plug.name, desc, payload, mutable)
	case pluginKindCSharp:
		if plug.csharp == nil {
			return
		}
		if err := plug.csharp.dispatch(m, plug, desc, payload); err != nil {
			slog.Error("dispatch csharp plugin event", "plugin", plug.name, "event", abi.EventName(eventID), "err", err)
		}
	}
}

func (m *Manager) resolveWorld(name string) *world.World {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return nil
	}
	ow := m.srv.World()
	if strings.ToLower(ow.Name()) == name || name == "overworld" {
		return ow
	}
	nw := m.srv.Nether()
	if strings.ToLower(nw.Name()) == name || name == "nether" {
		return nw
	}
	ew := m.srv.End()
	if strings.ToLower(ew.Name()) == name || name == "end" {
		return ew
	}
	return nil
}

func (m *Manager) setRegisteringPlugin(plug *pluginRuntime) {
	m.registrationMu.Lock()
	m.registeringPlugin = plug
	m.registrationMu.Unlock()
}

func (m *Manager) clearRegisteringPlugin(plug *pluginRuntime) {
	m.registrationMu.Lock()
	if m.registeringPlugin == plug {
		m.registeringPlugin = nil
	}
	m.registrationMu.Unlock()
}

func (m *Manager) registeringPluginForName(name string) *pluginRuntime {
	m.registrationMu.RLock()
	plug := m.registeringPlugin
	m.registrationMu.RUnlock()
	if plug != nil && (name == "" || plug.name == name) {
		return plug
	}
	return nil
}

func (m *Manager) pluginByNameAny(name string) *pluginRuntime {
	name = normalizePluginTarget(name)
	m.pluginsMu.RLock()
	defer m.pluginsMu.RUnlock()
	if plug, ok := m.pluginsByName[name]; ok {
		return plug
	}
	for _, plug := range m.pluginsByPath {
		if plug.name == name {
			return plug
		}
	}
	return nil
}

func (m *Manager) RegisterCommand(pluginName, name, description string, aliases []string, handlerID uint32, overloads []guest.CommandOverloadSpec) bool {
	plug := m.registeringPluginForName(pluginName)
	if plug == nil {
		plug = m.pluginByNameAny(pluginName)
	}
	if plug == nil {
		slog.Error("register plugin command", "plugin", pluginName, "name", name, "err", "plugin not found")
		return false
	}

	normalizedName := normalizePluginCommandToken(name)
	description = strings.TrimSpace(description)
	normalizedAliases := normalizeCommandAliases(aliases, normalizedName)
	if normalizedName == "" || handlerID == 0 {
		return false
	}

	if err := m.registerPluginCommand(plug, normalizedName, description, normalizedAliases, handlerID, convertGuestOverloads(overloads)); err != nil {
		slog.Error("register plugin command", "plugin", plug.name, "name", normalizedName, "err", err)
		return false
	}
	return true
}

func convertGuestOverloads(overloads []guest.CommandOverloadSpec) []pluginCommandOverloadSpec {
	if len(overloads) == 0 {
		return nil
	}
	out := make([]pluginCommandOverloadSpec, 0, len(overloads))
	for _, overload := range overloads {
		params := make([]pluginCommandParamSpec, 0, len(overload.Parameters))
		for _, parameter := range overload.Parameters {
			name := strings.TrimSpace(parameter.Name)
			if name == "" {
				continue
			}
			options := make([]string, 0, len(parameter.EnumOptions))
			for _, option := range parameter.EnumOptions {
				token := normalizePluginCommandToken(option)
				if token == "" {
					continue
				}
				options = append(options, token)
			}
			params = append(params, pluginCommandParamSpec{
				name:        name,
				kind:        pluginCommandParamKind(parameter.Kind),
				optional:    parameter.Optional,
				enumOptions: options,
			})
		}
		out = append(out, pluginCommandOverloadSpec{parameters: params})
	}
	return out
}

func normalizeCommandAliases(aliases []string, commandName string) []string {
	out := make([]string, 0, len(aliases))
	seen := map[string]struct{}{commandName: {}}
	for _, alias := range aliases {
		token := normalizePluginCommandToken(alias)
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

func (m *Manager) ManagePlugins(action uint32, target string) ([]string, error) {
	target = strings.TrimSpace(target)
	switch action {
	case abi.PluginManageList:
		return m.List(), nil
	case abi.PluginManageLoad:
		return m.Load(target)
	case abi.PluginManageUnload:
		return m.Unload(target)
	case abi.PluginManageReload:
		return m.Reload(target)
	default:
		return nil, fmt.Errorf("unsupported plugin manage action %d", action)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
