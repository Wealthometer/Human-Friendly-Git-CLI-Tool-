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

func PrintHumanFriendly(commits []git.Commit, compact bool, showStats bool, showGraph bool, showFiles bool) {
	for i, commit := range commits {
		if showGraph {
			printGraphLine(i, len(commits))
		}
		
		printCommitHeader(commit, compact)
		
		if !compact && commit.Body != "" {
			printCommitBody(commit.Body)
		}
		
		if showFiles && len(commit.FileChanges) > 0 {
			printFileChanges(commit.FileChanges)
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

func PrintDetailed(commits []git.Commit, showStats bool, showGraph bool, showFiles bool) {
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
		
		if showFiles && len(commit.FileChanges) > 0 {
			printDetailedFileChanges(commit.FileChanges)
			fmt.Println()
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

func PrintCompact(commits []git.Commit, showFiles bool) {
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
		
		if showFiles && len(commit.FileChanges) > 0 {
			printFileChangesCompact(commit.FileChanges)
		}
	}
}

func PrintOneline(commits []git.Commit, showFiles bool) {
	for _, commit := range commits {
		fmt.Printf("%s %s\n",
			green(commit.ShortHash),
			commit.Message,
		)
		
		if showFiles && len(commit.FileChanges) > 0 {
			for _, change := range commit.FileChanges {
				statusColor := getStatusColor(change.Status)
				fmt.Printf("  %s %s\n", statusColor(change.Status[:1]), change.FilePath)
			}
		}
	}
}

func PrintChangelog(commits []git.Commit, showFiles bool) {
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
		
		if showFiles && len(commit.FileChanges) > 0 {
			fmt.Println("  Changes:")
			for _, change := range commit.FileChanges {
				statusSymbol := getStatusSymbol(change.Status)
				statusColor := getStatusColor(change.Status)
				fmt.Printf("    %s %s", statusColor(statusSymbol), change.FilePath)
				if change.Insertions > 0 || change.Deletions > 0 {
					fmt.Printf(" (+%d/-%d)", change.Insertions, change.Deletions)
				}
				if change.OldPath != "" {
					fmt.Printf(" (from %s)", dim(change.OldPath))
				}
				fmt.Println()
			}
		}
		
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

func printFileChanges(changes []git.FileChange) {
	fmt.Printf("    %s\n", bold("Files:"))
	for _, change := range changes {
		statusColor := getStatusColor(change.Status)
		statusSymbol := getStatusSymbol(change.Status)
		
		fmt.Printf("    %s %s", statusColor(statusSymbol), change.FilePath)
		
		if change.Insertions > 0 || change.Deletions > 0 {
			fmt.Printf(" %s", dim(fmt.Sprintf("(+%d/-%d)", change.Insertions, change.Deletions)))
		}
		
		if change.OldPath != "" {
			fmt.Printf(" %s", dim(fmt.Sprintf("(renamed from %s)", change.OldPath)))
		}
		
		fmt.Println()
	}
}

func printDetailedFileChanges(changes []git.FileChange) {
	fmt.Printf("%s\n", bold("File Changes:"))
	
	added := []git.FileChange{}
	modified := []git.FileChange{}
	deleted := []git.FileChange{}
	renamed := []git.FileChange{}
	other := []git.FileChange{}
	
	for _, change := range changes {
		switch change.Status {
		case "Added":
			added = append(added, change)
		case "Modified":
			modified = append(modified, change)
		case "Deleted":
			deleted = append(deleted, change)
		case "Renamed":
			renamed = append(renamed, change)
		default:
			other = append(other, change)
		}
	}
	
	if len(added) > 0 {
		fmt.Printf("  %s:\n", green("Added"))
		for _, change := range added {
			fmt.Printf("    %s", change.FilePath)
			if change.Insertions > 0 {
				fmt.Printf(" %s", dim(fmt.Sprintf("(+%d lines)", change.Insertions)))
			}
			fmt.Println()
		}
	}
	
	if len(modified) > 0 {
		fmt.Printf("  %s:\n", yellow("Modified"))
		for _, change := range modified {
			fmt.Printf("    %s", change.FilePath)
			if change.Insertions > 0 || change.Deletions > 0 {
				fmt.Printf(" %s", dim(fmt.Sprintf("(+%d/-%d)", change.Insertions, change.Deletions)))
			}
			fmt.Println()
		}
	}
	
	if len(deleted) > 0 {
		fmt.Printf("  %s:\n", red("Deleted"))
		for _, change := range deleted {
			fmt.Printf("    %s", change.FilePath)
			if change.Deletions > 0 {
				fmt.Printf(" %s", dim(fmt.Sprintf("(-%d lines)", change.Deletions)))
			}
			fmt.Println()
		}
	}
	
	if len(renamed) > 0 {
		fmt.Printf("  %s:\n", cyan("Renamed"))
		for _, change := range renamed {
			fmt.Printf("    %s → %s", dim(change.OldPath), change.FilePath)
			fmt.Println()
		}
	}
	
	if len(other) > 0 {
		fmt.Printf("  %s:\n", magenta("Other"))
		for _, change := range other {
			fmt.Printf("    %s %s", change.Status, change.FilePath)
			fmt.Println()
		}
	}
}

func printFileChangesCompact(changes []git.FileChange) {
	for _, change := range changes {
		statusColor := getStatusColor(change.Status)
		statusSymbol := getStatusSymbol(change.Status)
		fmt.Printf("  %s %s", statusColor(statusSymbol), change.FilePath)
		if change.Insertions > 0 || change.Deletions > 0 {
			fmt.Printf(" %s", dim(fmt.Sprintf("(+%d/-%d)", change.Insertions, change.Deletions)))
		}
		fmt.Println()
	}
}

func getStatusColor(status string) func(...interface{}) string {
	switch status {
	case "Added":
		return green
	case "Modified":
		return yellow
	case "Deleted":
		return red
	case "Renamed", "Copied":
		return cyan
	default:
		return magenta
	}
}

func getStatusSymbol(status string) string {
	switch status {
	case "Added":
		return "A"
	case "Modified":
		return "M"
	case "Deleted":
		return "D"
	case "Renamed":
		return "R"
	case "Copied":
		return "C"
	default:
		return "•"
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