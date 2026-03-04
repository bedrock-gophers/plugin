# User Constraints (Current Session)

- Do not bring back or rely on old runtime code paths.
- Do not use or reintroduce the C# runtime bridge files the user rejected.
- Keep the solution generic and as small as possible (least code).
- Do not inspect or base changes on old code.
- Keep item API under the `Item` namespace in C# usage.
- Prefer typed wrappers/arguments over payload objects where applicable.

## Explicitly rejected files/patterns

- `plugin/native_runtime_bridge_generated.go`
- `plugin/native_runtime_cgo.go`
- `plugin/native_runtime_nocgo.go`
- `plugin/native_runtime_public.go`
- Large generated C# runtime bridge style glue in Go.

## Working style reminder

- Make minimal targeted changes.
- If a choice is ambiguous, choose the most generic implementation with the least code.
