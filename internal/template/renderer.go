package template

import (
	// "bytes"
	"fmt"
	"html/template"
	"human-git-history/internal/git"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

import "encoding/json"

type TemplateData struct {
	Commits     []git.Commit
	Title       string
	Description string
	GeneratedAt time.Time
	Stats       *RepoStats
	Options     RenderOptions
}

type RepoStats struct {
	TotalCommits   int
	TotalAuthors   int
	FirstCommit    time.Time
	LastCommit     time.Time
	FilesChanged   int
	TotalInsertions int
	TotalDeletions  int
	Authors        map[string]AuthorStats
}

type AuthorStats struct {
	Name       string
	Email      string
	Commits    int
	Insertions int
	Deletions  int
}

type RenderOptions struct {
	ShowFiles    bool
	ShowStats    bool
	GroupByDate  bool
	GroupByAuthor bool
	Since        string
	Until        string
	Author       string
	Theme        string // light, dark, auto
	CompactView  bool
}

type TemplateRenderer struct {
	templates map[string]*template.Template
	assetDir  string
}

func NewRenderer(templateDir, assetDir string) (*TemplateRenderer, error) {
	tr := &TemplateRenderer{
		templates: make(map[string]*template.Template),
		assetDir:  assetDir,
	}

	// Define template functions
	funcMap := template.FuncMap{
		"formatDate":       formatDate,
		"formatTimeAgo":    formatTimeAgo,
		"formatDateTime":   formatDateTime,
		"shortHash":        shortHash,
		"fileStatusColor":  fileStatusColor,
		"fileStatusIcon":   fileStatusIcon,
		"fileStatusText":   fileStatusText,
		"commitStatus":     commitStatus,
		"calculateAge":     calculateAge,
		"pluralize":        pluralize,
		"add":              add,
		"subtract":         subtract,
		"divide":           divide,
		"multiply":         multiply,
		"percentage":       percentage,
		"truncate":         truncate,
		"join":             strings.Join,
		"contains":         strings.Contains,
		"hasPrefix":        strings.HasPrefix,
		"hasSuffix":        strings.HasSuffix,
		"split":            strings.Split,
		"toUpper":          strings.ToUpper,
		"toLower":          strings.ToLower,
		"replace":          strings.ReplaceAll,
		"safeHTML":         func(s string) template.HTML { return template.HTML(s) },
		"safeJS":           func(s string) template.JS { return template.JS(s) },
	}

	// Load and parse templates
	templates := []string{
		"index.tpl",
		"commit.tpl",
		"changelog.tpl",
		"stats.tpl",
	}

	for _, tmpl := range templates {
		path := filepath.Join(templateDir, tmpl)
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read template %s: %v", tmpl, err)
		}

		t, err := template.New(tmpl).Funcs(funcMap).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %v", tmpl, err)
		}

		tr.templates[tmpl] = t
	}

	return tr, nil
}

func (tr *TemplateRenderer) RenderIndex(w io.Writer, data TemplateData) error {
	tmpl, ok := tr.templates["index.tpl"]
	if !ok {
		return fmt.Errorf("index template not found")
	}
	return tmpl.Execute(w, data)
}

func (tr *TemplateRenderer) RenderCommit(w io.Writer, commit git.Commit, options RenderOptions) error {
	tmpl, ok := tr.templates["commit.tpl"]
	if !ok {
		return fmt.Errorf("commit template not found")
	}
	return tmpl.Execute(w, struct {
		Commit  git.Commit
		Options RenderOptions
	}{
		Commit:  commit,
		Options: options,
	})
}

func (tr *TemplateRenderer) RenderChangelog(w io.Writer, data TemplateData) error {
	tmpl, ok := tr.templates["changelog.tpl"]
	if !ok {
		return fmt.Errorf("changelog template not found")
	}
	return tmpl.Execute(w, data)
}

func (tr *TemplateRenderer) RenderStats(w io.Writer, stats RepoStats) error {
	tmpl, ok := tr.templates["stats.tpl"]
	if !ok {
		return fmt.Errorf("stats template not found")
	}
	return tmpl.Execute(w, stats)
}

func (tr *TemplateRenderer) RenderToFile(filename string, data TemplateData) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Copy assets if asset directory exists
	if tr.assetDir != "" {
		assetDest := filepath.Join(dir, "assets")
		if err := copyDir(tr.assetDir, assetDest); err != nil {
			fmt.Printf("Warning: could not copy assets: %v\n", err)
		}
	}

	return tr.RenderIndex(file, data)
}

// Helper functions for templates
func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func formatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
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

func shortHash(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}

func fileStatusColor(status string) string {
	switch status {
	case "Added":
		return "success"
	case "Modified":
		return "warning"
	case "Deleted":
		return "danger"
	case "Renamed", "Copied":
		return "info"
	default:
		return "secondary"
	}
}

func fileStatusIcon(status string) string {
	switch status {
	case "Added":
		return "➕"
	case "Modified":
		return "✏️"
	case "Deleted":
		return "🗑️"
	case "Renamed":
		return "🔄"
	case "Copied":
		return "📋"
	default:
		return "📄"
	}
}

func fileStatusText(status string) string {
	return status
}

func commitStatus(commit git.Commit) string {
	if len(commit.ParentHashes) > 1 {
		return "merge"
	}
	if commit.Stats != nil && commit.Stats.Insertions == 0 && commit.Stats.Deletions > 0 {
		return "cleanup"
	}
	if commit.Stats != nil && commit.Stats.Insertions > 100 {
		return "major"
	}
	return "normal"
}

func calculateAge(t time.Time) string {
	now := time.Now()
	days := int(now.Sub(t).Hours() / 24)
	if days < 7 {
		return "recent"
	} else if days < 30 {
		return "week"
	} else if days < 365 {
		return "month"
	}
	return "year"
}

func pluralize(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func add(a, b int) int {
	return a + b
}

func subtract(a, b int) int {
	return a - b
}

func divide(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}

func multiply(a, b int) int {
	return a * b
}

func percentage(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

func copyAssets(srcDir, destDir string) error {
    if _, err := os.Stat(srcDir); os.IsNotExist(err) {
        // Create default CSS if assets don't exist
        return createDefaultAssets(destDir)
    }
    return copyDir(srcDir, destDir)
}

// Additional template functions
func toJSON(v interface{}) (template.JS, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return template.JS(""), err
	}
	return template.JS(b), nil
}

func slice(s string, start, end int) string {
	if start < 0 || end > len(s) || start > end {
		return s
	}
	return s[start:end]
}

func lower(s string) string {
	return strings.ToLower(s)
}

func upper(s string) string {
	return strings.ToUpper(s)
}

func replaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

func createDefaultAssets(destDir string) error {
    if err := os.MkdirAll(destDir, 0755); err != nil {
        return err
    }
    
    // Create default CSS file
    cssPath := filepath.Join(destDir, "style.css")
    return os.WriteFile(cssPath, []byte(defaultCSS), 0644)
}

const defaultCSS = `/* Default CSS included when assets are not found */
/* This ensures the page is always styled */
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: Arial, sans-serif; line-height: 1.6; }
.container { max-width: 1200px; margin: 0 auto; padding: 20px; }
.commit-card { background: #f8f9fa; padding: 20px; margin: 15px 0; border-radius: 8px; }
`

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}