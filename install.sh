#!/bin/bash
# Install or update doc-helper
# Usage: curl -fsSL https://raw.githubusercontent.com/onurkerem/doc-helper/main/install.sh | bash

set -e

if ! command -v go &>/dev/null; then
	echo "Error: Go is not installed. Install it from https://go.dev/dl/"
	exit 1
fi

echo "Installing doc-helper..."
go install github.com/onurkerem/doc-helper@latest

echo "Installed: $(go env GOPATH)/bin/doc-helper"
