#!/bin/bash

# Create a build directory
mkdir -p builds

# Build for Intel Mac (amd64)
GOOS=darwin GOARCH=amd64 go build -o builds/extract-mac-amd64
# Build for Apple Silicon (arm64)
GOOS=darwin GOARCH=arm64 go build -o builds/extract-mac-arm64

# Create universal binary using lipo (if you have macOS tools)
if command -v lipo &> /dev/null; then
    lipo -create -output builds/extract-mac builds/extract-mac-amd64 builds/extract-mac-arm64
    rm builds/extract-mac-amd64 builds/extract-mac-arm64
    echo "Created universal binary: builds/extract-mac"
else
    # Create a simple install script for Mac users
    cat > builds/install.sh << 'EOF'
#!/bin/bash

# Detect architecture
ARCH=$(uname -m)
if [ "$ARCH" = "arm64" ]; then
    BINARY="extract-mac-arm64"
else
    BINARY="extract-mac-amd64"
fi

# Move the appropriate binary to /usr/local/bin
sudo mv "$BINARY" /usr/local/bin/extract
sudo chmod +x /usr/local/bin/extract

echo "Extract CLI installed successfully!"
EOF
    chmod +x builds/install.sh
    echo "Created separate binaries for Intel and Apple Silicon Macs"
fi

# Create a zip file containing all files
cd builds
zip -r extract-mac.zip extract-mac* install.sh 2>/dev/null
cd ..

echo "Build complete! Files are in the builds directory"
echo "Distribute builds/extract-mac.zip to Mac users"
