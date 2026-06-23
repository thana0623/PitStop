#!/bin/bash
# Cross-compile PitStop for all platforms
set -e

cd "$(dirname "$0")/.."

echo "Building PitStop for all platforms..."

echo "  → win32-x64"
GOOS=windows GOARCH=amd64 go build -o npm/pitstop-win32-x64/pitstop.exe .

echo "  → linux-x64"
GOOS=linux GOARCH=amd64 go build -o npm/pitstop-linux-x64/pitstop .

echo "  → darwin-x64"
GOOS=darwin GOARCH=amd64 go build -o npm/pitstop-darwin-x64/pitstop .

echo "  → darwin-arm64"
GOOS=darwin GOARCH=arm64 go build -o npm/pitstop-darwin-arm64/pitstop .

echo ""
echo "Done! Binaries in npm/pitstop-*/"
echo ""
echo "To publish:"
echo "  npm publish npm/pitstop-win32-x64/"
echo "  npm publish npm/pitstop-linux-x64/"
echo "  npm publish npm/pitstop-darwin-x64/"
echo "  npm publish npm/pitstop-darwin-arm64/"
echo "  npm publish npm/pitstop/"
