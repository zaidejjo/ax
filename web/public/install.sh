#!/usr/bin/env bash
# ax — Smart Install Script
# https://github.com/zaidejjo/ax
# Detects OS and architecture, then installs the binary.
set -euo pipefail

# ── Color helpers ──
CYAN='\033[0;36m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# ── Detect OS ──
OS="$(uname -s)"
ARCH="$(uname -m)"

echo -e "${CYAN}==>${NC} Detected OS: $OS, Arch: $ARCH"

# ── macOS: Install via Homebrew ──
if [ "$OS" = "Darwin" ]; then
  if ! command -v brew &>/dev/null; then
    echo -e "${RED}==>${NC} Homebrew not found. Please install it first: https://brew.sh"
    exit 1
  fi
  brew tap zaidejjo/tap
  brew install ax-cli
  echo -e "${GREEN}==>${NC} ax installed via Homebrew!"
  exit 0
fi

# ── Linux: Check distro ──
if [ "$OS" = "Linux" ]; then
  if [ -f /etc/os-release ]; then
    . /etc/os-release
  fi

  case "${ID,,}" in
    arch|manjaro)
      echo -e "${CYAN}==>${NC} Installing via AUR..."
      if command -v yay &>/dev/null; then
        yay -S ax-cli
      elif command -v paru &>/dev/null; then
        paru -S ax-cli
      else
        echo -e "${RED}==>${NC} No AUR helper found. Install yay or paru first."
        exit 1
      fi
      ;;
    ubuntu|debian|pop|mint)
      echo -e "${CYAN}==>${NC} Downloading .deb package..."
      GH="https://github.com/zaidejjo/ax/releases/latest"
      DEB_ARCH="$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')"
      curl -sSL "$GH/download/ax_linux_${DEB_ARCH}.deb" -o /tmp/ax.deb
      sudo dpkg -i /tmp/ax.deb
      ;;
    fedora|rhel|centos)
      echo -e "${CYAN}==>${NC} Downloading .rpm package..."
      GH="https://github.com/zaidejjo/ax/releases/latest"
      RPM_ARCH="$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')"
      curl -sSL "$GH/download/ax_linux_${RPM_ARCH}.rpm" -o /tmp/ax.rpm
      sudo rpm -i /tmp/ax.rpm
      ;;
    *)
      echo -e "${CYAN}==>${NC} Downloading universal binary..."
      GH="https://github.com/zaidejjo/ax/releases/latest"
      TAR_ARCH="$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')"
      curl -sSL "$GH/download/ax_linux_${TAR_ARCH}.tar.gz" -o /tmp/ax.tar.gz
      sudo tar xzf /tmp/ax.tar.gz -C /usr/local/bin ax
      ;;
  esac

  echo -e "${GREEN}==>${NC} ax installed successfully!"
  echo -e "${CYAN}==>${NC} Run 'ax' to start the TUI."
  exit 0
fi

# ── Unsupported OS ──
echo -e "${RED}==>${NC} Unsupported OS: $OS"
echo "Please download the binary from https://github.com/zaidejjo/ax/releases"
exit 1
