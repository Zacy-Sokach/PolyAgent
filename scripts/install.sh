#!/bin/bash

set -e

# PolyAgent Installer Script
# Supports Linux and macOS

REPO="Zacy-Sokach/PolyAgent"
VERSION="${VERSION:-v25.12.13-nightly-1}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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
    for cmd in curl tar; do
        if ! command -v "$cmd" &> /dev/null; then
            echo -e "${RED}Error: $cmd is required but not installed.${NC}"
            exit 1
        fi
    done
}

# Download and install PolyAgent
install_polyagent() {
    local binary_name="polyagent-${OS}-${ARCH}"
    local download_url="https://github.com/${REPO}/releases/download/${VERSION}/${binary_name}"
    
    echo -e "${YELLOW}Downloading PolyAgent from:${NC}"
    echo -e "${YELLOW}$download_url${NC}"
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    
    # Download binary
    if ! curl -fsSL "$download_url" -o "$TEMP_DIR/polyagent"; then
        echo -e "${RED}Failed to download PolyAgent. Please check the version and network connection.${NC}"
        exit 1
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
        
        # Run polyagent to initialize configuration
        echo -e "${YELLOW}Running PolyAgent for initial setup...${NC}"
        polyagent --version 2>/dev/null || echo -e "${YELLOW}PolyAgent is ready to use! Run 'polyagent' to start.${NC}"
    else
        echo -e "${RED}Installation completed, but polyagent is not in PATH.${NC}"
        echo -e "${YELLOW}Please add $INSTALL_DIR to your PATH or run:${NC}"
        echo -e "${YELLOW}  export PATH=\$PATH:$INSTALL_DIR${NC}"
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
