# PolyAgent Installer Script for Windows
# PowerShell script

param(
    [string]$Version,
    [string]$InstallDir = "$env:LOCALAPPDATA\PolyAgent"
)

$ErrorActionPreference = "Stop"

# Configuration
$Repo = "Zacy-Sokach/PolyAgent"
$ReleaseUrl = "https://github.com/$Repo/releases/download/$Version"

# Colors for output
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

# Get latest version from GitHub API
function Get-LatestVersion {
    $apiUrl = "https://api.github.com/repos/$Repo/releases/latest"
    Write-ColorOutput "Fetching latest version from GitHub..." Yellow
    
    try {
        $response = Invoke-RestMethod -Uri $apiUrl -TimeoutSec 10
        $script:Version = $response.tag_name
        Write-ColorOutput "Latest version detected: $Version" Green
        return $Version
    }
    catch {
        Write-ColorOutput "Failed to fetch latest version: $_" Red
        Write-ColorOutput "Please check your network connection or specify -Version parameter." Yellow
        exit 1
    }
}

# Detect Windows architecture
function Get-WindowsArchitecture {
    try {
        $arch = (Get-CimInstance Win32_Processor).Architecture
        switch ($arch) {
            0 { return "amd64" }      # x86
            9 { return "amd64" }      # x64
            12 { return "arm64" }     # ARM64
            default { 
                Write-ColorOutput "Unsupported architecture detected: $arch" Red
                exit 1 
            }
        }
    }
    catch {
        Write-ColorOutput "Failed to detect architecture: $_" Red
        exit 1
    }
}

# Create installation directory
function New-InstallationDirectory {
    param([string]$Path)
    
    if (!(Test-Path $Path)) {
        Write-ColorOutput "Creating installation directory: $Path" Yellow
        New-Item -ItemType Directory -Path $Path -Force | Out-Null
    }
}

# Download file with retry
function Invoke-DownloadWithRetry {
    param(
        [string]$Url,
        [string]$OutputPath,
        [int]$MaxAttempts = 3
    )
    
    $attempt = 1
    while ($attempt -le $MaxAttempts) {
        Write-ColorOutput "Download attempt $attempt/$MaxAttempts..." Yellow
        
        try {
            $webClient = New-Object System.Net.WebClient
            $webClient.DownloadFile($Url, $OutputPath)
            Write-ColorOutput "Download completed successfully!" Green
            return $true
        }
        catch {
            Write-ColorOutput "Download failed: $_" Red
            $attempt++
            if ($attempt -le $MaxAttempts) {
                Write-ColorOutput "Retrying in 2 seconds..." Yellow
                Start-Sleep -Seconds 2
            }
        }
    }
    
    return $false
}

# Download PolyAgent binary
function Invoke-PolyAgentDownload {
    param(
        [string]$Url,
        [string]$OutputPath
    )
    
    Write-ColorOutput "Downloading PolyAgent from:" Yellow
    Write-ColorOutput $Url Yellow
    
    if (-not (Invoke-DownloadWithRetry -Url $Url -OutputPath $OutputPath)) {
        Write-ColorOutput "Failed to download PolyAgent after multiple attempts." Red
        exit 1
    }
}

# Add directory to PATH
function Add-ToPath {
    param([string]$Path)
    
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    
    if ($currentPath -notlike "*$Path*") {
        Write-ColorOutput "Adding $Path to user PATH..." Yellow
        
        $newPath = "$currentPath;$Path"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        
        # Also update current session so user can run polyagent immediately
        $env:PATH = "$env:PATH;$Path"
        
        Write-ColorOutput "PATH updated successfully!" Green
        Write-ColorOutput "You can now run 'polyagent' from any terminal window." Green
    }
}

# Verify installation
function Test-PolyAgentInstallation {
    if (Get-Command polyagent -ErrorAction SilentlyContinue) {
        Write-ColorOutput "PolyAgent installed successfully!" Green
        
        $polyagentPath = (Get-Command polyagent).Source
        Write-ColorOutput "Location: $polyagentPath" Green
        
        # Try to run polyagent
        try {
            & $polyagentPath --version 2>$null
        }
        catch {
            Write-ColorOutput "PolyAgent is ready to use! Run 'polyagent' to start." Yellow
        }
        
        return $true
    }
    else {
        Write-ColorOutput "Installation completed, but polyagent is not in PATH." Red
        return $false
    }
}

# Main installation process
function Install-PolyAgent {
    Write-ColorOutput "PolyAgent Installer for Windows" Green
    Write-ColorOutput "===============================" White
    Write-Host
    
    # Get version
    if ([string]::IsNullOrEmpty($Version)) {
        $script:Version = Get-LatestVersion
        Write-ColorOutput "Latest version detected: $Version" Green
    }
    else {
        Write-ColorOutput "Using specified version: $Version" Green
    }
    
    $script:ReleaseUrl = "https://github.com/$Repo/releases/download/$Version"
    
    # Detect architecture
    $arch = Get-WindowsArchitecture
    Write-ColorOutput "Detected architecture: $arch" Green
    
    $binaryName = "polyagent-windows-$arch.exe"
    $downloadUrl = "$ReleaseUrl/$binaryName"
    $checksumUrl = "$ReleaseUrl/checksums.txt"
    
    # Create installation directory
    New-InstallationDirectory -Path $InstallDir
    
    # Download binary
    $binaryPath = Join-Path $InstallDir "polyagent.exe"
    Invoke-PolyAgentDownload -Url $downloadUrl -OutputPath $binaryPath
    
    # Download and verify checksum if available
    Write-ColorOutput "Verifying checksum..." Yellow
    $checksumPath = Join-Path $env:TEMP "checksums.txt"
    if (Invoke-DownloadWithRetry -Url $checksumUrl -OutputPath $checksumPath) {
        $checksumContent = Get-Content $checksumPath
        $expectedChecksum = ($checksumContent | Select-String -Pattern $binaryName).ToString().Split()[0]
        
        if ($expectedChecksum) {
            $actualChecksum = (Get-FileHash -Path $binaryPath -Algorithm SHA256).Hash
            if ($expectedChecksum -ne $actualChecksum) {
                Write-ColorOutput "Checksum verification failed!" Red
                Write-ColorOutput "Expected: $expectedChecksum" Red
                Write-ColorOutput "Actual: $actualChecksum" Red
                exit 1
            }
            Write-ColorOutput "Checksum verified successfully!" Green
        }
        else {
            Write-ColorOutput "No checksum found for this binary, skipping verification." Yellow
        }
        
        Remove-Item $checksumPath -Force
    }
    else {
        Write-ColorOutput "Could not download checksums.txt, skipping verification." Yellow
    }
    
    # Add to PATH if not already there
    Add-ToPath -Path $InstallDir
    
    Write-Host
    
    # Verify installation
    if (Test-PolyAgentInstallation) {
        Write-ColorOutput "Installation complete!" Green
        Write-Host
        Write-ColorOutput "To get started, run:" Yellow
        Write-ColorOutput "  polyagent" Yellow
    }
    else {
        Write-ColorOutput "Installation completed with warnings." Yellow
        Write-ColorOutput "You may need to restart your terminal or manually add $InstallDir to your PATH." Yellow
    }
}

# Handle script interruption
trap {
    Write-ColorOutput "Installation interrupted." Red
    exit 1
}

# Run installation
Install-PolyAgent
