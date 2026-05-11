package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/justincampbell/gh-statusline/internal/cache"
	"github.com/justincampbell/gh-statusline/internal/pr"
	"github.com/justincampbell/gh-statusline/internal/render"
	"github.com/spf13/cobra"
)

func main() {
	var (
		tmpl         string
		cacheTTL     time.Duration
		noColor      bool
		noHyperlinks bool
	)

	addFlags := func(cmd *cobra.Command) {
		cmd.Flags().StringVar(&tmpl, "template", "", "Go template for output (default: built-in)")
		cmd.Flags().DurationVar(&cacheTTL, "cache", 30*time.Second, "Cache TTL (0 disables)")
		cmd.Flags().BoolVar(&noColor, "no-color", false, "Strip ANSI color codes")
		cmd.Flags().BoolVar(&noHyperlinks, "no-hyperlinks", false, "Strip OSC 8 hyperlinks")
	}

	run := func(cmd *cobra.Command, args []string) error {
		mode := render.DetectMode(noColor, noHyperlinks)
		return runPR(tmpl, cacheTTL, mode)
	}

	rootCmd := &cobra.Command{
		Use:   "statusline",
		Short: "Print a compact, colored status line for the current branch's PR",
		Long: "Prints a one-line summary of the GitHub PR for the current branch, suitable\n" +
			"for embedding in shell prompts, tmux statuslines, and Claude Code statuslines.\n" +
			"Output is cached briefly to keep latency off the GitHub API hot path.",
		SilenceUsage: true,
		RunE:         run,
	}
	addFlags(rootCmd)

	prCmd := &cobra.Command{
		Use:          "pr",
		Short:        "Status line for the current branch's PR (default)",
		SilenceUsage: true,
		RunE:         run,
	}
	addFlags(prCmd)

	rootCmd.AddCommand(prCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runPR(tmpl string, ttl time.Duration, mode render.Mode) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}

	c := cache.New()
	if ttl > 0 {
		if out, fresh, _ := c.Read(cwd, ttl); fresh {
			fmt.Println(out)
			return nil
		}
	}

	repo, err := repository.Current()
	if err != nil {
		// Not in a GitHub repo — silent empty output keeps prompts intact.
		return emit(c, cwd, "", ttl)
	}

	branch, err := currentBranch()
	if err != nil {
		return emit(c, cwd, "", ttl)
	}

	state, err := pr.Fetch(repo.Owner, repo.Name, branch)
	if err != nil {
		// API failure → fall back to last known good value, even if expired.
		if out, _, present := c.Read(cwd, 0); present {
			fmt.Println(out)
			return nil
		}
		return emit(c, cwd, "", ttl)
	}

	output, err := render.Render(state, mode, tmpl)
	if err != nil {
		return err
	}
	return emit(c, cwd, output, ttl)
}

func emit(c *cache.Cache, key, output string, ttl time.Duration) error {
	if ttl > 0 {
		c.Write(key, output)
	}
	fmt.Println(output)
	return nil
}

func currentBranch() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
