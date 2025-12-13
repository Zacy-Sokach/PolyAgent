#!/bin/bash

set -e

# PolyAgent Installer Script
# Supports Linux and macOS

REPO="Zacy-Sokach/PolyAgent"
VERSION="${VERSION:-}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get latest version from GitHub API
get_latest_version() {
    local api_url="https://api.github.com/repos/${REPO}/tags"
    echo -e "${YELLOW}Fetching latest version from GitHub...${NC}"
    
    if command -v curl &> /dev/null; then
        VERSION=$(curl -s "$api_url" | grep -o '"name": *"[^"]*"' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget &> /dev/null; then
        VERSION=$(wget -qO- "$api_url" | grep -o '"name": *"[^"]*"' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')
    else
        echo -e "${RED}Error: curl or wget is required but not installed.${NC}"
        exit 1
    fi
    
    if [ -z "$VERSION" ]; then
        echo -e "${RED}Failed to fetch latest version. Please check your network connection or specify VERSION manually.${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}Latest version: $VERSION${NC}"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac
    
    case "$OS" in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        *)
            echo -e "${RED}Unsupported operating system: $OS${NC}"
            exit 1
            ;;
    esac
    
    echo -e "${GREEN}Detected platform: $OS-$ARCH${NC}"
}

# Check for required commands
check_dependencies() {
    if ! command -v curl &> /dev/null && ! command -v wget &> /dev/null; then
        echo -e "${RED}Error: curl or wget is required but not installed.${NC}"
        exit 1
    fi
}

# Download file with retry
download_with_retry() {
    local url="$1"
    local output="$2"
    local max_attempts=3
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        echo -e "${YELLOW}Download attempt $attempt/$max_attempts...${NC}"
        
        if command -v curl &> /dev/null; then
            if curl -fsSL --connect-timeout 10 --retry 2 "$url" -o "$output"; then
                return 0
            fi
        elif command -v wget &> /dev/null; then
            if wget -q --timeout=10 --tries=3 "$url" -O "$output"; then
                return 0
            fi
        fi
        
        attempt=$((attempt + 1))
        if [ $attempt -le $max_attempts ]; then
            echo -e "${YELLOW}Download failed, retrying in 2 seconds...${NC}"
            sleep 2
        fi
    done
    
    return 1
}

# Download and install PolyAgent
install_polyagent() {
    local binary_name="polyagent-${OS}-${ARCH}"
    local download_url="https://github.com/${REPO}/releases/download/${VERSION}/${binary_name}"
    local checksum_url="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"
    
    echo -e "${YELLOW}Downloading PolyAgent from:${NC}"
    echo -e "${YELLOW}$download_url${NC}"
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    
    # Download binary
    if ! download_with_retry "$download_url" "$TEMP_DIR/polyagent"; then
        echo -e "${RED}Failed to download PolyAgent after multiple attempts.${NC}"
        echo -e "${YELLOW}Please check the version and network connection.${NC}"
        exit 1
    fi
    
    # Download and verify checksum if available
    echo -e "${YELLOW}Verifying checksum...${NC}"
    if download_with_retry "$checksum_url" "$TEMP_DIR/checksums.txt"; then
        local expected_checksum=$(grep "$binary_name" "$TEMP_DIR/checksums.txt" | awk '{print $1}')
        if [ -n "$expected_checksum" ]; then
            local actual_checksum=$(sha256sum "$TEMP_DIR/polyagent" | awk '{print $1}')
            if [ "$expected_checksum" != "$actual_checksum" ]; then
                echo -e "${RED}Checksum verification failed!${NC}"
                echo -e "${RED}Expected: $expected_checksum${NC}"
                echo -e "${RED}Actual: $actual_checksum${NC}"
                exit 1
            fi
            echo -e "${GREEN}Checksum verified successfully!${NC}"
        else
            echo -e "${YELLOW}No checksum found for this binary, skipping verification.${NC}"
        fi
    else
        echo -e "${YELLOW}Could not download checksums.txt, skipping verification.${NC}"
    fi
    
    # Make it executable
    chmod +x "$TEMP_DIR/polyagent"
    
    # Install to target directory
    echo -e "${YELLOW}Installing PolyAgent to $INSTALL_DIR...${NC}"
    
    if [ -w "$INSTALL_DIR" ]; then
        mv "$TEMP_DIR/polyagent" "$INSTALL_DIR/polyagent"
    else
        echo -e "${YELLOW}Sudo access required to install to $INSTALL_DIR${NC}"
        sudo mv "$TEMP_DIR/polyagent" "$INSTALL_DIR/polyagent"
    fi
    
    # Verify installation
    if command -v polyagent &> /dev/null; then
        echo -e "${GREEN}PolyAgent installed successfully!${NC}"
        echo -e "${GREEN}Version: $VERSION${NC}"
        echo -e "${GREEN}Location: $(which polyagent)${NC}"
        echo -e "${GREEN}You can now run 'polyagent' from anywhere!${NC}"
        
        # Run polyagent to initialize configuration
        echo -e "${YELLOW}Running PolyAgent for initial setup...${NC}"
        polyagent --version 2>/dev/null || echo -e "${YELLOW}PolyAgent is ready to use! Run 'polyagent' to start.${NC}"
    else
        echo -e "${RED}Installation completed, but polyagent is not in PATH.${NC}"
        echo -e "${YELLOW}Please add $INSTALL_DIR to your PATH or run:${NC}"
        echo -e "${YELLOW}  export PATH=\$PATH:$INSTALL_DIR${NC}"
        echo -e "${YELLOW}Then you can run 'polyagent' from anywhere.${NC}"
    fi
}

# Main installation process
main() {
    echo -e "${GREEN}PolyAgent Installer${NC}"
    echo "=================="
    echo ""
    
    if [ -z "$VERSION" ]; then
        get_latest_version
    else
        echo -e "${GREEN}Using specified version: $VERSION${NC}"
    fi
    
    detect_platform
    check_dependencies
    install_polyagent
    
    echo ""
    echo -e "${GREEN}Installation complete!${NC}"
    echo -e "${YELLOW}To get started, run:${NC}"
    echo -e "${YELLOW}  polyagent${NC}"
}

# Handle script interruption
trap 'echo -e "\n${RED}Installation interrupted.${NC}"; exit 1' INT

# Run main function
main "$@"
