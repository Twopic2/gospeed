
$ProjectName = "gospeed"
$GoVersion = "1.23.6"

function Test-GoInstalled {
    try {
        $goVer = & go version 2>$null
        Write-Host "Go is already installed: $goVer"
        return $true
    }
    catch {
        Write-Host "Go not found. Installing Go $GoVersion..."
        return $false
    }
}

function Install-Go {
    $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    $goUrl = "https://go.dev/dl/go$GoVersion.windows-$arch.zip"
    $tempFile = "$env:TEMP\go$GoVersion.zip"
    $installDir = "$env:USERPROFILE\go-install"
    
    Write-Host "Downloading Go for windows-$arch..."
    try {
        Invoke-WebRequest -Uri $goUrl -OutFile $tempFile -UseBasicParsing
    }
    catch {
        Write-Host "Failed to download Go: $($_.Exception.Message)"
        exit 1
    }
    
    Write-Host "Installing Go to $installDir..."
    if (Test-Path $installDir) {
        Remove-Item -Path $installDir -Recurse -Force
    }
    New-Item -Path $installDir -ItemType Directory -Force | Out-Null
    
    try {
        Expand-Archive -Path $tempFile -DestinationPath $installDir -Force
        Remove-Item $tempFile
        
        $env:PATH = "$installDir\go\bin;$env:PATH"
        
        Write-Host "Go installed successfully"
        Write-Host "Add this to your permanent PATH:"
        Write-Host "$installDir\go\bin"
    }
    catch {
        Write-Host "Failed to install Go: $($_.Exception.Message)"
        exit 1
    }
}

function Build-Project {
    Write-Host "Building $ProjectName..."
    
    if (-not (Test-Path "go.mod")) {
        Write-Host "Error: go.mod not found. Run this script from the project root."
        exit 1
    }
    
    try {
        & go build -o "$ProjectName.exe" .
        Write-Host "Built $ProjectName.exe successfully"
        
        $installDir = "$env:USERPROFILE\.local\bin"
        Write-Host "Installing to $installDir..."
        
        if (-not (Test-Path $installDir)) {
            New-Item -Path $installDir -ItemType Directory -Force | Out-Null
        }
        
        Copy-Item "$ProjectName.exe" "$installDir\" -Force
        
        Write-Host "Installation complete"
        Write-Host "Make sure $installDir is in your PATH"
        Write-Host "Run: `$env:PATH += ';$installDir'"
    }
    catch {
        Write-Host "Build failed: $($_.Exception.Message)"
        exit 1
    }
}

function Main {
    if (-not (Test-GoInstalled)) {
        Install-Go
    }
    
    Build-Project
    
    Write-Host "Done. You can now run: $ProjectName.exe"
}

Main