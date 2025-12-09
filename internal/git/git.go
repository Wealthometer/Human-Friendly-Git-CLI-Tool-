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
	FileChanges  []FileChange // New field for detailed file changes
}

type FileChange struct {
	Status   string // Added, Modified, Deleted, Renamed, Copied
	FilePath string
	OldPath  string // For renames/copies
	Insertions int
	Deletions  int
}

type CommitStats struct {
	FilesChanged int
	Insertions   int
	Deletions    int
}

type CommitOptions struct {
	Limit          int
	Author         string
	Since          string
	Until          string
	Branch         string
	MergesOnly     bool
	NoMerges       bool
	ShowFileChanges bool // New option
}

func GetCommits(options CommitOptions) ([]Commit, error) {
	args := []string{
		"log",
		"--pretty=format:%H|%h|%an|%ae|%ad|%cn|%cd|%s|%b|%P|%D",
		"--date=iso-strict",
		"--stat",
	}

	if options.ShowFileChanges {
		args = append(args, "--name-status") // Show detailed file changes
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

	commits, err := parseGitLog(string(output), options.ShowFileChanges)
	if err != nil {
		return nil, err
	}

	// Get detailed diff stats for each commit if requested
	if options.ShowFileChanges {
		for i := range commits {
			detailedStats, err := getDetailedDiffStats(commits[i].Hash)
			if err == nil {
				commits[i].FileChanges = detailedStats
			}
		}
	}

	return commits, nil
}

func parseGitLog(output string, showFileChanges bool) ([]Commit, error) {
	var commits []Commit
	scanner := bufio.NewScanner(strings.NewReader(output))
	
	var currentCommit *Commit
	var statLines []string
	var nameStatusLines []string
	var inStats bool
	var inNameStatus bool
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
				// Parse name-status lines
				if len(nameStatusLines) > 0 && showFileChanges {
					currentCommit.FileChanges = parseNameStatus(nameStatusLines)
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
			nameStatusLines = []string{}
			inStats = false
			inNameStatus = false
			
			// Add initial body if present
			if parts[8] != "" {
				bodyLines = append(bodyLines, strings.TrimSpace(parts[8]))
			}
			
		} else if line == "" {
			// Empty line could be end of body or start of stats/name-status
			if !inStats && len(bodyLines) > 0 {
				// This empty line might separate body from stats
				inStats = true
				if showFileChanges {
					inNameStatus = true
				}
			}
		} else if inNameStatus && !strings.Contains(line, "|") && 
		          (strings.Contains(line, "files changed") || 
		           strings.Contains(line, "insertion") || 
		           strings.Contains(line, "deletion")) {
			// This is a stat summary line, not a name-status line
			statLines = append(statLines, line)
			inNameStatus = false
		} else if inNameStatus && len(strings.Fields(line)) >= 2 {
			// Collect name-status lines (format: "M\tfile.go" or "R100\told.txt\tnew.txt")
			nameStatusLines = append(nameStatusLines, line)
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
		if len(nameStatusLines) > 0 && showFileChanges {
			currentCommit.FileChanges = parseNameStatus(nameStatusLines)
		}
		if len(bodyLines) > 0 {
			currentCommit.Body = strings.Join(bodyLines, "\n")
		}
		commits = append(commits, *currentCommit)
	}
	
	return commits, scanner.Err()
}

func parseNameStatus(lines []string) []FileChange {
	var changes []FileChange
	
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		
		status := fields[0]
		change := FileChange{Status: getStatusSymbol(status)}
		
		switch {
		case status[0] == 'R' || status[0] == 'C': // Renamed or Copied
			if len(fields) >= 3 {
				change.OldPath = fields[1]
				change.FilePath = fields[2]
			}
		default: // Added, Modified, Deleted
			change.FilePath = fields[1]
		}
		
		changes = append(changes, change)
	}
	
	return changes
}

func getStatusSymbol(status string) string {
	switch status[0] {
	case 'A':
		return "Added"
	case 'M':
		return "Modified"
	case 'D':
		return "Deleted"
	case 'R':
		return "Renamed"
	case 'C':
		return "Copied"
	case 'T':
		return "Type Changed"
	default:
		return "Changed"
	}
}

func getDetailedDiffStats(hash string) ([]FileChange, error) {
	// Use git show with --numstat for detailed per-file stats
	cmd := exec.Command("git", "show", "--numstat", "--pretty=format:", hash)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	return parseNumStatOutput(string(output)), nil
}

func parseNumStatOutput(output string) []FileChange {
	var changes []FileChange
	scanner := bufio.NewScanner(strings.NewReader(output))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			var insertions, deletions int
			fmt.Sscanf(fields[0], "%d", &insertions)
			fmt.Sscanf(fields[1], "%d", &deletions)
			
			// Determine status based on insertions/deletions
			status := "Modified"
			if insertions > 0 && deletions == 0 {
				status = "Added"
			} else if insertions == 0 && deletions > 0 {
				status = "Deleted"
			}
			
			change := FileChange{
				Status:     status,
				FilePath:   fields[2],
				Insertions: insertions,
				Deletions:  deletions,
			}
			
			changes = append(changes, change)
		}
	}
	
	return changes
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