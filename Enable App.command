#!/bin/bash
# This script removes the macOS quarantine attribute from Sanskrit Upāya.app
# allowing it to run without Gatekeeper blocking it.
#
# When you download an app from the internet, macOS adds a "quarantine"
# attribute to it. This tells Gatekeeper to check if the app is signed
# by a verified developer. Since this app is not signed, Gatekeeper
# blocks it. This script simply removes that quarantine attribute,
# telling macOS to trust the app.

APP_PATH="/Applications/Sanskrit Upāya.app"

echo ""
echo "  Sanskrit Upāya - Enable App"
echo "  ==========================="
echo ""

if [ -d "$APP_PATH" ]; then
    echo "  Removing quarantine attribute..."
    xattr -cr "$APP_PATH"
    echo ""
    echo "  ✓ Done! You can now open Sanskrit Upāya from Applications."
    echo ""
else
    echo "  ✗ Sanskrit Upāya.app not found in /Applications"
    echo ""
    echo "  Please drag Sanskrit Upāya.app to your Applications"
    echo "  folder first, then run this script again."
    echo ""
fi

echo "  Press any key to close..."
read -n 1
