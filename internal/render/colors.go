package render

import (
	"fmt"
	"os"
)

const (
	ansiReset = "\033[0m"
	ansiDim   = "\033[2m"
)

// Mode controls how strings are rendered.
type Mode struct {
	NoColor      bool
	NoHyperlinks bool
}

// DetectMode returns a Mode reflecting the user's preferences. Colors and
// hyperlinks are emitted by default — the typical consumer (a shell prompt,
// tmux statusline, or Claude Code statusline) captures stdout and is not a
// TTY, yet expects ANSI escapes. Opt out with --no-color/--no-hyperlinks or
// the standard NO_COLOR env var.
func DetectMode(noColorFlag, noHyperlinksFlag bool) Mode {
	_, noColorEnv := os.LookupEnv("NO_COLOR")
	return Mode{
		NoColor:      noColorFlag || noColorEnv,
		NoHyperlinks: noHyperlinksFlag,
	}
}

func (m Mode) color(code, text string) string {
	if m.NoColor || text == "" {
		return text
	}
	return code + text + ansiReset
}

func (m Mode) dim(text string) string  { return m.color(ansiDim, text) }
func (m Mode) green(text string) string { return m.color("\033[32m", text) }
func (m Mode) red(text string) string   { return m.color("\033[31m", text) }
func (m Mode) yellow(text string) string { return m.color("\033[33m", text) }
func (m Mode) gray(text string) string  { return m.color("\033[90m", text) }
func (m Mode) magenta(text string) string { return m.color("\033[35m", text) }
func (m Mode) cyan(text string) string   { return m.color("\033[36m", text) }

func (m Mode) hex(text, hexColor string) string {
	if m.NoColor || text == "" {
		return text
	}
	if len(hexColor) != 6 {
		return m.dim(text)
	}
	var r, g, b int
	fmt.Sscanf(hexColor, "%02x%02x%02x", &r, &g, &b)
	return fmt.Sprintf("\033[38;2;%d;%d;%dm%s%s", r, g, b, text, ansiReset)
}

// hyperlink wraps text in an OSC 8 terminal hyperlink unless disabled.
func (m Mode) hyperlink(text, url string) string {
	if m.NoHyperlinks || url == "" {
		return text
	}
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}
