package main

import (
	"context"
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
		timeout      time.Duration
		noColor      bool
		noHyperlinks bool
	)

	addFlags := func(cmd *cobra.Command) {
		cmd.Flags().StringVar(&tmpl, "template", "", "Go template for output (default: built-in)")
		cmd.Flags().DurationVar(&cacheTTL, "cache", 30*time.Second, "Cache TTL (0 disables)")
		cmd.Flags().DurationVar(&timeout, "timeout", 3*time.Second, "Deadline for the GraphQL request before falling back to stale cache (0 disables)")
		cmd.Flags().BoolVar(&noColor, "no-color", false, "Strip ANSI color codes")
		cmd.Flags().BoolVar(&noHyperlinks, "no-hyperlinks", false, "Strip OSC 8 hyperlinks")
	}

	run := func(cmd *cobra.Command, args []string) error {
		mode := render.DetectMode(noColor, noHyperlinks)
		return runPR(tmpl, cacheTTL, timeout, mode)
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

	fieldsCmd := &cobra.Command{
		Use:   "fields",
		Short: "Print every template variable and its current value",
		Long: "Fetches the PR for the current branch and prints each template variable\n" +
			"alongside its rendered value. Useful for designing custom --template strings.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := render.DetectMode(noColor, noHyperlinks)
			return runFields(timeout, mode)
		},
	}
	fieldsCmd.Flags().DurationVar(&timeout, "timeout", 3*time.Second, "Deadline for the GraphQL request (0 disables)")
	fieldsCmd.Flags().BoolVar(&noColor, "no-color", false, "Strip ANSI color codes")
	fieldsCmd.Flags().BoolVar(&noHyperlinks, "no-hyperlinks", false, "Strip OSC 8 hyperlinks")

	rootCmd.AddCommand(prCmd, fieldsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runPR(tmpl string, ttl, timeout time.Duration, mode render.Mode) error {
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

	if pr.ShouldSkip(branch, defaultBranch()) {
		return emit(c, cwd, "", ttl)
	}

	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	state, err := pr.Fetch(ctx, repo.Owner, repo.Name, branch)
	if err != nil {
		// API failure (network, timeout, rate limit) → fall back to last known
		// good value, even if expired.
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

func runFields(timeout time.Duration, mode render.Mode) error {
	repo, err := repository.Current()
	if err != nil {
		return fmt.Errorf("not in a GitHub repo: %w", err)
	}
	branch, err := currentBranch()
	if err != nil {
		return fmt.Errorf("get branch: %w", err)
	}

	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	state, err := pr.Fetch(ctx, repo.Owner, repo.Name, branch)
	if err != nil {
		return fmt.Errorf("fetch PR: %w", err)
	}
	if state == nil || state.Number == 0 {
		fmt.Fprintf(os.Stderr, "no PR for branch %q — showing zero values\n\n", branch)
	}

	fields := render.Fields(state, mode)
	var maxName int
	for _, f := range fields {
		if f.Category == "helper" && len(f.Name) > maxName {
			maxName = len(f.Name)
		}
	}
	for _, f := range fields {
		if f.Category != "helper" {
			continue
		}
		fmt.Printf("%-*s  %s\n", maxName, f.Name, f.Value)
	}
	return nil
}

func currentBranch() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// defaultBranch reads refs/remotes/origin/HEAD. Returns "" if unset or git
// fails, which callers treat as "unknown, don't skip".
func defaultBranch() string {
	out, err := exec.Command("git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "origin/")
}
