package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"human-git-history/internal/git"
	"human-git-history/internal/template"

	"github.com/spf13/cobra"
)

import "runtime"


var (
	outputFile   string
	title        string
	description  string
	groupByDate  bool
	groupByAuthor bool
	theme        string
	openBrowser  bool
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Generate HTML webpage from git history",
	Long:  `Generate a beautifully formatted HTML webpage displaying git history with interactive features.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get commits
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

		// Calculate repository statistics
		stats := calculateRepoStats(commits)

		// Initialize template renderer
		templateDir := "templates"
		assetDir := "assets"
		
		// Check if templates exist, create default if not
		if _, err := os.Stat(templateDir); os.IsNotExist(err) {
			fmt.Println("Creating default templates...")
			if err := createDefaultTemplates(templateDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating templates: %v\n", err)
				os.Exit(1)
			}
		}

		// Check if assets exist, create default if not
		if _, err := os.Stat(assetDir); os.IsNotExist(err) {
			fmt.Println("Creating default assets...")
			if err := createDefaultAssets(assetDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating assets: %v\n", err)
				os.Exit(1)
			}
		}

		renderer, err := template.NewRenderer(templateDir, assetDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing renderer: %v\n", err)
			os.Exit(1)
		}

		// Prepare template data
		data := template.TemplateData{
			Commits:     commits,
			Title:       getRepoTitle(title),
			Description: getRepoDescription(description),
			GeneratedAt: time.Now(),
			Stats:       stats,
			Options: template.RenderOptions{
				ShowFiles:     showFiles,
				ShowStats:     showStats,
				GroupByDate:   groupByDate,
				GroupByAuthor: groupByAuthor,
				Since:         since,
				Until:         until,
				Author:        author,
				Theme:         theme,
				CompactView:   compact,
			},
		}

		// Determine output file
		if outputFile == "" {
			outputFile = "git-history.html"
		}

		// Ensure output directory exists
		outputDir := filepath.Dir(outputFile)
		if outputDir != "." {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
				os.Exit(1)
			}
		}

		// Render to file
		fmt.Printf("Generating HTML to %s...\n", outputFile)
		if err := renderer.RenderToFile(outputFile, data); err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering HTML: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Successfully generated %s\n", outputFile)
		
		// Open in browser if requested
		if openBrowser {
			openInBrowser(outputFile)
		} else {
			fmt.Printf("📂 Open %s in your browser to view the git history.\n", outputFile)
		}
	},
}

func calculateRepoStats(commits []git.Commit) *template.RepoStats {
	if len(commits) == 0 {
		return &template.RepoStats{}
	}

	stats := &template.RepoStats{
		TotalCommits:   len(commits),
		Authors:        make(map[string]template.AuthorStats),
		FirstCommit:    commits[len(commits)-1].AuthorDate,
		LastCommit:     commits[0].AuthorDate,
	}

	authorsMap := make(map[string]bool)
	
	for _, commit := range commits {
		// Track unique authors
		authorKey := commit.AuthorName + "|" + commit.AuthorEmail
		if !authorsMap[authorKey] {
			authorsMap[authorKey] = true
		}

		// Update author stats
		if _, exists := stats.Authors[authorKey]; !exists {
			stats.Authors[authorKey] = template.AuthorStats{
				Name:    commit.AuthorName,
				Email:   commit.AuthorEmail,
				Commits: 0,
			}
		}
		
		authorStat := stats.Authors[authorKey]
		authorStat.Commits++
		if commit.Stats != nil {
			authorStat.Insertions += commit.Stats.Insertions
			authorStat.Deletions += commit.Stats.Deletions
			stats.FilesChanged += commit.Stats.FilesChanged
			stats.TotalInsertions += commit.Stats.Insertions
			stats.TotalDeletions += commit.Stats.Deletions
		}
		stats.Authors[authorKey] = authorStat
	}

	stats.TotalAuthors = len(authorsMap)
	return stats
}

func getRepoTitle(defaultTitle string) string {
	if defaultTitle != "" {
		return defaultTitle
	}
	
	// Try to get repo name from git
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		repoPath := strings.TrimSpace(string(output))
		repoName := filepath.Base(repoPath)
		return fmt.Sprintf("%s - Git History", repoName)
	}
	
	// Try git config
	cmd = exec.Command("git", "config", "--get", "remote.origin.url")
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		url := strings.TrimSpace(string(output))
		if repoName := extractRepoName(url); repoName != "" {
			return fmt.Sprintf("%s - Git History", repoName)
		}
	}
	
	return "Git Repository History"
}

func getRepoDescription(defaultDesc string) string {
	if defaultDesc != "" {
		return defaultDesc
	}
	
	// Try to get repo description
	cmd := exec.Command("git", "config", "--get", "gitweb.description")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output))
	}
	
	// Try README first line
	if content, err := os.ReadFile("README.md"); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				return trimmed
			}
		}
	}
	
	return "Interactive visualization of git commit history"
}

func extractRepoName(url string) string {
	// Extract repo name from git URL
	// Handle various formats: git@github.com:user/repo.git, https://github.com/user/repo.git
	url = strings.TrimSpace(url)
	
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")
	
	// Handle SSH format
	if strings.Contains(url, "git@") {
		parts := strings.Split(url, ":")
		if len(parts) > 1 {
			urlParts := strings.Split(parts[1], "/")
			if len(urlParts) > 0 {
				return urlParts[len(urlParts)-1]
			}
		}
	}
	
	// Handle HTTPS format
	if strings.Contains(url, "http") {
		parts := strings.Split(url, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	
	return ""
}

func createDefaultTemplates(templateDir string) error {
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		return err
	}
	
	// Create index.tpl
	indexTPL := `<!DOCTYPE html>
<html lang="en" data-theme="{{.Options.Theme | default "auto"}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Git History</title>
    <meta name="description" content="{{.Description}}">
    <link rel="stylesheet" href="assets/style.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        :root {
            --primary-color: #3498db;
            --secondary-color: #2ecc71;
            --danger-color: #e74c3c;
            --warning-color: #f39c12;
            --info-color: #9b59b6;
            --bg-color: #ffffff;
            --text-color: #333333;
            --border-color: #e0e0e0;
            --card-bg: #f8f9fa;
        }
        [data-theme="dark"] {
            --bg-color: #1a1a1a;
            --text-color: #ffffff;
            --border-color: #404040;
            --card-bg: #2d2d2d;
        }
    </style>
