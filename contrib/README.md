# AUR Package — ax

This directory contains the Arch Linux PKGBUILD for [ax](https://github.com/zaidejjo/ax).

## Publishing to AUR

```bash
# 1. Update pkgver and sha256sums in PKGBUILD
#    Get checksums from the GitHub release's checksums.txt:
#      https://github.com/zaidejjo/ax/releases/download/v1.0.0/checksums.txt

# 2. Generate .SRCINFO
makepkg --printsrcinfo > .SRCINFO

# 3. Clone the AUR repository
git clone ssh://aur@aur.archlinux.org/ax.git /tmp/aur-ax

# 4. Copy files
cp PKGBUILD .SRCINFO /tmp/aur-ax/

# 5. Commit and push
cd /tmp/aur-ax
git add -A
git commit -m "Update to v1.0.0"
git push
```

## Users install with

```bash
# Using an AUR helper
yay -S ax
paru -S ax

# Or manually
git clone https://aur.archlinux.org/ax.git
cd ax
makepkg -si
```
