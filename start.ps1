[CmdletBinding()]
param(
    [switch]$Help,
    [switch]$SelfTest
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Show-Usage {
    @"
Usage: .\start.ps1 (or .\start.bat)

Starts the server inside Docker (Linux runtime), compiling plugins into a temporary directory.
The server/runtime source is copied from a local cached clone of github.com/bedrock-gophers/plugin.

Supported source layouts:
  plugins\<name>\*.go      -> built as Go plugin (.so)
  plugins\<name>\*.csproj  -> published as NativeAOT C# plugin (.so)
  plugins\<name>\*.cs      -> auto-generates csproj + export entrypoint, then publishes NativeAOT C# plugin (.so)
  plugins\*.so             -> prebuilt Go plugin fallback
  plugins\csharp\**\*.so   -> prebuilt C# plugin fallback

Environment overrides:
  PLUGIN_SRC_DIR               Source plugin directory (default: `$PWD\plugins, fallback: cached repo .\plugins)
  HOST_REPO_URL                Host runtime repo URL (default: https://github.com/bedrock-gophers/plugin.git)
  BRANCH                       Optional git branch/tag to clone from HOST_REPO_URL
  COMMIT                       Optional commit SHA to checkout after clone (takes precedence over BRANCH tip)
  HOST_REPO_CACHE_DIR          Cached host repo directory (default: %TMP_BASE%\host-repo-cache)
  HOST_REPO_UPDATE             Set to 1 to manually update HOST_REPO_CACHE_DIR from origin
  PLUGIN_ARTIFACT_CACHE_DIR    Cached plugin artifact directory (default: %TMP_BASE%\plugin-artifact-cache)
  PLUGIN_SHA_CACHE_DIR         Plugin source hash directory (default: %TMP_BASE%\plugin-sha-cache)
  PLUGIN_GO_IMAGE              Go image (default: golang:1.25)
  PLUGIN_DOTNET_IMAGE          Dotnet SDK image (default: mcr.microsoft.com/dotnet/sdk:8.0)
  PLUGIN_DOTNET_AOT_IMAGE_TAG  Docker image tag used for NativeAOT publish (default: bedrock-plugin-dotnet-aot:8.0)
  PLUGIN_CS_RID                C# runtime identifier (default: linux-x64)
  DOTNET_INVARIANT             Set to 0 to disable invariant globalization mode (default: 1)
  SERVER_PORT                  UDP listen port exposed on host (default: 19132)
  KEEP_TEMP                    Set to 1 to keep temporary build/runtime directory
  TMP_BASE                     Base cache/temp directory (default: %LOCALAPPDATA%\bedrock-plugin)

Modes:
  -SelfTest                    Validate local prerequisites and exit
"@
}

if ($Help) {
    Show-Usage
    exit 0
}

function Get-EnvValue {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][AllowEmptyString()][string]$Default
    )
    $raw = [Environment]::GetEnvironmentVariable($Name)
    if ([string]::IsNullOrWhiteSpace($raw)) {
        return $Default
    }
    return $raw
}

function Invoke-Checked {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter(Mandatory = $true)][string[]]$Arguments
    )
    & $FilePath @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "$FilePath failed with exit code $LASTEXITCODE"
    }
}

function Invoke-Captured {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [switch]$IgnoreErrors
    )
    $output = (& $FilePath @Arguments 2>&1 | Out-String)
    $code = $LASTEXITCODE
    if ($code -ne 0 -and -not $IgnoreErrors) {
        $trimmed = $output.Trim()
        if (-not [string]::IsNullOrEmpty($trimmed)) {
            throw $trimmed
        }
        throw "$FilePath failed with exit code $code"
    }
    return @{
        Output   = $output.TrimEnd("`r", "`n")
        ExitCode = $code
    }
}

function To-DockerPath {
    param([Parameter(Mandatory = $true)][string]$Path)
    $full = [System.IO.Path]::GetFullPath($Path)
    return $full -replace '\\', '/'
}

function Ensure-Tool {
    param([Parameter(Mandatory = $true)][string]$Name)
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "error: $Name CLI not found."
    }
}

function Ensure-DotnetAotImage {
    param(
        [Parameter(Mandatory = $true)][string]$ImageTag,
        [Parameter(Mandatory = $true)][string]$BaseImage
    )
    $inspect = Invoke-Captured docker @("image", "inspect", $ImageTag) -IgnoreErrors
    if ($inspect.ExitCode -eq 0) {
        return
    }
    Write-Host "building NativeAOT dotnet image: $ImageTag"
    $dockerfile = @"
FROM $BaseImage
RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends clang zlib1g-dev \
    && rm -rf /var/lib/apt/lists/*
"@
    $dockerfile | docker build -t $ImageTag -
    if ($LASTEXITCODE -ne 0) {
        throw "docker build failed with exit code $LASTEXITCODE"
    }
}

function Resolve-LinuxSo {
    param(
        [Parameter(Mandatory = $true)][string]$OutDir,
        [Parameter(Mandatory = $true)][string]$BaseName
    )
    $direct = Join-Path $OutDir "$BaseName.so"
    if (Test-Path -LiteralPath $direct) {
        return $direct
    }
    $prefixed = Join-Path $OutDir "lib$BaseName.so"
    if (Test-Path -LiteralPath $prefixed) {
        return $prefixed
    }
    $any = Get-ChildItem -LiteralPath $OutDir -Filter "*.so" -File -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($null -ne $any) {
        return $any.FullName
    }
    return ""
}

$pluginSrcDirEnv = [Environment]::GetEnvironmentVariable("PLUGIN_SRC_DIR")
$pluginSrcDirSet = -not [string]::IsNullOrWhiteSpace($pluginSrcDirEnv)
$pluginSrcDir = if ($pluginSrcDirSet) { $pluginSrcDirEnv } else { Join-Path (Get-Location) "plugins" }

$hostRepoUrl = Get-EnvValue -Name "HOST_REPO_URL" -Default "https://github.com/bedrock-gophers/plugin.git"
$branch = Get-EnvValue -Name "BRANCH" -Default ""
$commit = Get-EnvValue -Name "COMMIT" -Default ""
$hostRepoUpdate = Get-EnvValue -Name "HOST_REPO_UPDATE" -Default "0"

$pluginGoImageEnv = [Environment]::GetEnvironmentVariable("PLUGIN_GO_IMAGE")
$pluginGoImageSet = $null -ne $pluginGoImageEnv
$pluginGoImage = if ([string]::IsNullOrWhiteSpace($pluginGoImageEnv)) { "golang:1.25" } else { $pluginGoImageEnv }
$pluginDotnetImage = Get-EnvValue -Name "PLUGIN_DOTNET_IMAGE" -Default "mcr.microsoft.com/dotnet/sdk:8.0"
$pluginDotnetAotImageTag = Get-EnvValue -Name "PLUGIN_DOTNET_AOT_IMAGE_TAG" -Default "bedrock-plugin-dotnet-aot:8.0"
$pluginCsRid = Get-EnvValue -Name "PLUGIN_CS_RID" -Default "linux-x64"
$goExperimentEnv = [Environment]::GetEnvironmentVariable("GOEXPERIMENT")
$goExperimentSet = $null -ne $goExperimentEnv
$goExperimentValue = if ($null -eq $goExperimentEnv) { "" } else { $goExperimentEnv }
$dotnetInvariant = Get-EnvValue -Name "DOTNET_INVARIANT" -Default "1"
$serverPort = Get-EnvValue -Name "SERVER_PORT" -Default "19132"
$keepTemp = Get-EnvValue -Name "KEEP_TEMP" -Default "0"
$tmpBaseDefault = if (-not [string]::IsNullOrWhiteSpace($env:LOCALAPPDATA)) {
    Join-Path $env:LOCALAPPDATA "bedrock-plugin"
} else {
    Join-Path $HOME ".cache/bedrock-plugin"
}
$tmpBase = Get-EnvValue -Name "TMP_BASE" -Default $tmpBaseDefault

$runtimeDir = Join-Path $tmpBase ("start-{0}-{1}-{2}" -f ([DateTimeOffset]::UtcNow.ToUnixTimeSeconds()), $PID, (Get-Random))
$runtimePluginDir = Join-Path $runtimeDir "plugins"
$runtimeBuildDir = Join-Path $runtimeDir "build"
$hostRepoDir = Join-Path $runtimeBuildDir "host-repo"
$stagedPluginSrcRel = ".start-plugin-src"
$stagedPluginSrcDir = Join-Path $hostRepoDir $stagedPluginSrcRel
$goBuildCacheDir = Join-Path $tmpBase "go-build-cache"
$goModCacheDir = Join-Path $tmpBase "go-mod-cache"
$dotnetNugetCacheDir = Join-Path $tmpBase "dotnet-nuget-cache"
$hostRepoCacheDir = Get-EnvValue -Name "HOST_REPO_CACHE_DIR" -Default (Join-Path $tmpBase "host-repo-cache")
$pluginArtifactCacheDir = Get-EnvValue -Name "PLUGIN_ARTIFACT_CACHE_DIR" -Default (Join-Path $tmpBase "plugin-artifact-cache")
$pluginShaCacheDir = Get-EnvValue -Name "PLUGIN_SHA_CACHE_DIR" -Default (Join-Path $tmpBase "plugin-sha-cache")

$builtGo = New-Object System.Collections.Generic.List[string]
$builtCs = New-Object System.Collections.Generic.List[string]
$copiedPrebuilt = New-Object System.Collections.Generic.List[string]
$prebuiltGoCopied = New-Object System.Collections.Generic.List[string]
$builtByName = New-Object "System.Collections.Generic.HashSet[string]" ([System.StringComparer]::OrdinalIgnoreCase)
$sourceGoNames = New-Object "System.Collections.Generic.HashSet[string]" ([System.StringComparer]::OrdinalIgnoreCase)

try {
    if ($SelfTest) {
        Write-Host "running start.ps1 self-test"
        Ensure-Tool git
        Ensure-Tool docker
        $null = Invoke-Captured docker @("info")

        if (-not $pluginCsRid.ToLowerInvariant().StartsWith("linux-")) {
            throw "error: PLUGIN_CS_RID must target linux-* for this Docker Linux runtime (got: $pluginCsRid)"
        }

        New-Item -ItemType Directory -Force -Path $tmpBase, $goBuildCacheDir, $goModCacheDir, $dotnetNugetCacheDir, $hostRepoCacheDir, $pluginArtifactCacheDir, $pluginShaCacheDir | Out-Null

        Write-Host "self-test passed"
        Write-Host "  docker daemon: reachable"
        Write-Host "  host repo cache: $hostRepoCacheDir"
        Write-Host "  plugin artifact cache: $pluginArtifactCacheDir"
        Write-Host "  plugin sha cache: $pluginShaCacheDir"
        Write-Host "  plugin RID: $pluginCsRid"
        exit 0
    }

    Ensure-Tool git
    Ensure-Tool docker
    $null = Invoke-Captured docker @("info")

    if (-not $pluginCsRid.ToLowerInvariant().StartsWith("linux-")) {
        throw "error: PLUGIN_CS_RID must target linux-* for this Docker Linux runtime (got: $pluginCsRid)"
    }

    New-Item -ItemType Directory -Force -Path $runtimePluginDir, (Join-Path $runtimePluginDir "csharp"), $runtimeBuildDir, $hostRepoDir, $hostRepoCacheDir, $goBuildCacheDir, $goModCacheDir, $dotnetNugetCacheDir, $pluginArtifactCacheDir, $pluginShaCacheDir | Out-Null

    if (Test-Path -LiteralPath (Join-Path $hostRepoCacheDir ".git")) {
        $currentOrigin = (Invoke-Captured git @("-C", $hostRepoCacheDir, "remote", "get-url", "origin") -IgnoreErrors).Output.Trim()
        if (-not [string]::IsNullOrWhiteSpace($currentOrigin) -and $currentOrigin -ne $hostRepoUrl) {
            Remove-Item -LiteralPath $hostRepoCacheDir -Recurse -Force
        }
    } elseif ((Test-Path -LiteralPath $hostRepoCacheDir) -and (Get-ChildItem -LiteralPath $hostRepoCacheDir -Force -ErrorAction SilentlyContinue | Select-Object -First 1)) {
        Remove-Item -LiteralPath $hostRepoCacheDir -Recurse -Force
    }

    if (-not (Test-Path -LiteralPath (Join-Path $hostRepoCacheDir ".git"))) {
        $cacheCloneArgs = @("clone", "--quiet", "--depth", "1")
        if (-not [string]::IsNullOrWhiteSpace($branch)) {
            $cacheCloneArgs += @("--branch", $branch)
        }
        $cacheCloneArgs += @($hostRepoUrl, $hostRepoCacheDir)
        Write-Host "initializing host runtime repo cache: $hostRepoCacheDir"
        Invoke-Checked git $cacheCloneArgs
    } elseif ($hostRepoUpdate -eq "1") {
        Write-Host "updating host runtime repo cache from origin"
        if (-not [string]::IsNullOrWhiteSpace($branch)) {
            Invoke-Checked git @("-C", $hostRepoCacheDir, "fetch", "--quiet", "--depth", "1", "origin", $branch)
        } else {
            Invoke-Checked git @("-C", $hostRepoCacheDir, "fetch", "--quiet", "--depth", "1", "origin")
        }
    }

    if (-not [string]::IsNullOrWhiteSpace($branch)) {
        $hasBranch = (Invoke-Captured git @("-C", $hostRepoCacheDir, "show-ref", "--verify", "--quiet", "refs/remotes/origin/$branch") -IgnoreErrors).ExitCode -eq 0
        if (-not $hasBranch -and $hostRepoUpdate -eq "1") {
            Invoke-Checked git @("-C", $hostRepoCacheDir, "fetch", "--quiet", "--depth", "1", "origin", $branch)
            $hasBranch = (Invoke-Captured git @("-C", $hostRepoCacheDir, "show-ref", "--verify", "--quiet", "refs/remotes/origin/$branch") -IgnoreErrors).ExitCode -eq 0
        }
        if (-not $hasBranch) {
            throw "error: branch '$branch' not found in cached host repo. Set HOST_REPO_UPDATE=1 to refresh cache."
        }
    }

    if (-not [string]::IsNullOrWhiteSpace($commit)) {
        $hasCommit = (Invoke-Captured git @("-C", $hostRepoCacheDir, "rev-parse", "--verify", "--quiet", "$commit^{commit}") -IgnoreErrors).ExitCode -eq 0
        if (-not $hasCommit -and $hostRepoUpdate -eq "1") {
            Invoke-Checked git @("-C", $hostRepoCacheDir, "fetch", "--quiet", "--depth", "1", "origin", $commit)
            $hasCommit = (Invoke-Captured git @("-C", $hostRepoCacheDir, "rev-parse", "--verify", "--quiet", "$commit^{commit}") -IgnoreErrors).ExitCode -eq 0
        }
        if (-not $hasCommit) {
            throw "error: commit '$commit' not found in cached host repo. Set HOST_REPO_UPDATE=1 to refresh cache."
        }
    }

    if (Test-Path -LiteralPath $hostRepoDir) {
        Remove-Item -LiteralPath $hostRepoDir -Recurse -Force
    }
    Write-Host "copying host runtime repo from cache"
    Invoke-Checked git @("clone", "--quiet", "--no-hardlinks", $hostRepoCacheDir, $hostRepoDir)

    if (-not [string]::IsNullOrWhiteSpace($commit)) {
        Invoke-Checked git @("-C", $hostRepoDir, "checkout", "--quiet", "--detach", $commit)
    } elseif (-not [string]::IsNullOrWhiteSpace($branch)) {
        Invoke-Checked git @("-C", $hostRepoDir, "checkout", "--quiet", "--detach", "origin/$branch")
    }

    if (-not (Test-Path -LiteralPath (Join-Path $hostRepoDir "cmd/main.go"))) {
        throw "error: cloned host repo is missing cmd/main.go"
    }

    $resolvedRef = (Invoke-Captured git @("-C", $hostRepoDir, "rev-parse", "--short=12", "HEAD")).Output.Trim()
    Write-Host "using host runtime ref: $resolvedRef"

    if (-not (Test-Path -LiteralPath $pluginSrcDir)) {
        if (-not $pluginSrcDirSet -and (Test-Path -LiteralPath (Join-Path $hostRepoDir "plugins"))) {
            $pluginSrcDir = Join-Path $hostRepoDir "plugins"
        } else {
            throw "error: plugin source directory not found: $pluginSrcDir"
        }
    }

    if (Test-Path -LiteralPath $stagedPluginSrcDir) {
        Remove-Item -LiteralPath $stagedPluginSrcDir -Recurse -Force
    }
    New-Item -ItemType Directory -Force -Path $stagedPluginSrcDir | Out-Null
    Get-ChildItem -LiteralPath $pluginSrcDir -Force | Copy-Item -Destination $stagedPluginSrcDir -Recurse -Force

    $pluginSrcDir = $stagedPluginSrcDir
    $pluginSrcRel = $stagedPluginSrcRel
    $pluginSrcContainer = "/workspace/$pluginSrcRel"

    $rootMount = To-DockerPath $hostRepoDir
    $runtimeMount = To-DockerPath $runtimeDir
    $goBuildCacheMount = To-DockerPath $goBuildCacheDir
    $goModCacheMount = To-DockerPath $goModCacheDir
    $dotnetNugetCacheMount = To-DockerPath $dotnetNugetCacheDir

    function Run-Go {
        param([Parameter(Mandatory = $true)][string[]]$InnerArgs)
        $args = @(
            "run", "--rm",
            "--mount", "type=bind,src=$rootMount,dst=/workspace",
            "--mount", "type=bind,src=$runtimeMount,dst=/runtime",
            "--mount", "type=bind,src=$goBuildCacheMount,dst=/cache/go-build",
            "--mount", "type=bind,src=$goModCacheMount,dst=/cache/go-mod",
            "-w", "/workspace",
            "-e", "CGO_ENABLED=1",
            "-e", "GOFLAGS=-buildvcs=false",
            "-e", "GOEXPERIMENT=$goExperimentValue",
            "-e", "GOCACHE=/cache/go-build",
            "-e", "GOMODCACHE=/cache/go-mod",
            $pluginGoImage
        ) + $InnerArgs
        Invoke-Checked docker $args
    }

    function Run-GoCaptured {
        param(
            [Parameter(Mandatory = $true)][string[]]$InnerArgs,
            [switch]$IgnoreErrors
        )
        $args = @(
            "run", "--rm",
            "--mount", "type=bind,src=$rootMount,dst=/workspace",
            "--mount", "type=bind,src=$runtimeMount,dst=/runtime",
            "--mount", "type=bind,src=$goBuildCacheMount,dst=/cache/go-build",
            "--mount", "type=bind,src=$goModCacheMount,dst=/cache/go-mod",
            "-w", "/workspace",
            "-e", "CGO_ENABLED=1",
            "-e", "GOFLAGS=-buildvcs=false",
            "-e", "GOEXPERIMENT=$goExperimentValue",
            "-e", "GOCACHE=/cache/go-build",
            "-e", "GOMODCACHE=/cache/go-mod",
            $pluginGoImage
        ) + $InnerArgs
        return Invoke-Captured docker $args -IgnoreErrors:$IgnoreErrors
    }

    function Run-DotNet {
        param([Parameter(Mandatory = $true)][string[]]$InnerArgs)
        Ensure-DotnetAotImage -ImageTag $pluginDotnetAotImageTag -BaseImage $pluginDotnetImage
        $args = @(
            "run", "--rm",
            "--mount", "type=bind,src=$rootMount,dst=/workspace",
            "--mount", "type=bind,src=$runtimeMount,dst=/runtime",
            "--mount", "type=bind,src=$dotnetNugetCacheMount,dst=/cache/nuget",
            "-w", "/workspace",
            "-e", "HOME=/tmp",
            "-e", "DOTNET_CLI_HOME=/tmp",
            "-e", "DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1",
            "-e", "DOTNET_CLI_TELEMETRY_OPTOUT=1",
            "-e", "NUGET_PACKAGES=/cache/nuget",
            $pluginDotnetAotImageTag
        ) + $InnerArgs
        Invoke-Checked docker $args
    }

    function Collect-SourceGoPluginNames {
        $entries = Get-ChildItem -LiteralPath $pluginSrcDir -Directory -ErrorAction SilentlyContinue
        foreach ($entry in $entries) {
            $goFiles = Get-ChildItem -LiteralPath $entry.FullName -Filter "*.go" -File -ErrorAction SilentlyContinue
            if ($goFiles.Count -gt 0) {
                $null = $sourceGoNames.Add($entry.Name)
            }
        }
    }

    function Detect-PrebuiltGoToolchain {
        $rootPrebuilt = Get-ChildItem -LiteralPath $pluginSrcDir -Filter "*.so" -File -ErrorAction SilentlyContinue
        if ($rootPrebuilt.Count -eq 0) {
            return
        }

        $detectedVersion = ""
        $detectedExperiment = ""
        $scanned = 0

        foreach ($so in $rootPrebuilt) {
            $name = [System.IO.Path]::GetFileNameWithoutExtension($so.Name)
            if ($sourceGoNames.Contains($name)) {
                continue
            }

            $soContainer = "$pluginSrcContainer/$($so.Name)"
            $meta = Run-GoCaptured -InnerArgs @("go", "version", "-m", $soContainer) -IgnoreErrors
            if ($meta.ExitCode -ne 0) {
                throw "error: failed to inspect prebuilt Go plugin metadata: $($so.FullName)`n$($meta.Output)"
            }

            $lines = $meta.Output -split "`r?`n"
            $firstLine = if ($lines.Count -gt 0) { $lines[0] } else { "" }
            if ($firstLine -notmatch '(go[0-9][^\s]*)') {
                throw "error: could not read Go toolchain version from prebuilt plugin: $($so.FullName)"
            }
            $version = $Matches[1]

            if ([string]::IsNullOrWhiteSpace($detectedVersion)) {
                $detectedVersion = $version
            } elseif ($detectedVersion -ne $version) {
                throw "error: prebuilt Go plugins use different Go versions ($detectedVersion vs $version).`n       Rebuild them with one toolchain, or keep only one toolchain's artifacts."
            }

            $experiment = ""
            foreach ($line in $lines) {
                if ($line -match '^\s*build\s+GOEXPERIMENT=(.*)$') {
                    $experiment = $Matches[1].Trim()
                    break
                }
            }

            if (-not [string]::IsNullOrWhiteSpace($experiment)) {
                if ([string]::IsNullOrWhiteSpace($detectedExperiment)) {
                    $detectedExperiment = $experiment
                } elseif ($detectedExperiment -ne $experiment) {
                    throw "error: prebuilt Go plugins use different GOEXPERIMENT values ($detectedExperiment vs $experiment)."
                }
            }

            $scanned++
        }

        if ($scanned -eq 0) {
            return
        }

        if (-not $pluginGoImageSet) {
            $pluginGoImage = "golang:$($detectedVersion.Substring(2))"
        }
        if (-not $goExperimentSet -and -not [string]::IsNullOrWhiteSpace($detectedExperiment)) {
            $goExperimentValue = $detectedExperiment
        }

        $toolchainMessage = "using prebuilt Go plugin toolchain: $detectedVersion"
        if (-not [string]::IsNullOrWhiteSpace($detectedExperiment)) {
            $toolchainMessage += " (GOEXPERIMENT=$detectedExperiment)"
        }
        Write-Host $toolchainMessage
        Write-Host "using Go image: $pluginGoImage"
    }

    function Get-CodeSourceSha {
        param([Parameter(Mandatory = $true)][string]$PluginDir)
        $codeFiles = Get-ChildItem -LiteralPath $PluginDir -File -ErrorAction SilentlyContinue | Where-Object { $_.Extension -in ".go", ".cs" } | Sort-Object Name
        if ($codeFiles.Count -eq 0) {
            return ""
        }

        $lines = New-Object System.Collections.Generic.List[string]
        foreach ($file in $codeFiles) {
            $fileHash = (Get-FileHash -LiteralPath $file.FullName -Algorithm SHA256).Hash.ToLowerInvariant()
            $lines.Add("$($file.Name):$fileHash")
        }
        $payload = [System.Text.Encoding]::UTF8.GetBytes(($lines -join "`n"))
        $sha = [System.Security.Cryptography.SHA256]::Create()
        try {
            $hashBytes = $sha.ComputeHash($payload)
        }
        finally {
            $sha.Dispose()
        }
        return ([BitConverter]::ToString($hashBytes)).Replace("-", "").ToLowerInvariant()
    }

    function Get-CacheSafeId {
        param([Parameter(Mandatory = $true)][string]$Id)
        return ($Id -replace '[\\/:*?"<>| ]', "__")
    }

    function Get-PluginShaFilePath {
        param(
            [Parameter(Mandatory = $true)][string]$Kind,
            [Parameter(Mandatory = $true)][string]$Id
        )
        $safeId = Get-CacheSafeId -Id $Id
        $kindDir = Join-Path $pluginShaCacheDir $Kind
        New-Item -ItemType Directory -Force -Path $kindDir | Out-Null
        return (Join-Path $kindDir "$safeId.sha")
    }

    function Get-PluginArtifactCachePath {
        param(
            [Parameter(Mandatory = $true)][string]$Kind,
            [Parameter(Mandatory = $true)][string]$Id,
            [Parameter(Mandatory = $true)][string]$Sha
        )
        $safeId = Get-CacheSafeId -Id $Id
        $artifactDir = Join-Path (Join-Path $pluginArtifactCacheDir $Kind) $safeId
        New-Item -ItemType Directory -Force -Path $artifactDir | Out-Null
        return (Join-Path $artifactDir "$Sha.so")
    }

    function Restore-CachedPluginArtifact {
        param(
            [Parameter(Mandatory = $true)][string]$Kind,
            [Parameter(Mandatory = $true)][string]$Id,
            [Parameter(Mandatory = $true)][string]$Sha,
            [Parameter(Mandatory = $true)][string]$Destination
        )
        $shaFile = Get-PluginShaFilePath -Kind $Kind -Id $Id
        if (-not (Test-Path -LiteralPath $shaFile)) {
            return $false
        }
        $cachedSha = (Get-Content -LiteralPath $shaFile -Raw -ErrorAction SilentlyContinue).Trim()
        if ([string]::IsNullOrWhiteSpace($cachedSha) -or $cachedSha -ne $Sha) {
            return $false
        }
        $artifactPath = Get-PluginArtifactCachePath -Kind $Kind -Id $Id -Sha $Sha
        if (-not (Test-Path -LiteralPath $artifactPath)) {
            return $false
        }
        New-Item -ItemType Directory -Force -Path (Split-Path -Parent $Destination) | Out-Null
        Copy-Item -LiteralPath $artifactPath -Destination $Destination -Force
        return $true
    }

    function Store-CachedPluginArtifact {
        param(
            [Parameter(Mandatory = $true)][string]$Kind,
            [Parameter(Mandatory = $true)][string]$Id,
            [Parameter(Mandatory = $true)][string]$Sha,
            [Parameter(Mandatory = $true)][string]$Source
        )
        $shaFile = Get-PluginShaFilePath -Kind $Kind -Id $Id
        $artifactPath = Get-PluginArtifactCachePath -Kind $Kind -Id $Id -Sha $Sha
        New-Item -ItemType Directory -Force -Path (Split-Path -Parent $artifactPath) | Out-Null
        Copy-Item -LiteralPath $Source -Destination $artifactPath -Force
        Set-Content -LiteralPath $shaFile -Value $Sha -NoNewline
    }

    function Publish-Csproj {
        param(
            [Parameter(Mandatory = $true)][string]$CsprojContainerPath,
            [Parameter(Mandatory = $true)][string]$ProjectName,
            [Parameter(Mandatory = $true)][string]$PluginGroup
        )
        $outDir = Join-Path $runtimeBuildDir "csharp/$PluginGroup/$ProjectName"
        $outContainer = "/runtime/build/csharp/$PluginGroup/$ProjectName"
        New-Item -ItemType Directory -Force -Path $outDir | Out-Null

        Run-DotNet @(
            "dotnet", "publish", $CsprojContainerPath,
            "-c", "Release",
            "-o", $outContainer,
            "-r", $pluginCsRid,
            "/p:PublishAot=true",
            "/p:NativeLib=Shared",
            "/p:SelfContained=true",
            "--nologo"
        )

        $artifact = Resolve-LinuxSo -OutDir $outDir -BaseName $ProjectName
        if ([string]::IsNullOrWhiteSpace($artifact)) {
            throw "error: failed to find C# plugin artifact (.so) for $ProjectName"
        }
        $targetDir = Join-Path $runtimePluginDir "csharp/$ProjectName"
        New-Item -ItemType Directory -Force -Path $targetDir | Out-Null
        $target = Join-Path $targetDir "$ProjectName.so"
        Copy-Item -LiteralPath $artifact -Destination $target -Force
        $builtCs.Add($target) | Out-Null
    }

    function Generate-AndPublish-Cs {
        param(
            [Parameter(Mandatory = $true)][string]$PluginDir,
            [Parameter(Mandatory = $true)][string]$PluginName
        )
        $genDir = Join-Path $runtimeBuildDir "generated-cs/$PluginName"
        $outDir = Join-Path $runtimeBuildDir "generated-cs-out/$PluginName"
        New-Item -ItemType Directory -Force -Path $genDir, $outDir | Out-Null

        $csFiles = Get-ChildItem -LiteralPath $PluginDir -Filter "*.cs" -File -ErrorAction SilentlyContinue
        if ($csFiles.Count -eq 0) {
            return
        }
        foreach ($file in $csFiles) {
            Copy-Item -LiteralPath $file.FullName -Destination (Join-Path $genDir $file.Name) -Force
        }

        @'
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
'@ | Set-Content -LiteralPath (Join-Path $genDir "__plugin_exports.generated.cs") -NoNewline

        @'
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
'@ | Set-Content -LiteralPath (Join-Path $genDir "$PluginName.csproj") -NoNewline

        Publish-Csproj -CsprojContainerPath "/runtime/build/generated-cs/$PluginName/$PluginName.csproj" -ProjectName $PluginName -PluginGroup $PluginName
    }

    function Validate-PrebuiltGoArtifacts {
        if ($prebuiltGoCopied.Count -eq 0) {
            return
        }

        @'
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
'@ | Set-Content -LiteralPath (Join-Path $runtimeBuildDir "check_prebuilt_go_plugin.go")

        $kept = New-Object System.Collections.Generic.List[string]
        foreach ($hostSo in @($prebuiltGoCopied)) {
            $runtimeSo = "/runtime/plugins/$([System.IO.Path]::GetFileName($hostSo))"
            $check = Run-GoCaptured -InnerArgs @("go", "run", "/runtime/build/check_prebuilt_go_plugin.go", $runtimeSo) -IgnoreErrors
            if ($check.ExitCode -eq 0) {
                $kept.Add($hostSo) | Out-Null
                continue
            }
            [Console]::Error.WriteLine("warning: skipping incompatible prebuilt Go plugin: $hostSo")
            if (-not [string]::IsNullOrWhiteSpace($check.Output)) {
                [Console]::Error.WriteLine($check.Output)
            }
            Remove-Item -LiteralPath $hostSo -Force -ErrorAction SilentlyContinue
        }

        $prebuiltGoCopied.Clear()
        foreach ($artifact in $kept) {
            $prebuiltGoCopied.Add($artifact) | Out-Null
        }

        $filtered = New-Object System.Collections.Generic.List[string]
        foreach ($artifact in @($copiedPrebuilt)) {
            if (Test-Path -LiteralPath $artifact) {
                $filtered.Add($artifact) | Out-Null
            }
        }
        $copiedPrebuilt.Clear()
        foreach ($artifact in $filtered) {
            $copiedPrebuilt.Add($artifact) | Out-Null
        }
    }

    Collect-SourceGoPluginNames
    Detect-PrebuiltGoToolchain

    $entries = Get-ChildItem -LiteralPath $pluginSrcDir -Directory -ErrorAction SilentlyContinue
    foreach ($entry in $entries) {
        $name = $entry.Name
        $goFiles = Get-ChildItem -LiteralPath $entry.FullName -Filter "*.go" -File -ErrorAction SilentlyContinue
        $csprojFiles = Get-ChildItem -LiteralPath $entry.FullName -Filter "*.csproj" -File -ErrorAction SilentlyContinue
        $csFiles = Get-ChildItem -LiteralPath $entry.FullName -Filter "*.cs" -File -ErrorAction SilentlyContinue

        if ($goFiles.Count -gt 0) {
            $goSha = Get-CodeSourceSha -PluginDir $entry.FullName
            $goTarget = Join-Path $runtimePluginDir "$name.so"
            if (-not [string]::IsNullOrWhiteSpace($goSha) -and (Restore-CachedPluginArtifact -Kind "go" -Id $name -Sha $goSha -Destination $goTarget)) {
                Write-Host "using cached Go plugin for $pluginSrcRel/$name"
                $builtGo.Add($goTarget) | Out-Null
            } else {
                Write-Host "building Go plugin from $pluginSrcRel/$name"
                Run-Go @("go", "build", "-buildmode=plugin", "-o", "/runtime/plugins/$name.so", "$pluginSrcContainer/$name")
                $builtGo.Add($goTarget) | Out-Null
                if (-not [string]::IsNullOrWhiteSpace($goSha) -and (Test-Path -LiteralPath $goTarget)) {
                    Store-CachedPluginArtifact -Kind "go" -Id $name -Sha $goSha -Source $goTarget
                }
            }
        }

        if ($csprojFiles.Count -gt 0) {
            $csSha = Get-CodeSourceSha -PluginDir $entry.FullName
            foreach ($csproj in $csprojFiles) {
                $projectName = [System.IO.Path]::GetFileNameWithoutExtension($csproj.Name)
                $csTarget = Join-Path (Join-Path $runtimePluginDir "csharp/$projectName") "$projectName.so"
                $cacheId = "$name/$projectName"
                if (-not [string]::IsNullOrWhiteSpace($csSha) -and (Restore-CachedPluginArtifact -Kind "csharp" -Id $cacheId -Sha $csSha -Destination $csTarget)) {
                    Write-Host "using cached C# plugin for $pluginSrcRel/$name/$($csproj.Name)"
                    $builtCs.Add($csTarget) | Out-Null
                    continue
                }
                Write-Host "building C# plugin from $pluginSrcRel/$name/$($csproj.Name)"
                Publish-Csproj -CsprojContainerPath "$pluginSrcContainer/$name/$projectName.csproj" -ProjectName $projectName -PluginGroup $name
                if (-not [string]::IsNullOrWhiteSpace($csSha) -and (Test-Path -LiteralPath $csTarget)) {
                    Store-CachedPluginArtifact -Kind "csharp" -Id $cacheId -Sha $csSha -Source $csTarget
                }
            }
        } elseif ($csFiles.Count -gt 0) {
            $csSha = Get-CodeSourceSha -PluginDir $entry.FullName
            $csTarget = Join-Path (Join-Path $runtimePluginDir "csharp/$name") "$name.so"
            if (-not [string]::IsNullOrWhiteSpace($csSha) -and (Restore-CachedPluginArtifact -Kind "csharp" -Id $name -Sha $csSha -Destination $csTarget)) {
                Write-Host "using cached C# plugin for $pluginSrcRel/$name/*.cs"
                $builtCs.Add($csTarget) | Out-Null
            } else {
                Write-Host "building C# plugin from $pluginSrcRel/$name/*.cs"
                Generate-AndPublish-Cs -PluginDir $entry.FullName -PluginName $name
                if (-not [string]::IsNullOrWhiteSpace($csSha) -and (Test-Path -LiteralPath $csTarget)) {
                    Store-CachedPluginArtifact -Kind "csharp" -Id $name -Sha $csSha -Source $csTarget
                }
            }
        }
    }

    foreach ($artifact in $builtGo) {
        $null = $builtByName.Add([System.IO.Path]::GetFileNameWithoutExtension($artifact))
    }
    foreach ($artifact in $builtCs) {
        $null = $builtByName.Add([System.IO.Path]::GetFileNameWithoutExtension($artifact))
    }

    $rootPrebuilt = Get-ChildItem -LiteralPath $pluginSrcDir -Filter "*.so" -File -ErrorAction SilentlyContinue
    foreach ($so in $rootPrebuilt) {
        $name = [System.IO.Path]::GetFileNameWithoutExtension($so.Name)
        if ($builtByName.Contains($name)) {
            continue
        }
        $dst = Join-Path $runtimePluginDir $so.Name
        Copy-Item -LiteralPath $so.FullName -Destination $dst -Force
        $copiedPrebuilt.Add($dst) | Out-Null
        $prebuiltGoCopied.Add($dst) | Out-Null
    }

    $csharpRoot = Join-Path $pluginSrcDir "csharp"
    if (Test-Path -LiteralPath $csharpRoot) {
        $csPrebuilt = Get-ChildItem -LiteralPath $csharpRoot -Filter "*.so" -File -Recurse -ErrorAction SilentlyContinue
        foreach ($so in $csPrebuilt) {
            $name = [System.IO.Path]::GetFileNameWithoutExtension($so.Name)
            if ($builtByName.Contains($name)) {
                continue
            }
            $rel = $so.FullName.Substring($pluginSrcDir.Length).TrimStart('\', '/')
            $dst = Join-Path $runtimePluginDir $rel
            New-Item -ItemType Directory -Force -Path (Split-Path -Parent $dst) | Out-Null
            Copy-Item -LiteralPath $so.FullName -Destination $dst -Force
            $copiedPrebuilt.Add($dst) | Out-Null
        }
    }

    Validate-PrebuiltGoArtifacts

    Write-Host "temporary plugin dir: $runtimePluginDir"
    Write-Host "resolved plugin artifacts:"
    if ($builtGo.Count -gt 0) { $builtGo | ForEach-Object { Write-Host "  $_" } }
    if ($builtCs.Count -gt 0) { $builtCs | ForEach-Object { Write-Host "  $_" } }
    if ($copiedPrebuilt.Count -gt 0) { $copiedPrebuilt | ForEach-Object { Write-Host "  $_" } }
    if ($builtGo.Count -eq 0 -and $builtCs.Count -eq 0 -and $copiedPrebuilt.Count -eq 0) {
        Write-Host "  none"
    }

    $ttyFlags = @("-i")
    try {
        if (-not [Console]::IsInputRedirected -and -not [Console]::IsOutputRedirected) {
            $ttyFlags += @("-t")
        }
    } catch {
        $ttyFlags = @("-i")
    }

    Write-Host "starting server in Docker on UDP $serverPort"
    $runArgs = @(
        "run", "--rm", "--init"
    ) + $ttyFlags + @(
        "--mount", "type=bind,src=$rootMount,dst=/workspace",
        "--mount", "type=bind,src=$runtimeMount,dst=/runtime",
        "--mount", "type=bind,src=$goBuildCacheMount,dst=/cache/go-build",
        "--mount", "type=bind,src=$goModCacheMount,dst=/cache/go-mod",
        "-w", "/workspace",
        "-e", "CGO_ENABLED=1",
        "-e", "GOFLAGS=-buildvcs=false",
        "-e", "GOEXPERIMENT=$goExperimentValue",
        "-e", "GOCACHE=/cache/go-build",
        "-e", "GOMODCACHE=/cache/go-mod",
        "-e", "PLUGIN_DIR=/runtime/plugins",
        "-e", "DOTNET_SYSTEM_GLOBALIZATION_INVARIANT=$dotnetInvariant",
        "-p", "$serverPort`:19132/udp",
        $pluginGoImage,
        "go", "run", "./cmd"
    )
    Invoke-Checked docker $runArgs
}
finally {
    if ($keepTemp -eq "1") {
        Write-Host "temporary runtime directory kept at: $runtimeDir"
    } else {
        if (Test-Path -LiteralPath $runtimeDir) {
            Remove-Item -LiteralPath $runtimeDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}
