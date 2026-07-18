// Command ax is a TUI API Client — a lightweight, fast alternative to
// Postman/Insomnia that runs entirely in your terminal.
//
// Usage:
//
//	ax                  Launch the TUI
//	ax --version        Print version and exit
//	ax completion bash  Generate bash completion script
//	ax completion zsh   Generate zsh completion script
//	ax completion fish  Generate fish completion script
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/zaidejjo/ax/internal/tui"
)

// ─── Version ─────────────────────────────────────────────────────────────────

// version is set at build time via -ldflags (see Makefile and .goreleaser.yaml).
var version = "dev"

// ─── Entry Point ─────────────────────────────────────────────────────────────

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("ax version %s\n", version)
			return

		case "--help", "-h":
			printHelp()
			return

		case "completion":
			if len(os.Args) < 3 {
				printCompletionHelp()
				return
			}
			switch os.Args[2] {
			case "bash":
				fmt.Print(completionBash())
			case "zsh":
				fmt.Print(completionZsh())
			case "fish":
				fmt.Print(completionFish())
			default:
				fmt.Fprintf(os.Stderr, "Unknown shell %q. Supported: bash, zsh, fish\n", os.Args[2])
				os.Exit(1)
			}
			return
		}
	}

	// Create the root TUI model.
	m := tui.New(version)

	// Start the Bubble Tea program.
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// ─── Help ────────────────────────────────────────────────────────────────────

func printHelp() {
	fmt.Println("ax — TUI API Client")
	fmt.Println()
	fmt.Println("Usage: ax [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  (no command)     Launch the TUI")
	fmt.Println("  completion bash  Generate bash completion script")
	fmt.Println("  completion zsh   Generate zsh completion script")
	fmt.Println("  completion fish  Generate fish completion script")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --version, -v  Print version and exit")
	fmt.Println("  --help, -h     Show this help")
	fmt.Println()
	fmt.Println("TUI Keybindings:")
	fmt.Println("  Tab         Cycle focus between panes")
	fmt.Println("  Shift+Tab   Cycle focus backwards")
	fmt.Println("  Enter       Execute request (from URL input)")
	fmt.Println("  Ctrl+R      Execute the current request")
	fmt.Println("  m           Cycle HTTP method (outside URL input)")
	fmt.Println("  Ctrl+P      Paste from clipboard into URL")
	fmt.Println("  Ctrl+Y      Copy response body to clipboard")
	fmt.Println("  h           Toggle headers/body-only view (response)")
	fmt.Println("  Ctrl+H      Toggle sidebar visibility")
	fmt.Println("  ?           Toggle help overlay")
	fmt.Println("  q / Ctrl+C  Quit")
}

func printCompletionHelp() {
	fmt.Println("Generate shell completion scripts for ax.")
	fmt.Println()
	fmt.Println("Usage: ax completion <shell>")
	fmt.Println()
	fmt.Println("Supported shells: bash, zsh, fish")
	fmt.Println()
	fmt.Println("Installation:")
	fmt.Println("  Bash:  source <(ax completion bash)")
	fmt.Println("  Zsh:   source <(ax completion zsh)")
	fmt.Println("  Fish:  ax completion fish > ~/.config/fish/completions/ax.fish")
}

// ─── Bash Completion ─────────────────────────────────────────────────────────

func completionBash() string {
	return `# ax — bash completion
# Source: source <(ax completion bash)

_ax_completion() {
    local cur prev words cword
    _init_completion -n : || return

    # List of subcommands
    local commands="completion"

    # List of flags
    local flags="--version -v --help -h"

    case $prev in
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish" -- "$cur"))
            return 0
            ;;
    esac

    # No command yet: suggest commands and flags
    if [[ $cword -eq 1 ]]; then
        COMPREPLY=($(compgen -W "$commands $flags" -- "$cur"))
        return 0
    fi

    COMPREPLY=($(compgen -W "$flags" -- "$cur"))
}

complete -F _ax_completion ax
`
}

// ─── Zsh Completion ──────────────────────────────────────────────────────────

func completionZsh() string {
	return `# ax — zsh completion
# Source: source <(ax completion zsh)

#compdef ax

_ax() {
    local -a commands
    commands=(
        'completion:Generate shell completion scripts'
    )

    local -a flags
    flags=(
        '--version:Print version and exit'
        '-v:Print version and exit'
        '--help:Show this help'
        '-h:Show this help'
    )

    _arguments -C \
        '1: :->command' \
        '*: :->args' \
        && return 0

    case $state in
        command)
            _describe -t commands 'ax commands' commands
            _describe -t flags 'ax flags' flags
            ;;
        args)
            case $words[1] in
                completion)
                    _arguments '2:shell:(bash zsh fish)'
                    ;;
            esac
            ;;
    esac
}

_ax
`
}

// ─── Fish Completion ─────────────────────────────────────────────────────────

func completionFish() string {
	return `# ax — fish completion
# Install: ax completion fish > ~/.config/fish/completions/ax.fish

complete -c ax -f

# Subcommands
complete -c ax -n "not __fish_seen_subcommand_from completion" \
    -a "completion" -d "Generate shell completion scripts"

# completion subcommand arguments
complete -c ax -n "__fish_seen_subcommand_from completion" \
    -a "bash" -d "Generate bash completion"
complete -c ax -n "__fish_seen_subcommand_from completion" \
    -a "zsh" -d "Generate zsh completion"
complete -c ax -n "__fish_seen_subcommand_from completion" \
    -a "fish" -d "Generate fish completion"

# Flags
complete -c ax -l version -s v -d "Print version and exit"
complete -c ax -l help -s h -d "Show this help"
`
}

// ─── Completion Script Assembly Helpers ──────────────────────────────────────
// (Unused hint for the compiler — these functions are used above.)

var _ = strings.Builder{}
