#!/usr/bin/env bash
set -euo pipefail

usage() {
	cat <<'USAGE'
Usage: ./start.sh

Starts the server inside Docker (Linux runtime), compiling plugins from ./plugins sources into a temporary directory.

Supported source layouts:
  plugins/<name>/*.go      -> built as Go plugin (.so)
  plugins/<name>/*.csproj  -> published as NativeAOT C# plugin (.so)
  plugins/<name>/*.cs      -> auto-generates csproj + export entrypoint, then publishes NativeAOT C# plugin (.so)
  plugins/*.so             -> prebuilt Go plugin fallback
  plugins/csharp/**.so     -> prebuilt C# plugin fallback

Environment overrides:
  PLUGIN_SRC_DIR      Source plugin directory (default: ./plugins)
  PLUGIN_GO_IMAGE     Go image (default: golang:1.25)
  PLUGIN_DOTNET_IMAGE Dotnet SDK image (default: mcr.microsoft.com/dotnet/sdk:8.0)
  PLUGIN_DOTNET_AOT_IMAGE_TAG Docker image tag used for NativeAOT publish (default: bedrock-plugin-dotnet-aot:8.0)
  PLUGIN_CS_RID       C# runtime identifier (default: linux-x64)
  DOTNET_INVARIANT    Set to 0 to disable invariant globalization mode (default: 1)
  SERVER_PORT         UDP listen port exposed on host (default: 19132)
  KEEP_TEMP           Set to 1 to keep temporary build/runtime directory
USAGE
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
	usage
	exit 0
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_SRC_DIR="${PLUGIN_SRC_DIR:-$ROOT_DIR/plugins}"
PLUGIN_GO_IMAGE_SET=0
if [[ -n "${PLUGIN_GO_IMAGE+x}" ]]; then
	PLUGIN_GO_IMAGE_SET=1
fi
PLUGIN_GO_IMAGE="${PLUGIN_GO_IMAGE:-golang:1.25}"
PLUGIN_DOTNET_IMAGE="${PLUGIN_DOTNET_IMAGE:-mcr.microsoft.com/dotnet/sdk:8.0}"
PLUGIN_DOTNET_AOT_IMAGE_TAG="${PLUGIN_DOTNET_AOT_IMAGE_TAG:-bedrock-plugin-dotnet-aot:8.0}"
PLUGIN_CS_RID="${PLUGIN_CS_RID:-linux-x64}"
GOEXPERIMENT_SET=0
if [[ -n "${GOEXPERIMENT+x}" ]]; then
	GOEXPERIMENT_SET=1
fi
GOEXPERIMENT_VALUE="${GOEXPERIMENT:-}"
DOTNET_INVARIANT="${DOTNET_INVARIANT:-1}"
SERVER_PORT="${SERVER_PORT:-19132}"
KEEP_TEMP="${KEEP_TEMP:-0}"

TMP_BASE="${TMP_BASE:-$ROOT_DIR/.tmp}"
RUNTIME_DIR="$TMP_BASE/start-$(date +%s)-$$-$RANDOM"
RUNTIME_PLUGIN_DIR="$RUNTIME_DIR/plugins"
RUNTIME_BUILD_DIR="$RUNTIME_DIR/build"
GO_BUILD_CACHE_DIR="$TMP_BASE/go-build-cache"
GO_MOD_CACHE_DIR="$TMP_BASE/go-mod-cache"
DOTNET_NUGET_CACHE_DIR="$TMP_BASE/dotnet-nuget-cache"

if ! command -v docker >/dev/null 2>&1; then
	echo "error: docker CLI not found. Install Docker Desktop / Docker Engine first." >&2
	exit 1
fi
if ! docker info >/dev/null 2>&1; then
	echo "error: docker daemon is not reachable. Is Docker running?" >&2
	exit 1
fi
if [[ ! -d "$PLUGIN_SRC_DIR" ]]; then
	echo "error: plugin source directory not found: $PLUGIN_SRC_DIR" >&2
	exit 1
fi
if [[ "$PLUGIN_CS_RID" != linux-* ]]; then
	echo "error: PLUGIN_CS_RID must target linux-* for this Docker Linux runtime (got: $PLUGIN_CS_RID)" >&2
	exit 1
fi

mkdir -p "$RUNTIME_PLUGIN_DIR" "$RUNTIME_PLUGIN_DIR/csharp" "$RUNTIME_BUILD_DIR"
mkdir -p "$GO_BUILD_CACHE_DIR" "$GO_MOD_CACHE_DIR" "$DOTNET_NUGET_CACHE_DIR"

PLUGIN_SRC_ABS="$(cd "$PLUGIN_SRC_DIR" && pwd)"
case "$PLUGIN_SRC_ABS" in
"$ROOT_DIR")
	PLUGIN_SRC_REL="."
	;;
"$ROOT_DIR"/*)
	PLUGIN_SRC_REL="${PLUGIN_SRC_ABS#"$ROOT_DIR"/}"
	;;
*)
	echo "error: PLUGIN_SRC_DIR must be inside repository root ($ROOT_DIR), got: $PLUGIN_SRC_ABS" >&2
	exit 1
	;;
esac
PLUGIN_SRC_CONTAINER="/workspace/$PLUGIN_SRC_REL"

cleanup() {
	if [[ "$KEEP_TEMP" == "1" ]]; then
		echo "temporary runtime directory kept at: $RUNTIME_DIR"
		return
	fi
	rm -rf "$RUNTIME_DIR"
}
trap cleanup EXIT

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
RUNTIME_MOUNT="$(host_path_for_docker "$RUNTIME_DIR")"
GO_BUILD_CACHE_MOUNT="$(host_path_for_docker "$GO_BUILD_CACHE_DIR")"
GO_MOD_CACHE_MOUNT="$(host_path_for_docker "$GO_MOD_CACHE_DIR")"
DOTNET_NUGET_CACHE_MOUNT="$(host_path_for_docker "$DOTNET_NUGET_CACHE_DIR")"

run_go() {
	docker run --rm \
		--mount "type=bind,src=$ROOT_MOUNT,dst=/workspace" \
		--mount "type=bind,src=$RUNTIME_MOUNT,dst=/runtime" \
		--mount "type=bind,src=$GO_BUILD_CACHE_MOUNT,dst=/root/.cache/go-build" \
		--mount "type=bind,src=$GO_MOD_CACHE_MOUNT,dst=/go/pkg/mod" \
		-w /workspace \
		-e CGO_ENABLED=1 \
		-e GOFLAGS=-buildvcs=false \
		-e GOEXPERIMENT="$GOEXPERIMENT_VALUE" \
		"$PLUGIN_GO_IMAGE" \
		"$@"
}

run_dotnet() {
	ensure_dotnet_aot_image
	docker run --rm \
		--mount "type=bind,src=$ROOT_MOUNT,dst=/workspace" \
		--mount "type=bind,src=$RUNTIME_MOUNT,dst=/runtime" \
		--mount "type=bind,src=$DOTNET_NUGET_CACHE_MOUNT,dst=/root/.nuget/packages" \
		-w /workspace \
		"$PLUGIN_DOTNET_AOT_IMAGE_TAG" \
		"$@"
}

ensure_dotnet_aot_image() {
	if docker image inspect "$PLUGIN_DOTNET_AOT_IMAGE_TAG" >/dev/null 2>&1; then
		return
	fi
	echo "building NativeAOT dotnet image: $PLUGIN_DOTNET_AOT_IMAGE_TAG"
	DOCKER_BUILDKIT=1 docker build -t "$PLUGIN_DOTNET_AOT_IMAGE_TAG" - >/dev/null <<DOCKER
FROM $PLUGIN_DOTNET_IMAGE
RUN apt-get update \
	&& DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends clang zlib1g-dev \
	&& rm -rf /var/lib/apt/lists/*
DOCKER
}

collect_source_go_plugin_names() {
	shopt -s nullglob
	local entries=("$PLUGIN_SRC_DIR"/*)
	shopt -u nullglob
	for entry in "${entries[@]}"; do
		[[ -d "$entry" ]] || continue
		shopt -s nullglob
		local go_files=("$entry"/*.go)
		shopt -u nullglob
		if [[ ${#go_files[@]} -eq 0 ]]; then
			continue
		fi
		local name
		name="$(basename "$entry")"
		SOURCE_GO_NAMES["$name"]=1
	done
}

detect_prebuilt_go_toolchain() {
	shopt -s nullglob
	local prebuilt_root=("$PLUGIN_SRC_DIR"/*.so)
	shopt -u nullglob
	if [[ ${#prebuilt_root[@]} -eq 0 ]]; then
		return
	fi

	local detected_version=""
	local detected_experiment=""
	local scanned=0

	for so_file in "${prebuilt_root[@]}"; do
		local name
		name="$(basename "${so_file%.so}")"
		if [[ -n "${SOURCE_GO_NAMES[$name]:-}" ]]; then
			continue
		fi

		local so_container="$PLUGIN_SRC_CONTAINER/$(basename "$so_file")"
		local meta
		if ! meta="$(run_go go version -m "$so_container" 2>&1)"; then
			echo "error: failed to inspect prebuilt Go plugin metadata: $so_file" >&2
			echo "$meta" >&2
			exit 1
		fi

		local version
		version="$(printf '%s\n' "$meta" | awk 'NR==1{for(i=1;i<=NF;i++){if($i ~ /^go[0-9]/){print $i; exit}}}')"
		if [[ -z "$version" ]]; then
			echo "error: could not read Go toolchain version from prebuilt plugin: $so_file" >&2
			exit 1
		fi
		if [[ -z "$detected_version" ]]; then
			detected_version="$version"
		elif [[ "$detected_version" != "$version" ]]; then
			echo "error: prebuilt Go plugins use different Go versions ($detected_version vs $version)." >&2
			echo "       Rebuild them with one toolchain, or keep only one toolchain's artifacts." >&2
			exit 1
		fi

		local experiment
		experiment="$(printf '%s\n' "$meta" | awk -F= '/^[[:space:]]*build[[:space:]]+GOEXPERIMENT=/{print $2; exit}')"
		if [[ -n "$experiment" ]]; then
			if [[ -z "$detected_experiment" ]]; then
				detected_experiment="$experiment"
			elif [[ "$detected_experiment" != "$experiment" ]]; then
				echo "error: prebuilt Go plugins use different GOEXPERIMENT values ($detected_experiment vs $experiment)." >&2
				exit 1
			fi
		fi

		scanned=1
	done

	if [[ "$scanned" -eq 0 ]]; then
		return
	fi

	if [[ "$PLUGIN_GO_IMAGE_SET" != "1" ]]; then
		PLUGIN_GO_IMAGE="golang:${detected_version#go}"
	fi
	if [[ "$GOEXPERIMENT_SET" != "1" && -n "$detected_experiment" ]]; then
		GOEXPERIMENT_VALUE="$detected_experiment"
	fi

	echo "using prebuilt Go plugin toolchain: $detected_version${detected_experiment:+ (GOEXPERIMENT=$detected_experiment)}"
	echo "using Go image: $PLUGIN_GO_IMAGE"
}

resolve_linux_so() {
	local out_dir="$1"
	local base_name="$2"
	local candidate=""

	if [[ -f "$out_dir/$base_name.so" ]]; then
		candidate="$out_dir/$base_name.so"
	elif [[ -f "$out_dir/lib$base_name.so" ]]; then
		candidate="$out_dir/lib$base_name.so"
	else
		shopt -s nullglob
		local so_candidates=("$out_dir"/*.so)
		shopt -u nullglob
		if [[ ${#so_candidates[@]} -gt 0 ]]; then
			candidate="${so_candidates[0]}"
		fi
	fi

	printf '%s' "$candidate"
}

publish_csproj() {
	local csproj_container_path="$1"
	local project_name="$2"
	local plugin_group="$3"

	local out_dir="$RUNTIME_BUILD_DIR/csharp/$plugin_group/$project_name"
	local out_container="/runtime/build/csharp/$plugin_group/$project_name"
	mkdir -p "$out_dir"

	local publish_output
	if ! publish_output="$(run_dotnet dotnet publish "$csproj_container_path" \
		-c Release \
		-o "$out_container" \
		-r "$PLUGIN_CS_RID" \
		/p:PublishAot=true \
		/p:NativeLib=Shared \
		/p:SelfContained=true \
		--nologo 2>&1)"; then
		printf '%s\n' "$publish_output" >&2
		return 1
	fi

	local artifact
	artifact="$(resolve_linux_so "$out_dir" "$project_name")"
	if [[ -z "$artifact" ]]; then
		echo "error: failed to find C# plugin artifact (.so) for $project_name" >&2
		exit 1
	fi

	local target_dir="$RUNTIME_PLUGIN_DIR/csharp/$project_name"
	mkdir -p "$target_dir"
	cp "$artifact" "$target_dir/$project_name.so"
	BUILT_CS+=("$target_dir/$project_name.so")
}

generate_and_publish_cs() {
	local plugin_dir="$1"
	local plugin_name="$2"
	local gen_dir="$RUNTIME_BUILD_DIR/generated-cs/$plugin_name"
	local out_dir="$RUNTIME_BUILD_DIR/generated-cs-out/$plugin_name"

	mkdir -p "$gen_dir" "$out_dir"
	shopt -s nullglob
	local cs_files=("$plugin_dir"/*.cs)
	shopt -u nullglob
	if [[ ${#cs_files[@]} -eq 0 ]]; then
		return
	fi
	cp "${cs_files[@]}" "$gen_dir/"
	cat > "$gen_dir/__plugin_exports.generated.cs" <<'CSWRAP'
using BedrockPlugin.Sdk.Guest;
using System.Linq;
using System.Reflection;
using System.Runtime.CompilerServices;
using System.Runtime.InteropServices;

namespace GeneratedPluginEntrypoint;

public static unsafe class GeneratedPluginExports
{
    [ModuleInitializer]
    internal static void RegisterFactory()
    {
        NativePlugin.Register((name, host) =>
        {
            var pluginType = Assembly
                .GetExecutingAssembly()
                .GetTypes()
                .FirstOrDefault(t => !t.IsAbstract && typeof(Plugin).IsAssignableFrom(t));
            if (pluginType is null)
            {
                throw new InvalidOperationException("No non-abstract type deriving BedrockPlugin.Sdk.Guest.Plugin was found.");
            }
            var ctor = pluginType.GetConstructor(new[] { typeof(string), typeof(IGuestHost) });
            if (ctor is null)
            {
                throw new InvalidOperationException($"Type {pluginType.FullName} must define constructor (string name, IGuestHost host).");
            }
            return (Plugin)ctor.Invoke(new object?[] { name, host });
        });
    }

    [UnmanagedCallersOnly(EntryPoint = "PluginLoad")]
    public static int PluginLoad(NativeHostApi* hostApi, byte* pluginName) => NativePlugin.Load(hostApi, pluginName);

    [UnmanagedCallersOnly(EntryPoint = "PluginUnload")]
    public static void PluginUnload() => NativePlugin.Unload();

    [UnmanagedCallersOnly(EntryPoint = "PluginDispatchEvent")]
    public static void PluginDispatchEvent(ushort version, ushort eventId, uint flags, ulong playerId, ulong requestKey, byte* payload, uint payloadLen)
        => NativePlugin.DispatchEvent(version, eventId, flags, playerId, requestKey, payload, payloadLen);
}
CSWRAP

	cat > "$gen_dir/$plugin_name.csproj" <<CSPROJ
<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
    <ImplicitUsings>enable</ImplicitUsings>
    <Nullable>enable</Nullable>
    <LangVersion>latest</LangVersion>
    <AllowUnsafeBlocks>true</AllowUnsafeBlocks>
    <OutputType>Library</OutputType>
    <PublishAot>true</PublishAot>
    <SelfContained>true</SelfContained>
    <NativeLib>Shared</NativeLib>
    <StripSymbols>false</StripSymbols>
  </PropertyGroup>
  <ItemGroup>
    <ProjectReference Include="/workspace/plugin/sdk/csharp/BedrockPlugin.Abi.csproj" />
  </ItemGroup>
</Project>
CSPROJ

	publish_csproj "/runtime/build/generated-cs/$plugin_name/$plugin_name.csproj" "$plugin_name" "$plugin_name"
}

compile_plugin_sources() {
	shopt -s nullglob
	local entries=("$PLUGIN_SRC_DIR"/*)
	shopt -u nullglob

	for entry in "${entries[@]}"; do
		[[ -d "$entry" ]] || continue

		local name
		name="$(basename "$entry")"

		shopt -s nullglob
		local go_files=("$entry"/*.go)
		local csproj_files=("$entry"/*.csproj)
		local cs_files=("$entry"/*.cs)
		shopt -u nullglob

		if [[ ${#go_files[@]} -gt 0 ]]; then
			echo "building Go plugin from $PLUGIN_SRC_REL/$name"
			run_go go build -buildmode=plugin -o "/runtime/plugins/$name.so" "$PLUGIN_SRC_CONTAINER/$name"
			BUILT_GO+=("$RUNTIME_PLUGIN_DIR/$name.so")
		fi

		if [[ ${#csproj_files[@]} -gt 0 ]]; then
			for csproj in "${csproj_files[@]}"; do
				local project_name
				project_name="$(basename "${csproj%.csproj}")"
				echo "building C# plugin from $PLUGIN_SRC_REL/$name/$(basename "$csproj")"
				publish_csproj "$PLUGIN_SRC_CONTAINER/$name/$project_name.csproj" "$project_name" "$name"
			done
		elif [[ ${#cs_files[@]} -gt 0 ]]; then
			echo "building C# plugin from $PLUGIN_SRC_REL/$name/*.cs"
			generate_and_publish_cs "$entry" "$name"
		fi
	done
}

copy_prebuilt_artifacts() {
	shopt -s nullglob
	local prebuilt_root=("$PLUGIN_SRC_DIR"/*.so)
	shopt -u nullglob
	for so_file in "${prebuilt_root[@]}"; do
		local name
		name="$(basename "${so_file%.so}")"
		if [[ -n "${BUILT_BY_NAME[$name]:-}" ]]; then
			continue
		fi
		local dst="$RUNTIME_PLUGIN_DIR/$(basename "$so_file")"
		cp "$so_file" "$dst"
		COPIED_PREBUILT+=("$dst")
		PREBUILT_GO_COPIED+=("$dst")
	done

	if [[ -d "$PLUGIN_SRC_DIR/csharp" ]]; then
		while IFS= read -r -d '' so_file; do
			local name
			name="$(basename "${so_file%.so}")"
			if [[ -n "${BUILT_BY_NAME[$name]:-}" ]]; then
				continue
			fi
			local rel="${so_file#"$PLUGIN_SRC_DIR"/}"
			local dst="$RUNTIME_PLUGIN_DIR/$rel"
			mkdir -p "$(dirname "$dst")"
			cp "$so_file" "$dst"
			COPIED_PREBUILT+=("$dst")
		done < <(find "$PLUGIN_SRC_DIR/csharp" -type f -name '*.so' -print0)
	fi
}

validate_prebuilt_go_artifacts() {
	if [[ ${#PREBUILT_GO_COPIED[@]} -eq 0 ]]; then
		return
	fi

	cat > "$RUNTIME_BUILD_DIR/check_prebuilt_go_plugin.go" <<'GOLOAD'
package main

import (
	"fmt"
	"os"
	goplugin "plugin"

	_ "github.com/bedrock-gophers/plugin/plugin/abi"
	_ "github.com/bedrock-gophers/plugin/plugin/sdk/go"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: check_prebuilt_go_plugin <plugin.so>")
		os.Exit(2)
	}
	if _, err := goplugin.Open(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
GOLOAD

	local kept=()
	for host_so in "${PREBUILT_GO_COPIED[@]}"; do
		local runtime_so="/runtime/plugins/$(basename "$host_so")"
		local out
		if out="$(run_go go run /runtime/build/check_prebuilt_go_plugin.go "$runtime_so" 2>&1)"; then
			kept+=("$host_so")
			continue
		fi
		echo "warning: skipping incompatible prebuilt Go plugin: $host_so" >&2
		echo "$out" >&2
		rm -f "$host_so"
	done

	PREBUILT_GO_COPIED=("${kept[@]}")
	local filtered=()
	for artifact in "${COPIED_PREBUILT[@]}"; do
		if [[ -f "$artifact" ]]; then
			filtered+=("$artifact")
		fi
	done
	COPIED_PREBUILT=("${filtered[@]}")
}

BUILT_GO=()
BUILT_CS=()
COPIED_PREBUILT=()
PREBUILT_GO_COPIED=()
declare -A BUILT_BY_NAME=()
declare -A SOURCE_GO_NAMES=()

collect_source_go_plugin_names
detect_prebuilt_go_toolchain
compile_plugin_sources
for so_path in "${BUILT_GO[@]}"; do
	name="$(basename "${so_path%.so}")"
	BUILT_BY_NAME["$name"]=1
done
for so_path in "${BUILT_CS[@]}"; do
	name="$(basename "${so_path%.so}")"
	BUILT_BY_NAME["$name"]=1
done
copy_prebuilt_artifacts
validate_prebuilt_go_artifacts

echo "temporary plugin dir: $RUNTIME_PLUGIN_DIR"
echo "resolved plugin artifacts:"
if [[ ${#BUILT_GO[@]} -gt 0 ]]; then
	printf '  %s\n' "${BUILT_GO[@]}"
fi
if [[ ${#BUILT_CS[@]} -gt 0 ]]; then
	printf '  %s\n' "${BUILT_CS[@]}"
fi
if [[ ${#COPIED_PREBUILT[@]} -gt 0 ]]; then
	printf '  %s\n' "${COPIED_PREBUILT[@]}"
fi
if [[ ${#BUILT_GO[@]} -eq 0 && ${#BUILT_CS[@]} -eq 0 && ${#COPIED_PREBUILT[@]} -eq 0 ]]; then
	echo "  none"
fi

tty_flags=()
if [[ -t 0 && -t 1 ]]; then
	tty_flags=(-it)
fi

echo "starting server in Docker on UDP $SERVER_PORT"
docker run --rm --init "${tty_flags[@]}" \
	--mount "type=bind,src=$ROOT_MOUNT,dst=/workspace" \
	--mount "type=bind,src=$RUNTIME_MOUNT,dst=/runtime" \
	--mount "type=bind,src=$GO_BUILD_CACHE_MOUNT,dst=/root/.cache/go-build" \
	--mount "type=bind,src=$GO_MOD_CACHE_MOUNT,dst=/go/pkg/mod" \
	-w /workspace \
	-e CGO_ENABLED=1 \
	-e GOFLAGS=-buildvcs=false \
	-e GOEXPERIMENT="$GOEXPERIMENT_VALUE" \
	-e PLUGIN_DIR=/runtime/plugins \
	-e DOTNET_SYSTEM_GLOBALIZATION_INVARIANT="$DOTNET_INVARIANT" \
	-p "$SERVER_PORT:19132/udp" \
	"$PLUGIN_GO_IMAGE" \
	go run ./cmd