</head>
<body>
    <div class="container">
        <!-- Header -->
        <header class="header">
            <h1><i class="fas fa-history"></i> {{.Title}}</h1>
            <p class="subtitle">{{.Description}}</p>
            <div class="meta">
                <span><i class="fas fa-calendar"></i> Generated: {{.GeneratedAt | formatDateTime}}</span>
                <span><i class="fas fa-code-branch"></i> Commits: {{.Stats.TotalCommits}}</span>
                <span><i class="fas fa-users"></i> Authors: {{.Stats.TotalAuthors}}</span>
            </div>
        </header>

        <!-- Controls -->
        <div class="controls">
            <button class="btn" onclick="toggleTheme()">
                <i class="fas fa-moon"></i> Toggle Theme
            </button>
            <button class="btn" onclick="toggleStats()">
                <i class="fas fa-chart-bar"></i> Toggle Stats
            </button>
            <div class="search">
                <input type="text" id="searchInput" placeholder="Search commits..." onkeyup="searchCommits()">
                <button class="btn" onclick="searchCommits()">
                    <i class="fas fa-search"></i>
                </button>
            </div>
        </div>

        <!-- Stats Panel -->
        <div id="statsPanel" class="stats-panel">
            {{template "stats.tpl" .Stats}}
        </div>

        <!-- Filter Bar -->
        <div class="filters">
            <div class="filter-group">
                <label for="authorFilter">Author:</label>
                <select id="authorFilter" onchange="filterByAuthor()">
                    <option value="">All Authors</option>
                    {{range $author, $stats := .Stats.Authors}}
                    <option value="{{$stats.Name}}">{{$stats.Name}} ({{$stats.Commits}})</option>
                    {{end}}
                </select>
            </div>
            <div class="filter-group">
                <label for="dateFilter">Date Range:</label>
                <input type="date" id="dateFrom" onchange="filterByDate()">
                <input type="date" id="dateTo" onchange="filterByDate()">
            </div>
        </div>

        <!-- Commit List -->
        <div class="commit-list">
            {{if .Options.GroupByDate}}
                {{$currentDate := ""}}
                {{range .Commits}}
                    {{$commitDate := .AuthorDate | formatDate}}
                    {{if ne $commitDate $currentDate}}
                        {{$currentDate = $commitDate}}
                        <div class="date-header">
                            <h2><i class="fas fa-calendar-day"></i> {{$currentDate}}</h2>
                        </div>
                    {{end}}
                    {{template "commit.tpl" (dict "Commit" . "Options" $.Options)}}
                {{end}}
            {{else if .Options.GroupByAuthor}}
                {{$currentAuthor := ""}}
                {{range .Commits}}
                    {{if ne .AuthorName $currentAuthor}}
                        {{$currentAuthor = .AuthorName}}
                        <div class="author-header">
                            <h2><i class="fas fa-user"></i> {{$currentAuthor}}</h2>
                        </div>
                    {{end}}
                    {{template "commit.tpl" (dict "Commit" . "Options" $.Options)}}
                {{end}}
            {{else}}
                {{range .Commits}}
                    {{template "commit.tpl" (dict "Commit" . "Options" $.Options)}}
                {{end}}
            {{end}}
        </div>

        <!-- Footer -->
        <footer class="footer">
            <p>
                Generated with <i class="fas fa-heart" style="color: #e74c3c;"></i> 
                by Human Git History Tool
            </p>
            <p class="footer-links">
                <a href="#" onclick="scrollToTop()"><i class="fas fa-arrow-up"></i> Back to Top</a>
            </p>
        </footer>
    </div>

    <script>
        // Theme handling
        function toggleTheme() {
            const html = document.documentElement;
            const currentTheme = html.getAttribute('data-theme');
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            html.setAttribute('data-theme', newTheme);
            localStorage.setItem('theme', newTheme);
        }

        // Load saved theme
        const savedTheme = localStorage.getItem('theme') || 'auto';
        document.documentElement.setAttribute('data-theme', savedTheme);

        // Stats panel toggle
        function toggleStats() {
            const panel = document.getElementById('statsPanel');
            panel.classList.toggle('hidden');
        }

        // Search functionality
        function searchCommits() {
            const searchTerm = document.getElementById('searchInput').value.toLowerCase();
            const commits = document.querySelectorAll('.commit-card');
            
            commits.forEach(commit => {
                const text = commit.textContent.toLowerCase();
                commit.style.display = text.includes(searchTerm) ? 'block' : 'none';
            });
        }

        // Filter by author
        function filterByAuthor() {
            const author = document.getElementById('authorFilter').value;
            filterCommits('data-author', author);
        }

        // Filter by date
        function filterByDate() {
            const from = document.getElementById('dateFrom').value;
            const to = document.getElementById('dateTo').value;
            
            if (!from && !to) {
                showAllCommits();
                return;
            }
            
            const commits = document.querySelectorAll('.commit-card');
            commits.forEach(commit => {
                const commitDate = commit.getAttribute('data-date');
                const show = (!from || commitDate >= from) && (!to || commitDate <= to);
                commit.style.display = show ? 'block' : 'none';
            });
        }

        function filterCommits(attribute, value) {
            const commits = document.querySelectorAll('.commit-card');
            commits.forEach(commit => {
                if (!value || commit.getAttribute(attribute) === value) {
                    commit.style.display = 'block';
                } else {
                    commit.style.display = 'none';
                }
            });
        }

        function showAllCommits() {
            document.querySelectorAll('.commit-card').forEach(c => c.style.display = 'block');
        }

        // Scroll to top
        function scrollToTop() {
            window.scrollTo({ top: 0, behavior: 'smooth' });
        }

        // Initialize
        document.addEventListener('DOMContentLoaded', function() {
            // Set today's date as default for dateTo
            const today = new Date().toISOString().split('T')[0];
            document.getElementById('dateTo').value = today;
        });
    </script>
