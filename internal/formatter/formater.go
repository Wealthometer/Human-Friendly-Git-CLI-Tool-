package formatter

import (
	"fmt"
	"human-git-history/internal/git"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	yellow    = color.New(color.FgYellow).SprintFunc()
	green     = color.New(color.FgGreen).SprintFunc()
	cyan      = color.New(color.FgCyan).SprintFunc()
	red       = color.New(color.FgRed).SprintFunc()
	blue      = color.New(color.FgBlue).SprintFunc()
	magenta   = color.New(color.FgMagenta).SprintFunc()
	white     = color.New(color.FgWhite).SprintFunc()
	bold      = color.New(color.Bold).SprintFunc()
	dim       = color.New(color.Faint).SprintFunc()
	highlight = color.New(color.BgHiBlack, color.FgHiWhite).SprintFunc()
)

func PrintHumanFriendly(commits []git.Commit, compact bool, showStats bool, showGraph bool) {
	for i, commit := range commits {
		if showGraph {
			printGraphLine(i, len(commits))
		}
		
		printCommitHeader(commit, compact)
		
		if !compact && commit.Body != "" {
			printCommitBody(commit.Body)
		}
		
		if showStats && commit.Stats != nil {
			printCommitStats(*commit.Stats)
		}
		
		if !compact && len(commit.RefNames) > 0 {
			printRefNames(commit.RefNames)
		}
		
		if i < len(commits)-1 && !compact {
			fmt.Println(dim(strings.Repeat("─", 80)))
		}
	}
}

func PrintDetailed(commits []git.Commit, showStats bool, showGraph bool) {
	for i, commit := range commits {
		if showGraph {
			printGraphLine(i, len(commits))
		}
		
		fmt.Printf("%s %s\n", bold("Commit:"), highlight(commit.ShortHash))
		fmt.Printf("%s %s\n", bold("Hash:"), commit.Hash)
		fmt.Printf("%s %s <%s>\n", bold("Author:"), yellow(commit.AuthorName), commit.AuthorEmail)
		fmt.Printf("%s %s\n", bold("Date:"), formatDate(commit.AuthorDate))
		fmt.Printf("%s %s\n\n", bold("Message:"), white(commit.Message))
		
		if commit.Body != "" {
			fmt.Printf("%s\n%s\n\n", bold("Description:"), cyan(commit.Body))
		}
		
		if showStats && commit.Stats != nil {
			printCommitStats(*commit.Stats)
			fmt.Println()
		}
		
		if len(commit.RefNames) > 0 {
			printRefNames(commit.RefNames)
			fmt.Println()
		}
		
		if i < len(commits)-1 {
			fmt.Println(strings.Repeat("=", 80))
			fmt.Println()
		}
	}
}

func PrintCompact(commits []git.Commit) {
	for _, commit := range commits {
		timeAgo := formatTimeAgo(commit.AuthorDate)
		branchInfo := ""
		if len(commit.RefNames) > 0 {
			branchInfo = fmt.Sprintf(" [%s]", strings.Join(getBranchNames(commit.RefNames), ", "))
		}
		fmt.Printf("%s %s - %s (%s)%s\n",
			green(commit.ShortHash),
			white(commit.Message),
			yellow(commit.AuthorName),
			dim(timeAgo),
			magenta(branchInfo),
		)
	}
}

func PrintOneline(commits []git.Commit) {
	for _, commit := range commits {
		fmt.Printf("%s %s\n",
			green(commit.ShortHash),
			commit.Message,
		)
	}
}

