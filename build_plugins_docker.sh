#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGINS_DIR="$ROOT_DIR/plugins"
OUT_BASE="$PLUGINS_DIR/csharp"
PLUGIN_DOTNET_IMAGE="${PLUGIN_DOTNET_IMAGE:-mcr.microsoft.com/dotnet/sdk:8.0}"
PLUGIN_CS_RID="${PLUGIN_CS_RID:-linux-x64}"
PLUGIN_DOTNET_AOT_IMAGE_TAG="${PLUGIN_DOTNET_AOT_IMAGE_TAG:-bedrock-plugin-dotnet-aot:8.0}"
TMP_BASE="${TMP_BASE:-$ROOT_DIR/.tmp}"
DOTNET_NUGET_CACHE_DIR="$TMP_BASE/dotnet-nuget-cache"

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

if ! command -v docker >/dev/null 2>&1; then
  echo "error: docker CLI not found" >&2
  exit 1
fi
if ! docker info >/dev/null 2>&1; then
  echo "error: docker daemon is not reachable" >&2
  exit 1
fi
if [[ "$PLUGIN_CS_RID" != linux-* ]]; then
  echo "error: PLUGIN_CS_RID must target linux-* (got: $PLUGIN_CS_RID)" >&2
  exit 1
fi

mkdir -p "$OUT_BASE" "$DOTNET_NUGET_CACHE_DIR"

shopt -s nullglob
csproj_files=("$PLUGINS_DIR"/*/*.csproj)
shopt -u nullglob

if [[ ${#csproj_files[@]} -eq 0 ]]; then
  echo "no C# plugins found under $PLUGINS_DIR"
  exit 0
fi

host_path_for_docker() {
  local path="$1"
  case "$(uname -s)" in
  MINGW*|MSYS*|CYGWIN*)
    if command -v cygpath >/dev/null 2>&1; then
      cygpath -m "$path"
    else
      printf '%s' "$path"
    fi
    ;;
  *)
    printf '%s' "$path"
    ;;
  esac
}

ROOT_MOUNT="$(host_path_for_docker "$ROOT_DIR")"
DOTNET_NUGET_CACHE_MOUNT="$(host_path_for_docker "$DOTNET_NUGET_CACHE_DIR")"

ensure_dotnet_aot_image() {
  if docker image inspect "$PLUGIN_DOTNET_AOT_IMAGE_TAG" >/dev/null 2>&1; then
    return
  fi

  DOCKER_BUILDKIT=1 docker build -t "$PLUGIN_DOTNET_AOT_IMAGE_TAG" - >/dev/null <<DOCKER
FROM $PLUGIN_DOTNET_IMAGE
RUN apt-get update \
  && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends clang zlib1g-dev \
  && rm -rf /var/lib/apt/lists/*
DOCKER
}

ensure_dotnet_aot_image

built=()
for csproj in "${csproj_files[@]}"; do
  project_name="$(basename "${csproj%.csproj}")"
  plugin_dir="$(dirname "$csproj")"
  rel_csproj="${csproj#$ROOT_DIR/}"
  out_dir="$OUT_BASE/$project_name"
  mkdir -p "$out_dir"

  generated_exports="$(ensure_cs_exports "$plugin_dir")"
  docker run --rm \
    --mount "type=bind,src=$ROOT_MOUNT,dst=/workspace" \
    --mount "type=bind,src=$DOTNET_NUGET_CACHE_MOUNT,dst=/root/.nuget/packages" \
    -w /workspace \
    "$PLUGIN_DOTNET_AOT_IMAGE_TAG" \
    /bin/bash -lc "dotnet restore \"/workspace/$rel_csproj\" -r \"$PLUGIN_CS_RID\" --nologo >/dev/null && dotnet publish \"/workspace/$rel_csproj\" -c Release -o \"/workspace/plugins/csharp/$project_name\" -r \"$PLUGIN_CS_RID\" /p:PublishAot=true /p:NativeLib=Shared /p:SelfContained=true --nologo >/dev/null"

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

echo "C# plugins built with Docker:"
printf '%s\n' "${built[@]}"
