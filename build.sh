#!/bin/bash

set -e

PROJECT_NAME="gospeed"
GO_VERSION="1.23.6"

check_go() {
    if command -v go >/dev/null 2>&1; then
        echo "Go is already installed: $(go version)"
        return 0
    else
        echo "Go not found. Installing Go $GO_VERSION..."
        return 1
    fi
}

install_go() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
    esac
    
    GO_URL="https://go.dev/dl/go${GO_VERSION}.${OS}-${ARCH}.tar.gz"
    
    echo "Downloading Go for $OS-$ARCH..."
    curl -L "$GO_URL" -o "/tmp/go${GO_VERSION}.tar.gz"
    
    echo "Installing Go to /usr/local/go..."
    if [ "$EUID" -eq 0 ]; then
        rm -rf /usr/local/go
        tar -C /usr/local -xzf "/tmp/go${GO_VERSION}.tar.gz"
    else
        echo "Installing Go to $HOME/go-install..."
        mkdir -p "$HOME/go-install"
        rm -rf "$HOME/go-install/go"
        tar -C "$HOME/go-install" -xzf "/tmp/go${GO_VERSION}.tar.gz"
        export PATH="$HOME/go-install/go/bin:$PATH"
        echo "Add this to your shell profile:"
        echo "export PATH=\"\$HOME/go-install/go/bin:\$PATH\""
    fi
    
    rm "/tmp/go${GO_VERSION}.tar.gz"
    echo "Go installed successfully"
}

build_project() {
    echo "Building $PROJECT_NAME..."
    
    if [ ! -f "go.mod" ]; then
        echo "Error: go.mod not found. Run this script from the project root."
        exit 1
    fi
    
    go build -o "$PROJECT_NAME" .
    echo "Built $PROJECT_NAME successfully"
    
    echo "Installation complete"
}

main() {
    if ! check_go; then
        install_go
    fi
    
    build_project
    
    echo "Done. You can now run: $PROJECT_NAME"
}

main
