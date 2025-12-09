package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Commit struct {
	Hash         string
	ShortHash    string
	AuthorName   string
	AuthorEmail  string
	AuthorDate   time.Time
	Committer    string
	CommitDate   time.Time
	Message      string
	Body         string
	ParentHashes []string
	RefNames     []string
	Stats        *CommitStats
}

type CommitStats struct {
	FilesChanged int
	Insertions   int
	Deletions    int
}

type CommitOptions struct {
	Limit      int
	Author     string
	Since      string
	Until      string
	Branch     string
	MergesOnly bool
	NoMerges   bool
}

func GetCommits(options CommitOptions) ([]Commit, error) {
	args := []string{
		"log",
		"--pretty=format:%H|%h|%an|%ae|%ad|%cn|%cd|%s|%b|%P|%D",
		"--date=iso-strict",
		"--stat",
	}

	if options.Limit > 0 {
		args = append(args, fmt.Sprintf("--max-count=%d", options.Limit))
	}
	if options.Author != "" {
		args = append(args, fmt.Sprintf("--author=%s", options.Author))
	}
	if options.Since != "" {
		args = append(args, fmt.Sprintf("--since=%s", options.Since))
	}
	if options.Until != "" {
		args = append(args, fmt.Sprintf("--until=%s", options.Until))
	}
	if options.Branch != "" {
		args = append(args, options.Branch)
	}
	if options.MergesOnly {
		args = append(args, "--merges")
	}
	if options.NoMerges {
		args = append(args, "--no-merges")
	}

	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute git log: %v", err)
	}

	return parseGitLog(string(output))
}

func parseGitLog(output string) ([]Commit, error) {
	var commits []Commit
	scanner := bufio.NewScanner(strings.NewReader(output))
	
	var currentCommit *Commit
	var parsingStats bool
	var statLines []string

	for scanner.Scan() {
		line := scanner.Text()
		
		if strings.Contains(line, "|") && !parsingStats {
			if currentCommit != nil {
				if len(statLines) > 0 {
					stats := parseStats(statLines)
					currentCommit.Stats = stats
					statLines = []string{}
				}
				commits = append(commits, *currentCommit)
			}
			
			parts := strings.SplitN(line, "|", 11)
			if len(parts) < 11 {
				continue
			}
			
			authorDate, _ := time.Parse(time.RFC3339, parts[4])
			commitDate, _ := time.Parse(time.RFC3339, parts[6])
			
			currentCommit = &Commit{
				Hash:         parts[0],
				ShortHash:    parts[1],
				AuthorName:   parts[2],
				AuthorEmail:  parts[3],
				AuthorDate:   authorDate,
				Committer:    parts[5],
				CommitDate:   commitDate,
				Message:      parts[7],
				Body:         parts[8],
				ParentHashes: strings.Fields(parts[9]),
				RefNames:     parseRefNames(parts[10]),
			}
			parsingStats = false
		} else if line == "" {
			parsingStats = true
		} else if parsingStats && strings.Contains(line, "|") {
			statLines = append(statLines, line)
		}
	}
	
	if currentCommit != nil {
		if len(statLines) > 0 {
			stats := parseStats(statLines)
			currentCommit.Stats = stats
		}
		commits = append(commits, *currentCommit)
	}
	
	return commits, scanner.Err()
}

func parseRefNames(refStr string) []string {
	var refs []string
	for _, ref := range strings.Split(refStr, ", ") {
		ref = strings.TrimSpace(ref)
		if ref != "" {
			refs = append(refs, ref)
		}
	}
	return refs
}

func parseStats(statLines []string) *CommitStats {
	stats := &CommitStats{}
	
	for _, line := range statLines {
		if strings.Contains(line, "|") {
			stats.FilesChanged++
			parts := strings.Split(line, "|")
			if len(parts) > 1 {
				changes := strings.TrimSpace(parts[1])
				if strings.Contains(changes, "+") && strings.Contains(changes, "-") {
					var insertions, deletions int
					fmt.Sscanf(changes, "%d insertions(+), %d deletions(-)", &insertions, &deletions)
					stats.Insertions += insertions
					stats.Deletions += deletions
				}
			}
		} else if strings.Contains(line, "insertion") || strings.Contains(line, "deletion") {
			var insertions, deletions int
			fmt.Sscanf(line, "%d files changed, %d insertions(+), %d deletions(-)", 
				&stats.FilesChanged, &insertions, &deletions)
			stats.Insertions = insertions
			stats.Deletions = deletions
		}
	}
	
	return stats
}