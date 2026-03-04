# bedrock-gophers/plugin

Plugin host + SDK for Dragonfly, with support for:
- C# plugins (NativeAOT `.so`)

## Quick Start

Run the server + plugin build pipeline in Docker:

```bash
./start.sh
```

Run with PowerShell on Windows:

```powershell
./start.ps1
```

## Project Layout

- `cmd/` - server entrypoint
- `plugin/` - C# SDK
- `plugins/` - C# example plugins
- `internal/generator/output/` - generated Go/C/C# interop bindings (Go output kept as generated SDK example)

## Development

Show available make targets:

```bash
make help
```

Typical local checks:

```bash
make test
```

Regenerate interop bindings:

```bash
make generate
```

## Notes

- C# plugins are loaded from native shared libraries.
- For plugin reload while running, use:

```text
/pl reload all
```
