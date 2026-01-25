#!/usr/bin/env bash
set -euo pipefail

rm -rf build
mkdir -p build

# Build Windows executable
GOOS=windows GOARCH=amd64 go build -o build/gallery.exe .

# Download ffmpeg (Windows build)
mkdir -p build/tmp
curl -L -o build/tmp/ffmpeg.zip https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip
unzip -q build/tmp/ffmpeg.zip -d build/tmp/ffmpeg

# Find ffmpeg.exe and ffprobe.exe inside the extracted folder
ffprobe_path=$(find build/tmp/ffmpeg -name ffprobe.exe | head -n 1)

cp "$ffprobe_path" build/

# Download cloudflared (Windows)
curl -L -o build/cloudflared.exe https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-windows-amd64.exe

# Drop temp artifacts so they don't end up in the ZIP.
rm -rf build/tmp

# Create ZIP
cd build
zip -r ../Gallery-windows-x64.zip .
