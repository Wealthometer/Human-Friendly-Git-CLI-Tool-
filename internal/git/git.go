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
	var statLines []string
	var inStats bool
	var bodyLines []string

	for scanner.Scan() {
		line := scanner.Text()
		
		// Check if this is a commit header line
		if strings.Count(line, "|") >= 10 {
			// Save previous commit if exists
			if currentCommit != nil {
				// Process stats for previous commit
				if len(statLines) > 0 {
					stats := parseStats(statLines)
					currentCommit.Stats = stats
				}
				// Join body lines
				if len(bodyLines) > 0 {
					currentCommit.Body = strings.Join(bodyLines, "\n")
				}
				commits = append(commits, *currentCommit)
			}
			
			// Parse new commit
			parts := strings.SplitN(line, "|", 11)
			
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
				Message:      strings.TrimSpace(parts[7]),
				Body:         "",
				ParentHashes: strings.Fields(parts[9]),
				RefNames:     parseRefNames(parts[10]),
			}
			
			// Reset for new commit
			bodyLines = []string{}
			statLines = []string{}
			inStats = false
			
			// Add initial body if present
			if parts[8] != "" {
				bodyLines = append(bodyLines, strings.TrimSpace(parts[8]))
			}
			
		} else if line == "" {
			// Empty line could be end of body or start of stats
			if !inStats && len(bodyLines) > 0 {
				// This empty line might separate body from stats
				inStats = true
			}
		} else if inStats {
			// Collect stat lines
			statLines = append(statLines, line)
		} else if currentCommit != nil {
			// This is part of the commit body
			bodyLines = append(bodyLines, line)
		}
	}
	
	// Don't forget the last commit
	if currentCommit != nil {
		if len(statLines) > 0 {
			stats := parseStats(statLines)
			currentCommit.Stats = stats
		}
		if len(bodyLines) > 0 {
			currentCommit.Body = strings.Join(bodyLines, "\n")
		}
		commits = append(commits, *currentCommit)
	}
	
	return commits, scanner.Err()
}

func parseRefNames(refStr string) []string {
	var refs []string
	if refStr == "" {
		return refs
	}
	
	// Split by comma and trim spaces
	for _, ref := range strings.Split(refStr, ",") {
		ref = strings.TrimSpace(ref)
		if ref != "" {
			refs = append(refs, ref)
		}
	}
	return refs
}

func parseStats(statLines []string) *CommitStats {
	stats := &CommitStats{}
	
	// Look for the summary line
	for _, line := range statLines {
		if strings.Contains(line, "files changed") || 
		   strings.Contains(line, "insertion") || 
		   strings.Contains(line, "deletion") {
			// Parse summary line like: "2 files changed, 15 insertions(+), 3 deletions(-)"
			var files, insertions, deletions int
			n, _ := fmt.Sscanf(line, "%d files changed, %d insertions(+), %d deletions(-)", 
				&files, &insertions, &deletions)
			if n >= 1 {
				stats.FilesChanged = files
			}
			if n >= 2 {
				stats.Insertions = insertions
			}
			if n >= 3 {
				stats.Deletions = deletions
			}
			break
		}
	}
	
	return stats
}