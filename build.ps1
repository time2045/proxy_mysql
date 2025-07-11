# build.ps1

# Clean and create the output directory
$outputDir = "builds"
if (Test-Path $outputDir) {
    Remove-Item $outputDir -Recurse
}
New-Item -ItemType Directory -Path $outputDir | Out-Null

# Define source files and their base output names
$targets = @{
    "server_proxy/server_proxy.go" = "server_proxy"
    "local_client/local_client.go" = "local_client"
}

# Define target platforms (OS/Architecture)
$platforms = @(
    @{os="windows"; arch="amd64"},
    @{os="linux";   arch="amd64"},
    @{os="linux";   arch="arm64"},
    @{os="darwin";  arch="amd64"},  # For macOS (Intel)
    @{os="darwin";  arch="arm64"}   # For macOS (Apple Silicon M1/M2)
)

# Disable CGO
$env:CGO_ENABLED = 0

# Iterate over all source files
foreach ($item in $targets.GetEnumerator()) {
    $sourceFile = $item.Key
    $outputName = $item.Value
    Write-Host "--- Building $($outputName) ---" -ForegroundColor Green

    # Iterate over all platforms
    foreach ($p in $platforms) {
        $goos = $p.os
        $goarch = $p.arch

        Write-Host "Building for $goos/$goarch..." -ForegroundColor Yellow

        # Set environment variables
        $env:GOOS = $goos
        $env:GOARCH = $goarch

        # Construct the output filename
        $fileName = "$($outputName)_$($goos)_$($goarch)"
        if ($goos -eq "windows") {
            $fileName += ".exe"
        }

        # Construct the full output path
        $outputPath = Join-Path $outputDir $fileName

        # Execute the build command
        go build -ldflags "-s -w" -o $outputPath $sourceFile

        if ($?) {
            Write-Host "Successfully built $outputPath" -ForegroundColor Cyan
        } else {
            Write-Host "Failed to build for $goos/$goarch" -ForegroundColor Red
        }
    }
}

Write-Host "--- All builds completed! ---" -ForegroundColor Green