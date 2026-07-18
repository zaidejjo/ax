# typed: false
# frozen_string_literal: true

# ─── ax — TUI API Client ──────────────────────────────────────────────────────
# Manual Homebrew formula for zaidejjo/tap/ax.
#
# Install:
#   brew tap zaidejjo/tap
#   brew install ax
#
# Upgrade:
#   brew upgrade ax
# ────────────────────────────────────────────────────────────────────────────────

class Ax < Formula
  desc "ax — TUI API Client: a terminal-based HTTP client with xh/httpie-style syntax parsing"
  homepage "https://github.com/zaidejjo/ax"
  license "MIT"
  version "0.1.0"

  # ── macOS Intel ────────────────────────────────────────────────────────────
  on_macos do
    on_intel do
      url "https://github.com/zaidejjo/ax/releases/download/v#{version}/ax_#{version}_macOS_amd64.tar.gz"
      # ⚠  Fill after downloading: shasum -a 256 ax_0.1.0_macOS_amd64.tar.gz
      sha256 "AMD64_MACOS_SHA256_PLACEHOLDER"
    end

    # ── macOS Apple Silicon ─────────────────────────────────────────────────
    on_arm do
      url "https://github.com/zaidejjo/ax/releases/download/v#{version}/ax_#{version}_macOS_arm64.tar.gz"
      # ⚠  Fill after downloading: shasum -a 256 ax_0.1.0_macOS_arm64.tar.gz
      sha256 "ARM64_MACOS_SHA256_PLACEHOLDER"
    end
  end

  # ── Linux Intel ───────────────────────────────────────────────────────────
  on_linux do
    on_intel do
      url "https://github.com/zaidejjo/ax/releases/download/v#{version}/ax_#{version}_linux_amd64.tar.gz"
      # ⚠  Fill after downloading: shasum -a 256 ax_0.1.0_linux_amd64.tar.gz
      sha256 "AMD64_LINUX_SHA256_PLACEHOLDER"
    end

    # ── Linux ARM64 ─────────────────────────────────────────────────────────
    on_arm do
      url "https://github.com/zaidejjo/ax/releases/download/v#{version}/ax_#{version}_linux_arm64.tar.gz"
      # ⚠  Fill after downloading: shasum -a 256 ax_0.1.0_linux_arm64.tar.gz
      sha256 "ARM64_LINUX_SHA256_PLACEHOLDER"
    end
  end

  # ── Optional clipboard dependencies ───────────────────────────────────────
  depends_on "xclip" => :optional      # X11 clipboard (Linux)
  depends_on "wl-clipboard" => :optional # Wayland clipboard (Linux)

  # ── Install ───────────────────────────────────────────────────────────────
  def install
    bin.install "ax"
  end

  # ── Test ──────────────────────────────────────────────────────────────────
  test do
    assert_match version.to_s, shell_output("#{bin}/ax --version")
  end
end
