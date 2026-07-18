package tui

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

// ─── Copy ────────────────────────────────────────────────────────────────────

// copyToClipboard copies the given text to the system clipboard by shelling out
// to platform-appropriate tools. It tries xclip (X11), wl-copy (Wayland), and
// pbcopy (macOS) in order. Returns an error if no clipboard tool is found.
func copyToClipboard(text string) error {
	// Try xclip (X11) first.
	if _, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			return err
		}
		return nil
	}

	// Try wl-copy (Wayland).
	if _, err := exec.LookPath("wl-copy"); err == nil {
		cmd := exec.Command("wl-copy")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			return err
		}
		return nil
	}

	// Try pbcopy (macOS).
	if _, err := exec.LookPath("pbcopy"); err == nil {
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			return err
		}
		return nil
	}

	return errors.New("clipboard: no clipboard tool found (install xclip, wl-copy, or pbcopy)")
}

// ─── Paste ───────────────────────────────────────────────────────────────────

// pasteFromClipboard reads the current system clipboard content.
// It tries xclip (X11), wl-paste (Wayland), and pbpaste (macOS) in order.
// Returns the clipboard text, or an error if no clipboard tool is found.
func pasteFromClipboard() (string, error) {
	// Try xclip (X11) first.
	if _, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command("xclip", "-o", "-selection", "clipboard")
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return "", err
		}
		return strings.TrimRight(out.String(), "\n\r"), nil
	}

	// Try wl-paste (Wayland).
	if _, err := exec.LookPath("wl-paste"); err == nil {
		cmd := exec.Command("wl-paste")
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return "", err
		}
		return strings.TrimRight(out.String(), "\n\r"), nil
	}

	// Try pbpaste (macOS).
	if _, err := exec.LookPath("pbpaste"); err == nil {
		cmd := exec.Command("pbpaste")
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return "", err
		}
		return strings.TrimRight(out.String(), "\n\r"), nil
	}

	return "", errors.New("clipboard: no clipboard tool found (install xclip, wl-paste, or pbpaste)")
}
