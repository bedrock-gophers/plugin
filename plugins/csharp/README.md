# C# Ping Plugin Example

This example shows a C# plugin module that:
- handles `/ping` and replies `pong`
- listens to move events and sends a player message when yaw changes

Files:
- `PingPlugin.csproj`
- `PingPlugin.cs`

`PingPlugin` uses `BedrockPlugin.Sdk.Guest.Plugin` clean API (same surface can also be used directly as `var plugin = new Plugin(name, host)`):
- `RegisterCommand(...)`
- convention `OnMove(...)` (no explicit `HandleMove(...)` needed)
- `ctx.Message(...)`
- `ctx.TryPlayer(out var player)` and `player.Message(...)`

Host integration:
- build as NativeAOT shared library (`.so`)
- call `NativePlugin.Register<PingPlugin>()` (keeps convention handlers like `OnMove` rooted under NativeAOT trimming)
- export tiny forwarding entrypoints in plugin assembly that call:
  - `NativePlugin.Load(...)`
  - `NativePlugin.Unload()`
  - `NativePlugin.DispatchEvent(...)`

`IGuestHost` includes the already-implemented player-ref operations (health, gamemode, movement toggles, messaging, etc.) so plugin code can use `PlayerRef` directly without writing wrappers.
