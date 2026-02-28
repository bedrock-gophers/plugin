#!/usr/bin/env bash
set -euo pipefail

usage() {
	cat <<'USAGE'
Usage: ./start.sh

Starts the server inside Docker (Linux runtime), compiling plugins into a temporary directory.
The server/runtime source is copied from a local cached clone of github.com/bedrock-gophers/plugin.

Supported source layouts:
  plugins/<name>/*.go      -> built as Go plugin (.so)
  plugins/<name>/*.csproj  -> published as NativeAOT C# plugin (.so)
  plugins/<name>/*.cs      -> auto-generates csproj + export entrypoint, then publishes NativeAOT C# plugin (.so)
  plugins/*.so             -> prebuilt Go plugin fallback
  plugins/csharp/**.so     -> prebuilt C# plugin fallback

Environment overrides:
  PLUGIN_SRC_DIR      Source plugin directory (default: $PWD/plugins, fallback: cached repo ./plugins)
  HOST_REPO_URL       Host runtime repo URL (default: https://github.com/bedrock-gophers/plugin.git)
  USE_LOCAL_HOST_WORKTREE Set to 1 to use the current working tree as host runtime source
                      (default: 0, uses cached/fetched HOST_REPO_URL source)
  BRANCH              Optional git branch/tag to clone from HOST_REPO_URL
  COMMIT              Optional commit SHA to checkout after clone (takes precedence over BRANCH tip)
  HOST_REPO_CACHE_DIR Cached host repo directory (default: $TMP_BASE/host-repo-cache)
  HOST_REPO_UPDATE    Set to 1 to manually update HOST_REPO_CACHE_DIR from origin
  PLUGIN_ARTIFACT_CACHE_DIR Cached plugin artifact directory (default: $TMP_BASE/plugin-artifact-cache)
  PLUGIN_SHA_CACHE_DIR Plugin source hash directory (default: $TMP_BASE/plugin-sha-cache)
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

PLUGIN_SRC_DIR_SET=0
if [[ -n "${PLUGIN_SRC_DIR+x}" ]]; then
	PLUGIN_SRC_DIR_SET=1
fi
PLUGIN_SRC_DIR="${PLUGIN_SRC_DIR:-$PWD/plugins}"
HOST_REPO_URL="${HOST_REPO_URL:-https://github.com/bedrock-gophers/plugin.git}"
USE_LOCAL_HOST_WORKTREE="${USE_LOCAL_HOST_WORKTREE:-0}"
BRANCH="${BRANCH:-}"
COMMIT="${COMMIT:-}"
HOST_REPO_UPDATE="${HOST_REPO_UPDATE:-0}"
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
DOCKER_RUN_USER="$(id -u):$(id -g)"

TMP_BASE="${TMP_BASE:-${XDG_CACHE_HOME:-$HOME/.cache}/bedrock-plugin}"
RUNTIME_DIR="$TMP_BASE/start-$(date +%s)-$$-$RANDOM"
RUNTIME_PLUGIN_DIR="$RUNTIME_DIR/plugins"
RUNTIME_BUILD_DIR="$RUNTIME_DIR/build"
HOST_REPO_DIR="$RUNTIME_BUILD_DIR/host-repo"
STAGED_PLUGIN_SRC_REL=".start-plugin-src"
STAGED_PLUGIN_SRC_DIR="$HOST_REPO_DIR/$STAGED_PLUGIN_SRC_REL"
GO_BUILD_CACHE_DIR="$TMP_BASE/go-build-cache"
GO_MOD_CACHE_DIR="$TMP_BASE/go-mod-cache"
DOTNET_NUGET_CACHE_DIR="$TMP_BASE/dotnet-nuget-cache"
HOST_REPO_CACHE_DIR="${HOST_REPO_CACHE_DIR:-$TMP_BASE/host-repo-cache}"
PLUGIN_ARTIFACT_CACHE_DIR="${PLUGIN_ARTIFACT_CACHE_DIR:-$TMP_BASE/plugin-artifact-cache}"
PLUGIN_SHA_CACHE_DIR="${PLUGIN_SHA_CACHE_DIR:-$TMP_BASE/plugin-sha-cache}"
HOST_RUNTIME_ABI_SHA=""

if [[ "$USE_LOCAL_HOST_WORKTREE" != "0" && "$USE_LOCAL_HOST_WORKTREE" != "1" ]]; then
	echo "error: USE_LOCAL_HOST_WORKTREE must be 0 or 1 (got: $USE_LOCAL_HOST_WORKTREE)" >&2
	exit 1
fi

if ! command -v git >/dev/null 2>&1; then
	echo "error: git CLI not found. Install git first." >&2
	exit 1
fi
if ! command -v docker >/dev/null 2>&1; then
	echo "error: docker CLI not found. Install Docker Desktop / Docker Engine first." >&2
	exit 1
fi
if ! docker info >/dev/null 2>&1; then
	echo "error: docker daemon is not reachable. Is Docker running?" >&2
	exit 1
fi
if [[ "$PLUGIN_CS_RID" != linux-* ]]; then
	echo "error: PLUGIN_CS_RID must target linux-* for this Docker Linux runtime (got: $PLUGIN_CS_RID)" >&2
	exit 1
fi

mkdir -p "$RUNTIME_PLUGIN_DIR" "$RUNTIME_PLUGIN_DIR/csharp" "$RUNTIME_BUILD_DIR" "$HOST_REPO_DIR" "$HOST_REPO_CACHE_DIR"
mkdir -p "$GO_BUILD_CACHE_DIR" "$GO_MOD_CACHE_DIR" "$DOTNET_NUGET_CACHE_DIR"
mkdir -p "$PLUGIN_ARTIFACT_CACHE_DIR" "$PLUGIN_SHA_CACHE_DIR"

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

ensure_host_repo_cache() {
	if [[ "$USE_LOCAL_HOST_WORKTREE" == "1" ]]; then
		return
	fi

	if [[ -d "$HOST_REPO_CACHE_DIR/.git" ]]; then
		local current_origin
		current_origin="$(git -C "$HOST_REPO_CACHE_DIR" remote get-url origin 2>/dev/null || true)"
		if [[ -n "$current_origin" && "$current_origin" != "$HOST_REPO_URL" ]]; then
			rm -rf "$HOST_REPO_CACHE_DIR"
		fi
	elif [[ -d "$HOST_REPO_CACHE_DIR" ]] && [[ -n "$(ls -A "$HOST_REPO_CACHE_DIR" 2>/dev/null)" ]]; then
		rm -rf "$HOST_REPO_CACHE_DIR"
	fi

	if [[ ! -d "$HOST_REPO_CACHE_DIR/.git" ]]; then
		local clone_args=("--depth" "1")
		if [[ -n "$BRANCH" ]]; then
			clone_args+=("--branch" "$BRANCH")
		fi
		echo "initializing host runtime repo cache: $HOST_REPO_CACHE_DIR"
		git clone --quiet "${clone_args[@]}" "$HOST_REPO_URL" "$HOST_REPO_CACHE_DIR"
	elif [[ "$HOST_REPO_UPDATE" == "1" ]]; then
		echo "updating host runtime repo cache from origin"
		if [[ -n "$BRANCH" ]]; then
			git -C "$HOST_REPO_CACHE_DIR" fetch --quiet --depth 1 origin "$BRANCH"
		else
			git -C "$HOST_REPO_CACHE_DIR" fetch --quiet --depth 1 origin
		fi
	fi

	if [[ -n "$BRANCH" ]] && ! git -C "$HOST_REPO_CACHE_DIR" show-ref --verify --quiet "refs/remotes/origin/$BRANCH"; then
		if [[ "$HOST_REPO_UPDATE" == "1" ]]; then
			git -C "$HOST_REPO_CACHE_DIR" fetch --quiet --depth 1 origin "$BRANCH"
		else
			echo "error: branch '$BRANCH' not found in cached host repo. Set HOST_REPO_UPDATE=1 to refresh cache." >&2
			exit 1
		fi
	fi

	if [[ -n "$COMMIT" ]] && ! git -C "$HOST_REPO_CACHE_DIR" rev-parse --verify --quiet "$COMMIT^{commit}" >/dev/null; then
		if [[ "$HOST_REPO_UPDATE" == "1" ]]; then
			git -C "$HOST_REPO_CACHE_DIR" fetch --quiet --depth 1 origin "$COMMIT"
		else
			echo "error: commit '$COMMIT' not found in cached host repo. Set HOST_REPO_UPDATE=1 to refresh cache." >&2
			exit 1
		fi
	fi
}

prepare_host_repo() {
	rm -rf "$HOST_REPO_DIR"
	if [[ "$USE_LOCAL_HOST_WORKTREE" == "1" ]]; then
		echo "using local host runtime source from working tree: $PWD"
		if command -v rsync >/dev/null 2>&1; then
			rsync -a --delete \
				--exclude ".git" \
				--exclude ".tmp" \
				--exclude ".data" \
				"$PWD"/ "$HOST_REPO_DIR"/
		else
			mkdir -p "$HOST_REPO_DIR"
			cp -a "$PWD"/. "$HOST_REPO_DIR"/
			rm -rf "$HOST_REPO_DIR/.git" "$HOST_REPO_DIR/.tmp" "$HOST_REPO_DIR/.data"
		fi
		local resolved_ref
		resolved_ref="$(git -C "$PWD" rev-parse --short=12 HEAD 2>/dev/null || echo "worktree")"
		echo "using host runtime source: local worktree ($resolved_ref)"
		return
	fi

	echo "copying host runtime repo from cache"
	git clone --quiet --no-hardlinks "$HOST_REPO_CACHE_DIR" "$HOST_REPO_DIR"

	if [[ -n "$COMMIT" ]]; then
		git -C "$HOST_REPO_DIR" checkout --quiet --detach "$COMMIT"
	elif [[ -n "$BRANCH" ]]; then
		git -C "$HOST_REPO_DIR" checkout --quiet --detach "origin/$BRANCH"
	fi
	if [[ ! -f "$HOST_REPO_DIR/cmd/main.go" ]]; then
		echo "error: cloned host repo is missing cmd/main.go" >&2
		exit 1
	fi

	local resolved_ref
	resolved_ref="$(git -C "$HOST_REPO_DIR" rev-parse --short=12 HEAD)"
	echo "using host runtime ref: $resolved_ref"
}

resolve_plugin_source_dir() {
	if [[ -d "$PLUGIN_SRC_DIR" ]]; then
		return
	fi
	if [[ "$PLUGIN_SRC_DIR_SET" == "0" && -d "$HOST_REPO_DIR/plugins" ]]; then
		PLUGIN_SRC_DIR="$HOST_REPO_DIR/plugins"
		return
	fi
	echo "error: plugin source directory not found: $PLUGIN_SRC_DIR" >&2
	exit 1
}

stage_plugin_sources() {
	rm -rf "$STAGED_PLUGIN_SRC_DIR"
	mkdir -p "$STAGED_PLUGIN_SRC_DIR"
	cp -a "$PLUGIN_SRC_DIR"/. "$STAGED_PLUGIN_SRC_DIR"/
	PLUGIN_SRC_DIR="$STAGED_PLUGIN_SRC_DIR"
	PLUGIN_SRC_REL="$STAGED_PLUGIN_SRC_REL"
	PLUGIN_SRC_CONTAINER="/workspace/$PLUGIN_SRC_REL"
}

ensure_host_repo_cache
prepare_host_repo
resolve_plugin_source_dir
stage_plugin_sources

ROOT_MOUNT="$(host_path_for_docker "$HOST_REPO_DIR")"
RUNTIME_MOUNT="$(host_path_for_docker "$RUNTIME_DIR")"
GO_BUILD_CACHE_MOUNT="$(host_path_for_docker "$GO_BUILD_CACHE_DIR")"
GO_MOD_CACHE_MOUNT="$(host_path_for_docker "$GO_MOD_CACHE_DIR")"
DOTNET_NUGET_CACHE_MOUNT="$(host_path_for_docker "$DOTNET_NUGET_CACHE_DIR")"

run_go() {
	docker run --rm \
		--user "$DOCKER_RUN_USER" \
		--mount "type=bind,src=$ROOT_MOUNT,dst=/workspace" \
		--mount "type=bind,src=$RUNTIME_MOUNT,dst=/runtime" \
		--mount "type=bind,src=$GO_BUILD_CACHE_MOUNT,dst=/cache/go-build" \
		--mount "type=bind,src=$GO_MOD_CACHE_MOUNT,dst=/cache/go-mod" \
		-w /workspace \
		-e CGO_ENABLED=1 \
		-e GOFLAGS=-buildvcs=false \
		-e GOEXPERIMENT="$GOEXPERIMENT_VALUE" \
		-e GOCACHE=/cache/go-build \
		-e GOMODCACHE=/cache/go-mod \
		"$PLUGIN_GO_IMAGE" \
		"$@"
}

run_dotnet() {
	ensure_dotnet_aot_image
	docker run --rm \
		--user "$DOCKER_RUN_USER" \
		--mount "type=bind,src=$ROOT_MOUNT,dst=/workspace" \
		--mount "type=bind,src=$RUNTIME_MOUNT,dst=/runtime" \
		--mount "type=bind,src=$DOTNET_NUGET_CACHE_MOUNT,dst=/cache/nuget" \
		-w /workspace \
		-e HOME=/tmp \
		-e DOTNET_CLI_HOME=/tmp \
		-e DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1 \
		-e DOTNET_CLI_TELEMETRY_OPTOUT=1 \
		-e NUGET_PACKAGES=/cache/nuget \
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

hash_text_sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum | awk '{print $1}'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 | awk '{print $1}'
	else
		echo "error: neither sha256sum nor shasum is available to hash plugin sources." >&2
		exit 1
	fi
}

hash_file_sha256() {
	local file="$1"
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$file" | awk '{print $1}'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$file" | awk '{print $1}'
	else
		echo "error: neither sha256sum nor shasum is available to hash plugin sources." >&2
		exit 1
	fi
}

host_runtime_abi_sha() {
	local digest_input=""
	local found=0

	while IFS= read -r -d '' source_file; do
		found=1
		local rel="${source_file#"$HOST_REPO_DIR"/}"
		local file_sha
		file_sha="$(hash_file_sha256 "$source_file")"
		digest_input+="$rel:$file_sha"$'\n'
	done < <(find "$HOST_REPO_DIR/plugin" -type f -name '*.go' -print0 | sort -z)

	for mod_file in "$HOST_REPO_DIR/go.mod" "$HOST_REPO_DIR/go.sum"; do
		[[ -f "$mod_file" ]] || continue
		found=1
		local rel="${mod_file#"$HOST_REPO_DIR"/}"
		local file_sha
		file_sha="$(hash_file_sha256 "$mod_file")"
		digest_input+="$rel:$file_sha"$'\n'
	done

	if [[ "$found" -eq 0 ]]; then
		printf ''
		return
	fi
	printf '%s' "$digest_input" | hash_text_sha256
}

plugin_code_sha() {
	local plugin_dir="$1"
	local digest_input=""
	local found=0
	while IFS= read -r -d '' source_file; do
		found=1
		local rel="${source_file#"$plugin_dir"/}"
		local file_sha
		file_sha="$(hash_file_sha256 "$source_file")"
		digest_input+="$rel:$file_sha"$'\n'
	done < <(find "$plugin_dir" -maxdepth 1 -type f \( -name '*.go' -o -name '*.cs' \) -print0 | sort -z)

	if [[ "$found" -eq 0 ]]; then
		printf ''
		return
	fi

	printf '%s' "$digest_input" | hash_text_sha256
}

cache_key_for_id() {
	local id="$1"
	id="${id//\//__}"
	id="${id//\\/__}"
	printf '%s' "$id"
}

plugin_sha_file_path() {
	local kind="$1"
	local id="$2"
	local safe_id
	safe_id="$(cache_key_for_id "$id")"
	mkdir -p "$PLUGIN_SHA_CACHE_DIR/$kind"
	printf '%s' "$PLUGIN_SHA_CACHE_DIR/$kind/$safe_id.sha"
}

plugin_artifact_cache_path() {
	local kind="$1"
	local id="$2"
	local sha="$3"
	local safe_id
	safe_id="$(cache_key_for_id "$id")"
	mkdir -p "$PLUGIN_ARTIFACT_CACHE_DIR/$kind/$safe_id"
	printf '%s' "$PLUGIN_ARTIFACT_CACHE_DIR/$kind/$safe_id/$sha.so"
}

restore_cached_plugin_artifact() {
	local kind="$1"
	local id="$2"
	local sha="$3"
	local dest="$4"
	local sha_file
	sha_file="$(plugin_sha_file_path "$kind" "$id")"
	if [[ ! -f "$sha_file" ]]; then
		return 1
	fi
	local cached_sha
	cached_sha="$(cat "$sha_file" 2>/dev/null || true)"
	if [[ "$cached_sha" != "$sha" ]]; then
		return 1
	fi
	local cache_path
	cache_path="$(plugin_artifact_cache_path "$kind" "$id" "$sha")"
	if [[ ! -f "$cache_path" ]]; then
		return 1
	fi
	mkdir -p "$(dirname "$dest")"
	cp "$cache_path" "$dest"
	return 0
}

store_cached_plugin_artifact() {
	local kind="$1"
	local id="$2"
	local sha="$3"
	local src="$4"
	local sha_file
	sha_file="$(plugin_sha_file_path "$kind" "$id")"
	local cache_path
	cache_path="$(plugin_artifact_cache_path "$kind" "$id" "$sha")"
	mkdir -p "$(dirname "$cache_path")"
	cp "$src" "$cache_path"
	printf '%s\n' "$sha" >"$sha_file"
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

ensure_csproj_exports() {
	local plugin_dir="$1"
	local generated_exports="$plugin_dir/__plugin_exports.generated.cs"

	shopt -s nullglob
	local cs_files=("$plugin_dir"/*.cs)
	shopt -u nullglob

	if [[ ${#cs_files[@]} -gt 0 ]] && rg -n 'EntryPoint\s*=\s*"PluginLoad"' "${cs_files[@]}" >/dev/null 2>&1; then
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
			local go_sha
			go_sha="$(plugin_code_sha "$entry")"
			if [[ -n "$go_sha" && -n "$HOST_RUNTIME_ABI_SHA" ]]; then
				go_sha="$(printf '%s\n%s\n' "$go_sha" "$HOST_RUNTIME_ABI_SHA" | hash_text_sha256)"
			fi
			local go_target="$RUNTIME_PLUGIN_DIR/$name.so"
			if [[ -n "$go_sha" ]] && restore_cached_plugin_artifact "go" "$name" "$go_sha" "$go_target"; then
				echo "using cached Go plugin for $PLUGIN_SRC_REL/$name"
				BUILT_GO+=("$go_target")
			else
				echo "building Go plugin from $PLUGIN_SRC_REL/$name"
				run_go go build -buildmode=plugin -o "/runtime/plugins/$name.so" "$PLUGIN_SRC_CONTAINER/$name"
				BUILT_GO+=("$go_target")
				if [[ -n "$go_sha" && -f "$go_target" ]]; then
					store_cached_plugin_artifact "go" "$name" "$go_sha" "$go_target"
				fi
			fi
		fi

		if [[ ${#csproj_files[@]} -gt 0 ]]; then
			ensure_csproj_exports "$entry"
			local cs_sha
			cs_sha="$(plugin_code_sha "$entry")"
			for csproj in "${csproj_files[@]}"; do
				local project_name
				project_name="$(basename "${csproj%.csproj}")"
				local cs_target="$RUNTIME_PLUGIN_DIR/csharp/$project_name/$project_name.so"
				local cache_id="$name/$project_name"
				if [[ -n "$cs_sha" ]] && restore_cached_plugin_artifact "csharp" "$cache_id" "$cs_sha" "$cs_target"; then
					echo "using cached C# plugin for $PLUGIN_SRC_REL/$name/$(basename "$csproj")"
					BUILT_CS+=("$cs_target")
					continue
				fi
				echo "building C# plugin from $PLUGIN_SRC_REL/$name/$(basename "$csproj")"
				publish_csproj "$PLUGIN_SRC_CONTAINER/$name/$project_name.csproj" "$project_name" "$name"
				if [[ -n "$cs_sha" && -f "$cs_target" ]]; then
					store_cached_plugin_artifact "csharp" "$cache_id" "$cs_sha" "$cs_target"
				fi
			done
		elif [[ ${#cs_files[@]} -gt 0 ]]; then
			local cs_sha
			cs_sha="$(plugin_code_sha "$entry")"
			local cs_target="$RUNTIME_PLUGIN_DIR/csharp/$name/$name.so"
			if [[ -n "$cs_sha" ]] && restore_cached_plugin_artifact "csharp" "$name" "$cs_sha" "$cs_target"; then
				echo "using cached C# plugin for $PLUGIN_SRC_REL/$name/*.cs"
				BUILT_CS+=("$cs_target")
			else
				echo "building C# plugin from $PLUGIN_SRC_REL/$name/*.cs"
				generate_and_publish_cs "$entry" "$name"
				if [[ -n "$cs_sha" && -f "$cs_target" ]]; then
					store_cached_plugin_artifact "csharp" "$name" "$cs_sha" "$cs_target"
				fi
			fi
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
HOST_RUNTIME_ABI_SHA="$(host_runtime_abi_sha)"
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

tty_flags=(-i)
if [[ -t 0 && -t 1 ]]; then
	tty_flags+=(-t)
fi

echo "starting server in Docker on UDP $SERVER_PORT"
docker run --rm --init "${tty_flags[@]}" \
	--user "$DOCKER_RUN_USER" \
	--mount "type=bind,src=$ROOT_MOUNT,dst=/workspace" \
	--mount "type=bind,src=$RUNTIME_MOUNT,dst=/runtime" \
	--mount "type=bind,src=$GO_BUILD_CACHE_MOUNT,dst=/cache/go-build" \
	--mount "type=bind,src=$GO_MOD_CACHE_MOUNT,dst=/cache/go-mod" \
	-w /workspace \
	-e CGO_ENABLED=1 \
	-e GOFLAGS=-buildvcs=false \
	-e GOEXPERIMENT="$GOEXPERIMENT_VALUE" \
	-e GOCACHE=/cache/go-build \
	-e GOMODCACHE=/cache/go-mod \
	-e PLUGIN_DIR=/runtime/plugins \
	-e DOTNET_SYSTEM_GLOBALIZATION_INVARIANT="$DOTNET_INVARIANT" \
	-p "$SERVER_PORT:19132/udp" \
	"$PLUGIN_GO_IMAGE" \
	go run ./cmd
