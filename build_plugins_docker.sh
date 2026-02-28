#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="$ROOT_DIR/plugins"
PLUGINS_DIR="$ROOT_DIR/examples/plugins"
CS_OUT_DIR="$OUT_DIR/csharp"
PLUGIN_GO_IMAGE="${PLUGIN_GO_IMAGE:-golang:1.25.7}"
PLUGIN_DOTNET_IMAGE="${PLUGIN_DOTNET_IMAGE:-mcr.microsoft.com/dotnet/sdk:8.0}"
PLUGIN_CS_RID="${PLUGIN_CS_RID:-linux-x64}"
PLUGIN_DOTNET_AOT_IMAGE_TAG="${PLUGIN_DOTNET_AOT_IMAGE_TAG:-bedrock-plugin-dotnet-aot:8.0}"
TMP_BASE="${TMP_BASE:-$ROOT_DIR/.tmp}"
GO_BUILD_CACHE_DIR="$TMP_BASE/go-build-cache"
GO_MOD_CACHE_DIR="$TMP_BASE/go-mod-cache"
DOTNET_NUGET_CACHE_DIR="$TMP_BASE/dotnet-nuget-cache"

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

mkdir -p "$OUT_DIR"
rm -rf "$CS_OUT_DIR"
mkdir -p "$CS_OUT_DIR"
mkdir -p "$GO_BUILD_CACHE_DIR" "$GO_MOD_CACHE_DIR" "$DOTNET_NUGET_CACHE_DIR"

shopt -s nullglob
plugin_dirs=("$PLUGINS_DIR"/*)
if [ ${#plugin_dirs[@]} -eq 0 ]; then
	echo "no plugins found in $PLUGINS_DIR"
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
GO_BUILD_CACHE_MOUNT="$(host_path_for_docker "$GO_BUILD_CACHE_DIR")"
GO_MOD_CACHE_MOUNT="$(host_path_for_docker "$GO_MOD_CACHE_DIR")"
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

run_go() {
	docker run --rm \
		--mount "type=bind,src=$ROOT_MOUNT,dst=/workspace" \
		--mount "type=bind,src=$GO_BUILD_CACHE_MOUNT,dst=/root/.cache/go-build" \
		--mount "type=bind,src=$GO_MOD_CACHE_MOUNT,dst=/go/pkg/mod" \
		-w /workspace \
		-e CGO_ENABLED=1 \
		-e GOFLAGS=-buildvcs=false \
		-e GOEXPERIMENT="${GOEXPERIMENT:-}" \
		"$PLUGIN_GO_IMAGE" \
		"$@"
}

run_dotnet_publish() {
	local csproj="$1"
	local out_dir="$2"
	ensure_dotnet_aot_image
	docker run --rm \
		--mount "type=bind,src=$ROOT_MOUNT,dst=/workspace" \
		--mount "type=bind,src=$DOTNET_NUGET_CACHE_MOUNT,dst=/root/.nuget/packages" \
		-w /workspace \
		"$PLUGIN_DOTNET_AOT_IMAGE_TAG" \
		dotnet publish "$csproj" -c Release -o "$out_dir" -r "$PLUGIN_CS_RID" /p:PublishAot=true /p:NativeLib=Shared /p:SelfContained=true --nologo >/dev/null
}

built_go=()
built_cs=()

for dir in "${plugin_dirs[@]}"; do
	[ -d "$dir" ] || continue
	name="$(basename "$dir")"

	shopt -s nullglob
	go_files=("$dir"/*.go)
	csproj_files=("$dir"/*.csproj)
	shopt -u nullglob

	if [ ${#go_files[@]} -gt 0 ]; then
		run_go go build -buildmode=plugin -o "/workspace/plugins/$name.so" "./examples/plugins/$name"
		built_go+=("$OUT_DIR/$name.so")
	fi

	if [ ${#csproj_files[@]} -gt 0 ]; then
		for csproj in "${csproj_files[@]}"; do
			proj_name="$(basename "${csproj%.csproj}")"
			out="$CS_OUT_DIR/$proj_name"
			mkdir -p "$out"
			run_dotnet_publish "/workspace/examples/plugins/$name/$proj_name.csproj" "/workspace/plugins/csharp/$proj_name"

			artifact=""
			if [ -f "$out/$proj_name.so" ]; then
				artifact="$out/$proj_name.so"
			elif [ -f "$out/lib$proj_name.so" ]; then
				artifact="$out/lib$proj_name.so"
			else
				shopt -s nullglob
				candidates=("$out"/*.so)
				shopt -u nullglob
				if [ ${#candidates[@]} -gt 0 ]; then
					artifact="${candidates[0]}"
				fi
			fi
			if [ -z "$artifact" ]; then
				echo "failed to find native C# plugin artifact (.so) for $proj_name" >&2
				exit 1
			fi

			target="$out/$proj_name.so"
			if [ "$artifact" != "$target" ]; then
				cp "$artifact" "$target"
			fi
			built_cs+=("$target")
		done
	fi
done

echo "built plugins:"
if [ ${#built_go[@]} -gt 0 ]; then
	printf '%s\n' "${built_go[@]}"
fi
if [ ${#built_cs[@]} -gt 0 ]; then
	printf '%s\n' "${built_cs[@]}"
fi
if [ ${#built_go[@]} -eq 0 ] && [ ${#built_cs[@]} -eq 0 ]; then
	echo "none"
fi
