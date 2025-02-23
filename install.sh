#!/bin/bash

# Detect OS and architecture
OS=$(uname -s)
ARCH=$(uname -m)

# Set the output binary name
BINARY_NAME="kubectllama"

# Define a function to build for each OS
build_for_os() {
  case $OS in
    Linux)
      GOOS=linux GOARCH=amd64 go build -o $BINARY_NAME-linux
      ;;
    Darwin)
      GOOS=darwin GOARCH=amd64 go build -o $BINARY_NAME-darwin
      ;;
    CYGWIN*|MINGW*)
      GOOS=windows GOARCH=amd64 go build -o $BINARY_NAME.exe
      ;;
    *)
      echo "Unsupported OS: $OS"
      exit 1
      ;;
  esac
}

# Build the project for the detected OS
echo "Building for $OS ($ARCH)..."
build_for_os

# Define the directory for global installation (usually /usr/local/bin for Linux/macOS)
INSTALL_DIR="$HOME/bin"  # Change this to a directory you have write access to
mkdir -p $INSTALL_DIR

# Move the appropriate binary to the installation directory
echo "Moving binary to $INSTALL_DIR..."

if [ "$OS" == "Linux" ]; then
  mv $BINARY_NAME-linux $INSTALL_DIR/$BINARY_NAME
elif [ "$OS" == "Darwin" ]; then
  mv $BINARY_NAME-darwin $INSTALL_DIR/$BINARY_NAME
elif [ "$OS" == "CYGWIN"* ] || [ "$OS" == "MINGW"* ]; then
  mv $BINARY_NAME.exe $INSTALL_DIR/$BINARY_NAME.exe
fi

# Make sure the binary is executable
chmod +x $INSTALL_DIR/$BINARY_NAME

echo "Installation complete. You can now use the command: $BINARY_NAME"