func PrintChangelog(commits []git.Commit) {
	currentDate := ""
	for _, commit := range commits {
		commitDate := commit.AuthorDate.Format("2006-01-02")
		if commitDate != currentDate {
			currentDate = commitDate
			fmt.Printf("\n%s %s\n", bold("##"), formatDate(commit.AuthorDate))
		}
		fmt.Printf("- %s", commit.Message)
		
		if len(commit.RefNames) > 0 {
			fmt.Printf(" %s", magenta("["+strings.Join(getBranchNames(commit.RefNames), ", ")+"]"))
		}
		fmt.Printf(" %s\n", dim("("+commit.AuthorName+")"))
		
		if commit.Body != "" {
			lines := strings.Split(strings.TrimSpace(commit.Body), "\n")
			for _, line := range lines {
				if line != "" {
					fmt.Printf("  %s\n", dim("  "+line))
				}
			}
		}
	}
}

func printCommitHeader(commit git.Commit, compact bool) {
	timeAgo := formatTimeAgo(commit.AuthorDate)
	
	if compact {
		fmt.Printf("%s %s - %s (%s)\n",
			green(commit.ShortHash),
			white(commit.Message),
			yellow(commit.AuthorName),
			dim(timeAgo),
		)
	} else {
		fmt.Printf("%s %s\n", bold("commit"), highlight(commit.ShortHash))
		fmt.Printf("%s: %s <%s>\n", bold("Author"), yellow(commit.AuthorName), commit.AuthorEmail)
		fmt.Printf("%s: %s\n\n", bold("Date"), formatDate(commit.AuthorDate))
		fmt.Printf("    %s\n\n", white(commit.Message))
	}
}

func printCommitBody(body string) {
	lines := strings.Split(strings.TrimSpace(body), "\n")
	for _, line := range lines {
		if line != "" {
			fmt.Printf("    %s\n", cyan(line))
		}
	}
	fmt.Println()
}

func printCommitStats(stats git.CommitStats) {
	changeColor := green
	if stats.Deletions > stats.Insertions {
		changeColor = red
	}
	
	fmt.Printf("    %s: %d %s(+%d/-%d)\n",
		bold("Changes"),
		stats.FilesChanged,
		changeColor("█"),
		stats.Insertions,
		stats.Deletions,
	)
}

func printRefNames(refs []string) {
	fmt.Printf("    %s: ", bold("Refs"))
	for i, ref := range refs {
		if strings.HasPrefix(ref, "tag: ") {
			fmt.Print(blue(strings.TrimPrefix(ref, "tag: ")))
		} else if ref == "HEAD" {
			fmt.Print(red("HEAD"))
		} else if strings.HasPrefix(ref, "origin/") {
			fmt.Print(magenta(ref))
		} else {
			fmt.Print(green(ref))
		}
		if i < len(refs)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Println()
}

func printGraphLine(index, total int) {
	position := float64(index) / float64(total-1)
	width := 50
	
	bar := make([]rune, width)
	for i := 0; i < width; i++ {
		if float64(i)/float64(width) < position {
			bar[i] = '█'
		} else {
			bar[i] = '░'
		}
	}
	
	symbol := "●"
	if index == 0 {
		symbol = "⭓"
	} else if index == total-1 {
		symbol = "⭔"
	}
	
	fmt.Printf("%s %s\n", symbol, string(bar))
}

func formatDate(t time.Time) string {
	return t.Format("Mon, 02 Jan 2006 15:04:05 MST")
}

func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)
	
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d minute%s ago", minutes, pluralize(minutes))
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%d hour%s ago", hours, pluralize(hours))
	case diff < 30*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d day%s ago", days, pluralize(days))
	case diff < 365*24*time.Hour:
		months := int(diff.Hours() / (24 * 30))
		return fmt.Sprintf("%d month%s ago", months, pluralize(months))
	default:
		years := int(diff.Hours() / (24 * 365))
		return fmt.Sprintf("%d year%s ago", years, pluralize(years))
	}
}

func pluralize(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func getBranchNames(refs []string) []string {
	var branches []string
	for _, ref := range refs {
		if !strings.HasPrefix(ref, "tag: ") && ref != "HEAD" {
			branches = append(branches, ref)
		}
	}
	return branches
}