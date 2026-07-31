// Package shell implements the shared --shell completions/init support
// required by AI.md PART 33 ("Shell Completions (Built-in, NON-NEGOTIABLE)")
// for every cassocial binary.
package shell

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Supported lists the shells --shell completions/init accept, in the order
// documented by AI.md PART 33.
var Supported = []string{"bash", "zsh", "fish", "sh", "dash", "ksh", "powershell", "pwsh"}

// Detect returns the caller's shell, derived from $SHELL, falling back to
// "bash" when it cannot be determined.
func Detect() string {
	sh := os.Getenv("SHELL")
	if sh == "" {
		return "bash"
	}
	name := filepath.Base(sh)
	for _, s := range Supported {
		if s == name {
			return s
		}
	}
	return "bash"
}

// isSupported reports whether shell is one of the Supported names.
func isSupported(sh string) bool {
	for _, s := range Supported {
		if s == sh {
			return true
		}
	}
	return false
}

// Handle implements the `--shell {completions|init} [SHELL]` subcommand
// shared by every binary. binary is the program name to embed in generated
// scripts (per AI.md PART 7 binary-naming rules, the caller passes the
// actual running binary name, not a hardcoded one). flags is the list of
// top-level long flags to offer as completion candidates. args is the
// remaining positional command line (i.e. fs.Args() after --shell was
// parsed) — args[0], if present, names the target shell; otherwise the
// shell is auto-detected. Output is written to w; the return value is the
// process exit code.
func Handle(binary, cmd string, args []string, flags []string, w io.Writer) int {
	sh := Detect()
	if len(args) > 0 && args[0] != "" {
		sh = strings.ToLower(args[0])
	}
	if !isSupported(sh) {
		fmt.Fprintf(w, "unsupported shell %q (supported: %s)\n", sh, strings.Join(Supported, ", "))
		return 1
	}

	switch cmd {
	case "completions":
		fmt.Fprint(w, Completions(binary, sh, flags))
		return 0
	case "init":
		fmt.Fprint(w, Init(binary, sh))
		return 0
	default:
		fmt.Fprintf(w, "unknown --shell command %q (expected completions|init)\n", cmd)
		return 1
	}
}

// Completions renders a completion script for the given shell, offering the
// supplied top-level long flags as candidates.
func Completions(binary, sh string, flags []string) string {
	switch sh {
	case "zsh":
		return zshCompletions(binary, flags)
	case "fish":
		return fishCompletions(binary, flags)
	case "sh", "dash", "ksh":
		return posixCompletions(binary, flags)
	case "powershell", "pwsh":
		return powershellCompletions(binary, flags)
	default: // bash
		return bashCompletions(binary, flags)
	}
}

// Init renders the eval-ready init command for the given shell (AI.md
// PART 33 "Shell Completions" — init is a convenience wrapper around
// completions, printed for the user to add to their shell rc file).
func Init(binary, sh string) string {
	switch sh {
	case "bash":
		return fmt.Sprintf("source <(%s --shell completions bash)\n", binary)
	case "zsh":
		return fmt.Sprintf("source <(%s --shell completions zsh)\n", binary)
	case "fish":
		return fmt.Sprintf("%s --shell completions fish | source\n", binary)
	case "powershell", "pwsh":
		return fmt.Sprintf("Invoke-Expression (& %s --shell completions %s)\n", binary, sh)
	default: // sh, dash, ksh
		return fmt.Sprintf("eval \"$(%s --shell completions %s)\"\n", binary, sh)
	}
}

func bashCompletions(binary string, flags []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# bash completion for %s\n", binary)
	fmt.Fprintf(&b, "_%s_completions() {\n", sanitize(binary))
	fmt.Fprintf(&b, "  local cur=\"${COMP_WORDS[COMP_CWORD]}\"\n")
	fmt.Fprintf(&b, "  COMPREPLY=( $(compgen -W \"%s\" -- \"$cur\") )\n", strings.Join(flags, " "))
	fmt.Fprintf(&b, "}\n")
	fmt.Fprintf(&b, "complete -F _%s_completions %s\n", sanitize(binary), binary)
	return b.String()
}

func zshCompletions(binary string, flags []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "#compdef %s\n", binary)
	fmt.Fprintf(&b, "_%s() {\n", sanitize(binary))
	fmt.Fprintf(&b, "  local -a opts\n")
	fmt.Fprintf(&b, "  opts=(%s)\n", strings.Join(flags, " "))
	fmt.Fprintf(&b, "  _describe 'option' opts\n")
	fmt.Fprintf(&b, "}\n")
	fmt.Fprintf(&b, "_%s\n", sanitize(binary))
	return b.String()
}

func fishCompletions(binary string, flags []string) string {
	var b strings.Builder
	for _, f := range flags {
		name := strings.TrimPrefix(f, "--")
		fmt.Fprintf(&b, "complete -c %s -l %s\n", binary, name)
	}
	return b.String()
}

func posixCompletions(binary string, flags []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# POSIX-shell completion helper for %s\n", binary)
	fmt.Fprintf(&b, "%s_OPTS=\"%s\"\n", strings.ToUpper(sanitize(binary)), strings.Join(flags, " "))
	return b.String()
}

func powershellCompletions(binary string, flags []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Register-ArgumentCompleter -Native -CommandName %s -ScriptBlock {\n", binary)
	fmt.Fprintf(&b, "  param($wordToComplete, $commandAst, $cursorPosition)\n")
	fmt.Fprintf(&b, "  @(%s) | Where-Object { $_ -like \"$wordToComplete*\" }\n", quotedList(flags))
	fmt.Fprintf(&b, "}\n")
	return b.String()
}

func quotedList(flags []string) string {
	quoted := make([]string, len(flags))
	for i, f := range flags {
		quoted[i] = "'" + f + "'"
	}
	return strings.Join(quoted, ", ")
}

// sanitize turns a binary name into a valid shell-function-name fragment.
func sanitize(binary string) string {
	return strings.NewReplacer("-", "_", ".", "_").Replace(binary)
}