</body>
</html>`
	
	if err := os.WriteFile(filepath.Join(templateDir, "index.tpl"), []byte(indexTPL), 0644); err != nil {
		return err
	}
	
	// Create commit.tpl (simplified version)
	commitTPL := `<div class="commit-card" 
     data-author="{{.Commit.AuthorName}}"
     data-date="{{.Commit.AuthorDate | formatDate}}">
    
    <div class="commit-header">
        <div class="commit-hash">
            <i class="fas fa-code-commit"></i>
            <span class="hash-link" title="{{.Commit.Hash}}">
                {{.Commit.ShortHash | shortHash}}
            </span>
        </div>
        <div class="commit-meta">
            <span class="author">
                <i class="fas fa-user"></i>
                {{.Commit.AuthorName}}
            </span>
            <span class="date">
                <i class="fas fa-clock"></i>
                {{.Commit.AuthorDate | formatTimeAgo}}
            </span>
        </div>
    </div>

    <div class="commit-message">
        <h3>{{.Commit.Message}}</h3>
        {{if .Commit.Body}}
        <div class="commit-body">
            <p>{{.Commit.Body}}</p>
        </div>
        {{end}}
    </div>

    {{if .Options.ShowFiles}}
    <div class="file-changes">
        <h4><i class="fas fa-file-alt"></i> Changed Files ({{len .Commit.FileChanges}})</h4>
        <div class="file-list">
            {{range .Commit.FileChanges}}
            <div class="file-item status-{{.Status | fileStatusColor}}">
                <span class="file-status">
                    {{.Status | fileStatusIcon}} {{.Status | fileStatusText}}
                </span>
                <span class="file-path">{{.FilePath}}</span>
                {{if .Insertions .Deletions}}
                <span class="file-stats">
                    <span class="insertions">+{{.Insertions}}</span>
                    <span class="deletions">-{{.Deletions}}</span>
                </span>
                {{end}}
            </div>
            {{end}}
        </div>
    </div>
    {{end}}

    <div class="commit-footer">
        <div class="commit-refs">
            {{if .Commit.RefNames}}
            <span class="refs-label">Refs:</span>
            {{range .Commit.RefNames}}
            <span class="ref {{if hasPrefix . "tag:"}}tag{{else if eq . "HEAD"}}head{{else}}branch{{end}}">
                {{if hasPrefix . "tag:"}}
                <i class="fas fa-tag"></i>
                {{replace . "tag: " ""}}
                {{else if eq . "HEAD"}}
                <i class="fas fa-code-branch"></i> HEAD
                {{else}}
                <i class="fas fa-code-branch"></i>
                {{.}}
                {{end}}
            </span>
            {{end}}
            {{end}}
        </div>
    </div>
