package cmd

import (
	"fmt"
	"os"
	"human-git-history/internal/formatter"
	"human-git-history/internal/git"

	"github.com/spf13/cobra"
)

var (
	limit      int
	author     string
	since      string
	until      string
	branch     string
	format     string
	compact    bool
	showStats  bool
	showFiles  bool // New flag
	graph      bool
	mergesOnly bool
	noMerges   bool
)

var rootCmd = &cobra.Command{
	Use:   "git-history",
	Short: "A human-friendly git history viewer",
	Long: `A CLI tool that presents git history in a more readable,
human-friendly format with various display options.`,
	Run: func(cmd *cobra.Command, args []string) {
		commits, err := git.GetCommits(git.CommitOptions{
			Limit:          limit,
			Author:         author,
			Since:          since,
			Until:          until,
			Branch:         branch,
			MergesOnly:     mergesOnly,
			NoMerges:       noMerges,
			ShowFileChanges: showFiles,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting commits: %v\n", err)
			os.Exit(1)
		}

		switch format {
		case "detailed":
			formatter.PrintDetailed(commits, showStats, graph, showFiles)
		case "compact":
			formatter.PrintCompact(commits, showFiles)
		case "oneline":
			formatter.PrintOneline(commits, showFiles)
		case "changelog":
			formatter.PrintChangelog(commits, showFiles)
		default:
			formatter.PrintHumanFriendly(commits, compact, showStats, graph, showFiles)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&limit, "limit", "n", 50, "Limit number of commits")
	rootCmd.PersistentFlags().StringVarP(&author, "author", "a", "", "Filter by author")
	rootCmd.PersistentFlags().StringVar(&since, "since", "", "Show commits more recent than specific date")
	rootCmd.PersistentFlags().StringVar(&until, "until", "", "Show commits older than specific date")
	rootCmd.PersistentFlags().StringVarP(&branch, "branch", "b", "", "Show commits from specific branch")
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "", "Output format (detailed, compact, oneline, changelog)")
	rootCmd.PersistentFlags().BoolVarP(&compact, "compact", "c", false, "Compact output")
	rootCmd.PersistentFlags().BoolVar(&showStats, "stats", false, "Show file statistics")
	rootCmd.PersistentFlags().BoolVar(&showFiles, "files", false, "Show changed files with details") // New flag
	rootCmd.PersistentFlags().BoolVar(&graph, "graph", false, "Show ASCII commit graph")
	rootCmd.PersistentFlags().BoolVar(&mergesOnly, "merges", false, "Show only merge commits")
	rootCmd.PersistentFlags().BoolVar(&noMerges, "no-merges", false, "Exclude merge commits")
	
	rootCmd.MarkFlagsMutuallyExclusive("merges", "no-merges")
}