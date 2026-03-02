# Bedrock Plugin C# ABI SDK

This directory provides a C# ABI implementation that mirrors the Go wire ABI in `plugin/abi` and event payload shapes used by the Go SDK.

Included:
- Binary encoder/decoder (`src/Abi/BinaryCodec.cs`)
- Event descriptor codec and protocol constants (`src/Abi/EventDescriptor.cs`, `src/Abi/EventIds.cs`, `src/Abi/PluginManageAction.cs`)
- Guest-side payload models and full event payload decoding (`src/Guest/*`)
- High-level guest runtime API:
  - `var plugin = new Plugin("<name>", host); plugin.RegisterCommand(...)`
  - optional command allowers via `RegisterCommand(..., allow, run)` overloads
  - convention handlers for all events (e.g. `OnMove`, `OnChat`, `OnBlockBreak`, ...)
  - `Plugin.HandleEvent(...)` / `Plugin.HandleMove(...)` for explicit registration
  - `CommandContext`/`EventContext` messaging (`Message`, `Messagef`)
  - `PlayerRef` operations backed by host API
- plugin management helpers (`ListPlugins`, `LoadPlugins`, `UnloadPlugins`, `ReloadPlugins`)
- native in-process plugin ABI bridge (`NativePlugin`):
  - `NativePlugin.Register<YourPlugin>()` (recommended for NativeAOT + convention handlers)
  - `NativePlugin.Load(...)`, `NativePlugin.Unload()`, `NativePlugin.DispatchEvent(...)`
  - plugin assembly exports can forward to these methods

The payload decoder supports all event IDs currently defined in `plugin/abi/events.go`, including `EventPluginCommand`.