</div>`
	
	if err := os.WriteFile(filepath.Join(templateDir, "commit.tpl"), []byte(commitTPL), 0644); err != nil {
		return err
	}
	
	// Create stats.tpl (simplified)
	statsTPL := `<div class="stats-container">
    <h2><i class="fas fa-chart-bar"></i> Repository Statistics</h2>
    
    <div class="stats-grid">
        <div class="stat-card">
            <div class="stat-icon total">
                <i class="fas fa-code-commit"></i>
            </div>
            <div class="stat-content">
                <h3>{{.TotalCommits}}</h3>
                <p>Total Commits</p>
            </div>
        </div>
        
        <div class="stat-card">
            <div class="stat-icon authors">
                <i class="fas fa-users"></i>
            </div>
            <div class="stat-content">
                <h3>{{.TotalAuthors}}</h3>
                <p>Contributors</p>
            </div>
        </div>
        
        <div class="stat-card">
            <div class="stat-icon lines">
                <i class="fas fa-code"></i>
            </div>
            <div class="stat-content">
                <h3>{{add .TotalInsertions .TotalDeletions}}</h3>
                <p>Lines of Code</p>
                <small>(+{{.TotalInsertions}}/-{{.TotalDeletions}})</small>
            </div>
        </div>
    </div>

    <div class="stat-section">
        <h3><i class="fas fa-trophy"></i> Top Contributors</h3>
        <div class="authors-list">
            {{range $author, $stats := .Authors}}
            <div class="author-item">
                <div class="author-info">
                    <span class="author-name">{{$stats.Name}}</span>
                    <span class="author-email">{{$stats.Email}}</span>
                </div>
                <div class="author-stats">
                    <span class="stat commits">{{$stats.Commits}} commits</span>
                </div>
                <div class="author-progress">
                    <div class="progress-bar" style="width: {{percentage $stats.Commits $.TotalCommits}}%"></div>
                </div>
            </div>
            {{end}}
        </div>
    </div>
