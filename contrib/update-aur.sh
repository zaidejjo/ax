#!/usr/bin/env bash
set -euo pipefail

# ─── update-aur.sh ─────────────────────────────────────────────────────────────
# Updates contrib/PKGBUILD and contrib/.SRCINFO for a new ax release.
# Usage: ./contrib/update-aur.sh <version>
# Example: ./contrib/update-aur.sh 1.0.0
# ────────────────────────────────────────────────────────────────────────────────

cd "$(dirname "$0")"

if [ $# -ne 1 ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 1.0.0"
  exit 1
fi

VERSION="$1"
RELEASE_URL="https://github.com/zaidejjo/ax/releases/download/v${VERSION}"

echo "==> Fetching checksums for v${VERSION} ..."

CHECKSUMS=$(curl -sfL "${RELEASE_URL}/checksums.txt" || true)
if [ -z "$CHECKSUMS" ]; then
  echo "!! Could not fetch checksums.txt from ${RELEASE_URL}"
  echo "!! Release may not exist yet or the URL is wrong."
  exit 1
fi

SHA256_AMD64=$(echo "$CHECKSUMS" | grep "linux_amd64.tar.gz" | awk '{print $1}')
SHA256_ARM64=$(echo "$CHECKSUMS" | grep "linux_arm64.tar.gz" | awk '{print $1}')

if [ -z "$SHA256_AMD64" ] || [ -z "$SHA256_ARM64" ]; then
  echo "!! Could not find sha256 for linux_amd64 and/or linux_arm64 in checksums.txt"
  exit 1
fi

echo "  amd64: ${SHA256_AMD64}"
echo "  arm64: ${SHA256_ARM64}"

# Update PKGBUILD
sed -i "s/^pkgver=.*/pkgver=${VERSION}/" PKGBUILD
sed -i "s|^source_x86_64=.*|source_x86_64=(\"\${url}/releases/download/v\${pkgver}/ax_\${pkgver}_linux_amd64.tar.gz\")|" PKGBUILD
sed -i "s|^source_aarch64=.*|source_aarch64=(\"\${url}/releases/download/v\${pkgver}/ax_\${pkgver}_linux_arm64.tar.gz\")|" PKGBUILD
sed -i "s/^sha256sums_x86_64=.*/sha256sums_x86_64=('${SHA256_AMD64}')/" PKGBUILD
sed -i "s/^sha256sums_aarch64=.*/sha256sums_aarch64=('${SHA256_ARM64}')/" PKGBUILD

echo "==> PKGBUILD updated."

# Regenerate .SRCINFO
makepkg --printsrcinfo > .SRCINFO
echo "==> .SRCINFO regenerated."

echo ""
echo "Done! Review the changes, then copy PKGBUILD and .SRCINFO to the AUR repo."
echo "  git diff"
echo "  git clone ssh://aur@aur.archlinux.org/ax.git /tmp/aur-ax"
echo "  cp PKGBUILD .SRCINFO /tmp/aur-ax/"
echo "  cd /tmp/aur-ax && git commit -m 'Update to v${VERSION}' && git push"
