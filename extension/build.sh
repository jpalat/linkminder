#!/bin/bash

# BookMinder Extension Build Script
# Generates browser-specific packages for Chrome, Firefox, and Safari

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/build"
VERSION=$(grep '"version"' "$SCRIPT_DIR/manifest.json" | sed 's/.*"version": "\([^"]*\)".*/\1/')

echo "Building BookMinder Extension v$VERSION"
echo "========================================="

# Clean and create build directory
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Function to copy common files
copy_common_files() {
    local dest_dir="$1"
    
    # Copy all files except manifests and build scripts
    cp "$SCRIPT_DIR"/*.js "$dest_dir/"
    cp "$SCRIPT_DIR"/*.html "$dest_dir/"
    cp "$SCRIPT_DIR"/*.png "$dest_dir/"
    
    # Copy README if it exists
    if [ -f "$SCRIPT_DIR/README.md" ]; then
        cp "$SCRIPT_DIR/README.md" "$dest_dir/"
    fi
}

# Build Chrome/Edge package (Manifest V3)
echo "Building Chrome/Edge package..."
CHROME_DIR="$BUILD_DIR/chrome"
mkdir -p "$CHROME_DIR"
copy_common_files "$CHROME_DIR"
cp "$SCRIPT_DIR/manifest.json" "$CHROME_DIR/"

# Create Chrome zip
cd "$BUILD_DIR"
zip -r "bookminder-chrome-v$VERSION.zip" chrome/
echo "✓ Chrome package: bookminder-chrome-v$VERSION.zip"

# Build Firefox package (Manifest V2)
echo "Building Firefox package..."
FIREFOX_DIR="$BUILD_DIR/firefox"
mkdir -p "$FIREFOX_DIR"
copy_common_files "$FIREFOX_DIR"
cp "$SCRIPT_DIR/manifest_v2.json" "$FIREFOX_DIR/manifest.json"

# Create Firefox zip
zip -r "bookminder-firefox-v$VERSION.zip" firefox/
echo "✓ Firefox package: bookminder-firefox-v$VERSION.zip"

# Build Safari package (same as Chrome but separate for clarity)
echo "Building Safari package..."
SAFARI_DIR="$BUILD_DIR/safari"
mkdir -p "$SAFARI_DIR"
copy_common_files "$SAFARI_DIR"
cp "$SCRIPT_DIR/manifest.json" "$SAFARI_DIR/"

# Create Safari zip
zip -r "bookminder-safari-v$VERSION.zip" safari/
echo "✓ Safari package: bookminder-safari-v$VERSION.zip"

cd "$SCRIPT_DIR"

echo ""
echo "Build Summary:"
echo "=============="
echo "✓ Chrome/Edge:  $BUILD_DIR/bookminder-chrome-v$VERSION.zip"
echo "✓ Firefox:      $BUILD_DIR/bookminder-firefox-v$VERSION.zip"  
echo "✓ Safari:       $BUILD_DIR/bookminder-safari-v$VERSION.zip"
echo ""
echo "Installation Instructions:"
echo "========================="
echo "Chrome/Edge: chrome://extensions/ → Load unpacked or drag zip"
echo "Firefox:     about:debugging → This Firefox → Load Temporary Add-on"
echo "Safari:      Requires Safari Web Extension converter tool"
echo ""
echo "Note: Remember to configure the API URL in each browser's extension settings!"