</div>`
	
	if err := os.WriteFile(filepath.Join(templateDir, "stats.tpl"), []byte(statsTPL), 0644); err != nil {
		return err
	}
	
	return nil
}

func createDefaultAssets(assetDir string) error {
	if err := os.MkdirAll(assetDir, 0755); err != nil {
		return err
	}
	
	

	// For brevity, let me provide a minimal version
	minimalCSS := `* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; color: #333; line-height: 1.6; }
.container { max-width: 1200px; margin: 0 auto; padding: 20px; }
.header { background: white; padding: 30px; border-radius: 12px; margin-bottom: 30px; box-shadow: 0 2px 15px rgba(0,0,0,0.08); }
.header h1 { color: #3498db; margin-bottom: 10px; }
.commit-card { background: white; padding: 25px; border-radius: 12px; margin: 15px 0; box-shadow: 0 2px 10px rgba(0,0,0,0.05); }
.commit-card:hover { box-shadow: 0 5px 20px rgba(0,0,0,0.1); }
.commit-header { display: flex; justify-content: space-between; margin-bottom: 15px; }
.commit-hash { font-family: monospace; color: #3498db; font-weight: bold; }
.commit-meta { color: #666; font-size: 0.9rem; }
.commit-message h3 { margin-bottom: 10px; }
.file-changes { margin: 20px 0; }
.file-item { padding: 10px; background: #f8f9fa; margin: 5px 0; border-radius: 6px; }
.insertions { color: #2ecc71; margin-right: 10px; }
.deletions { color: #e74c3c; }
@media (max-width: 768px) {
    .container { padding: 10px; }
    .header { padding: 20px; }
    .commit-header { flex-direction: column; }
}`
	
	return os.WriteFile(filepath.Join(assetDir, "style.css"), []byte(minimalCSS), 0644)
}

func openInBrowser(filePath string) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		return
	}

	fileURL := "file://" + absPath
	
	// Platform-specific browser opening
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", fileURL).Start()
	case "windows":
		exec.Command("cmd", "/c", "start", fileURL).Start()
	case "linux":
		exec.Command("xdg-open", fileURL).Start()
	default:
		fmt.Printf("Open %s in your browser\n", fileURL)
	}
}

func init() {
    rootCmd.AddCommand(webCmd)

    // Web-specific flags (these are fine)
    webCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output HTML file (default: git-history.html)")
    webCmd.Flags().StringVar(&title, "title", "", "Custom title for the webpage")
    webCmd.Flags().StringVar(&description, "description", "", "Custom description for the webpage")
    webCmd.Flags().BoolVar(&groupByDate, "group-by-date", false, "Group commits by date")
    webCmd.Flags().BoolVar(&groupByAuthor, "group-by-author", false, "Group commits by author")
    webCmd.Flags().StringVar(&theme, "theme", "auto", "Theme (light, dark, auto)")
    webCmd.Flags().BoolVarP(&openBrowser, "open", "p", false, "Open in browser after generation")

    // OPTIONAL: Inherit root flags cleanly (DO NOT re-declare)
    webCmd.Flags().AddFlagSet(rootCmd.PersistentFlags())
    webCmd.Flags().AddFlagSet(rootCmd.Flags())
}
