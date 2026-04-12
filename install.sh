#!/bin/bash
# Install or update doc-helper
# Usage: curl -fsSL https://raw.githubusercontent.com/onurkerem/doc-helper/main/install.sh | bash

set -e

if ! command -v go &>/dev/null; then
	echo "Error: Go is not installed."
	echo "Install it with: brew install go"
	exit 1
fi

GOBIN="$(go env GOPATH)/bin"

echo "Installing doc-helper..."
go install github.com/onurkerem/doc-helper@latest

case ":$PATH:" in
	*":$GOBIN:"*)
		echo "Installed: $GOBIN/doc-helper"
		;;
	*)
		echo "" >> ~/.zshrc
		echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
		echo "Added $GOBIN to PATH in ~/.zshrc"
		echo "Installed: $GOBIN/doc-helper"
		echo "Run 'source ~/.zshrc' or open a new terminal to use doc-helper"
		;;
esac
