# PolyAgent Installer Script for Windows
# PowerShell script

param(
    [string]$Version = "",
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

# Detect Windows architecture
function Get-WindowsArchitecture {
    $arch = (Get-WmiObject Win32_Processor).Architecture
    switch ($arch) {
        0 { return "amd64" }      # x86
        9 { return "amd64" }      # x64
        12 { return "arm64" }     # ARM64
        default { 
            Write-ColorOutput "Unsupported architecture detected." Red
            exit 1 
        }
    }
}

# Check if running as Administrator
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Create installation directory
function New-InstallationDirectory {
    param([string]$Path)
    
    if (!(Test-Path $Path)) {
        Write-ColorOutput "Creating installation directory: $Path" Yellow
        New-Item -ItemType Directory -Path $Path -Force | Out-Null
    }
}

# Download PolyAgent binary
function Invoke-PolyAgentDownload {
    param(
        [string]$Url,
        [string]$OutputPath
    )
    
    Write-ColorOutput "Downloading PolyAgent from:" Yellow
    Write-ColorOutput $Url - Yellow
    
    try {
        $webClient = New-Object System.Net.WebClient
        $webClient.DownloadFile($Url, $OutputPath)
        Write-ColorOutput "Download completed successfully!" Green
    }
    catch {
        Write-ColorOutput "Failed to download PolyAgent: $_" Red
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
        
        # Also update current session
        $env:PATH = "$env:PATH;$Path"
        
        Write-ColorOutput "PATH updated. Please restart your terminal or run:" Yellow
        Write-ColorOutput "  `$env:PATH = `"$Path;`$env:PATH`"" Yellow
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
    
    # Create installation directory
    New-InstallationDirectory -Path $InstallDir
    
    # Download binary
    $binaryPath = Join-Path $InstallDir "polyagent.exe"
    Invoke-PolyAgentDownload -Url $downloadUrl -OutputPath $binaryPath
    
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
