#!/bin/bash
set -e

APP_NAME="Gallery"
BIN_NAME="Gallery-bin"
GOARCH="${1:-amd64}"
DMG_NAME="${2:-Gallery-macos-${GOARCH}.dmg}"
VOLUME_NAME="Gallery Installer"
FFPROBE_URL="https://evermeet.cx/ffmpeg/getrelease/ffprobe/zip"

echo "==> Cleanup"
rm -rf "$APP_NAME.app"
rm -f "$DMG_NAME"

echo "==> Building Go binary (${GOARCH})"
GOOS=darwin GOARCH="$GOARCH" go build -o "$BIN_NAME"

echo "==> Creating AppleScript launcher"
cat > launcher.applescript <<EOF
on run
    set appPath to POSIX path of (path to me)
    set binDir to appPath & "Contents/Resources"

    tell application "Terminal"
        activate
        do script "cd " & quoted form of binDir & "; ./${BIN_NAME}; echo ''; echo 'Process exited. Press Enter to close.'; read"
    end tell
end run
EOF

echo "==> Compiling AppleScript app"
osacompile -o "$APP_NAME.app" launcher.applescript
rm launcher.applescript

echo "==> Installing binaries"
mkdir -p "$APP_NAME.app/Contents/Resources"

mv "$BIN_NAME" "$APP_NAME.app/Contents/Resources/"

echo "==> Bundling ffprobe (static)"
TMP_FFPROBE_DIR="$(mktemp -d)"
curl -L -o "$TMP_FFPROBE_DIR/ffprobe.zip" "$FFPROBE_URL"
unzip -q "$TMP_FFPROBE_DIR/ffprobe.zip" -d "$TMP_FFPROBE_DIR"
if [[ -x "$TMP_FFPROBE_DIR/ffprobe" ]]; then
  cp "$TMP_FFPROBE_DIR/ffprobe" "$APP_NAME.app/Contents/Resources/"
else
  ffprobe_bin="$(find "$TMP_FFPROBE_DIR" -type f -name ffprobe -perm -111 | head -n 1)"
  if [[ -z "$ffprobe_bin" ]]; then
    echo "ffprobe not found in downloaded archive" >&2
    exit 1
  fi
  cp "$ffprobe_bin" "$APP_NAME.app/Contents/Resources/"
fi
rm -rf "$TMP_FFPROBE_DIR"

echo "==> Bundling cloudflared"
cp "$(which cloudflared)" "$APP_NAME.app/Contents/Resources/"

chmod +x "$APP_NAME.app/Contents/Resources/"*

echo "==> Code signing (ad-hoc)"
codesign --force --deep --sign - "$APP_NAME.app"

echo "==> Creating DMG layout"
WORKDIR=$(mktemp -d)
mkdir -p "$WORKDIR/$APP_NAME"

cp -R "$APP_NAME.app" "$WORKDIR/$APP_NAME/"
ln -s /Applications "$WORKDIR/$APP_NAME/Applications"

echo "==> Building DMG"
hdiutil create \
  -srcfolder "$WORKDIR/$APP_NAME" \
  -volname "$VOLUME_NAME" \
  -fs HFS+ \
  -format UDZO \
  "$DMG_NAME"

rm -rf "$WORKDIR"

echo "==> Done"
echo "Created $DMG_NAME"
