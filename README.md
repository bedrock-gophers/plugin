# bedrock-gophers/plugin

Plugin host + SDK for Dragonfly, with support for:
- Go plugins (`.so`)
- C# plugins (NativeAOT `.so`)
- Rust plugins (`cdylib` `.so`)

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
- `plugin/` - host runtime + ABI + SDKs
- `plugins/` - example plugins (`vanilla` in Rust, `plugin` in Go, `csharp`)

## Development

Show available make targets:

```bash
make help
```

Typical local checks:

```bash
make test
```

Regenerate internal generated code:

```bash
make generate
```

## Notes

- C# plugins require `cgo` in the host build and are loaded from native shared libraries.
- For plugin reload while running, use:

```text
/pl reload all
```
