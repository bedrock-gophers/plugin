#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGINS_DIR="$ROOT_DIR/plugins"
OUT_BASE="$PLUGINS_DIR/csharp"

ensure_cs_exports() {
  local plugin_dir="$1"
  local generated_exports="$plugin_dir/__plugin_exports.generated.cs"

  if rg -n 'EntryPoint\s*=\s*"PluginLoad"' "$plugin_dir" --glob '*.cs' >/dev/null 2>&1; then
    if [[ -f "$generated_exports" ]]; then
      rm "$generated_exports"
    fi
    printf '%s' ""
    return
  fi

  cat > "$generated_exports" <<'CSWRAP'
using BedrockPlugin.Sdk.Guest;
using System.Runtime.InteropServices;

namespace GeneratedPluginEntrypoint;

public static unsafe class GeneratedPluginExports
{
    [UnmanagedCallersOnly(EntryPoint = "PluginLoad")]
    public static int PluginLoad(NativeHostApi* hostApi, byte* pluginName) => NativePlugin.Load(hostApi, pluginName);

    [UnmanagedCallersOnly(EntryPoint = "PluginUnload")]
    public static void PluginUnload() => NativePlugin.Unload();

    [UnmanagedCallersOnly(EntryPoint = "PluginDispatchEvent")]
    public static void PluginDispatchEvent(ushort version, ushort eventId, uint flags, ulong playerId, ulong requestKey, byte* payload, uint payloadLen)
        => NativePlugin.DispatchEvent(version, eventId, flags, playerId, requestKey, payload, payloadLen);
}
CSWRAP

  printf '%s' "$generated_exports"
}

shopt -s nullglob
csproj_files=("$PLUGINS_DIR"/*/*.csproj)
shopt -u nullglob

if [[ ${#csproj_files[@]} -eq 0 ]]; then
  echo "no C# plugins found under $PLUGINS_DIR"
  exit 0
fi

mkdir -p "$OUT_BASE"

built=()
for csproj in "${csproj_files[@]}"; do
  project_name="$(basename "${csproj%.csproj}")"
  plugin_dir="$(dirname "$csproj")"
  out_dir="$OUT_BASE/$project_name"
  mkdir -p "$out_dir"

  generated_exports="$(ensure_cs_exports "$plugin_dir")"
  dotnet restore "$csproj" -r linux-x64
  dotnet publish "$csproj" \
    -c Release \
    -r linux-x64 \
    -p:PublishAot=true \
    -p:SelfContained=true \
    -p:NativeLib=Shared \
    -o "$out_dir"

  artifact=""
  if [[ -f "$out_dir/$project_name.so" ]]; then
    artifact="$out_dir/$project_name.so"
  elif [[ -f "$out_dir/lib$project_name.so" ]]; then
    artifact="$out_dir/lib$project_name.so"
  else
    shopt -s nullglob
    candidates=("$out_dir"/*.so)
    shopt -u nullglob
    if [[ ${#candidates[@]} -gt 0 ]]; then
      artifact="${candidates[0]}"
    fi
  fi

  if [[ -z "$artifact" ]]; then
    echo "failed to find native C# plugin artifact (.so) for $project_name" >&2
    exit 1
  fi

  target="$out_dir/$project_name.so"
  if [[ "$artifact" != "$target" ]]; then
    cp "$artifact" "$target"
  fi

  if [[ -n "$generated_exports" && -f "$generated_exports" ]]; then
    rm "$generated_exports"
  fi

  built+=("$target")
done

echo "C# plugins built successfully:"
printf '%s\n' "${built[@]}"
