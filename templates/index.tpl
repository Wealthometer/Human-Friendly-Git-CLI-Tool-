<!DOCTYPE html>
<html lang="en" data-theme="{{.Options.Theme}}">
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
            <button class="btn" onclick="exportToJSON()">
                <i class="fas fa-download"></i> Export JSON
            </button>
            <div class="search">
                <input type="text" id="searchInput" placeholder="Search commits...">
                <button class="btn" onclick="searchCommits()">
                    <i class="fas fa-search"></i>
                </button>
            </div>
        </div>

        <!-- Stats Panel -->
<div id="statsPanel" class="stats-panel">
    <h2>Repository Statistics</h2>
    <div class="stat-item">
        <strong>Total Commits:</strong> {{.Stats.TotalCommits}}
    </div>
    <div class="stat-item">
        <strong>Total Authors:</strong> {{.Stats.TotalAuthors}}
    </div>
    <div class="stat-item">
        <strong>Lines of Code:</strong> {{add .Stats.TotalInsertions .Stats.TotalDeletions}}
        <small>(+{{.Stats.TotalInsertions}} / -{{.Stats.TotalDeletions}})</small>
    </div>
</div>


        <!-- Filter Bar -->
        <div class="filters">
            <div class="filter-group">
                <label for="authorFilter">Author:</label>
                <select id="authorFilter" onchange="filterByAuthor()">
                    <option value="">All Authors</option>
                    {{range $author, $stats := .Stats.Authors}}
                    <option value="{{$author}}">{{$stats.Name}} ({{$stats.Commits}})</option>
                    {{end}}
                </select>
            </div>
            <div class="filter-group">
                <label for="dateFilter">Date Range:</label>
                <input type="date" id="dateFrom" onchange="filterByDate()">
                <input type="date" id="dateTo" onchange="filterByDate()">
            </div>
            <div class="filter-group">
                <label for="typeFilter">Change Type:</label>
                <select id="typeFilter" onchange="filterByType()">
                    <option value="">All Changes</option>
                    <option value="A">Added Files</option>
                    <option value="M">Modified Files</option>
                    <option value="D">Deleted Files</option>
                    <option value="R">Renamed Files</option>
                </select>
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
                    tmpl, err := template.New("index.tpl").Funcs(funcMap).ParseFiles(
    "templates/index.tpl",
    "templates/commit.tpl",
    "templates/stats.tpl",
)
if err != nil {
    fmt.Fprintf(os.Stderr, "Error parsing templates: %v\n", err)
    os.Exit(1)
}

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
                    <div class="commit-card" 
     data-author="{{.AuthorName}}" 
     data-date="{{.AuthorDate | formatDate}}">
    <div class="commit-header">
        <span class="commit-hash">{{.ShortHash}}</span>
        <span class="commit-author">{{.AuthorName}}</span>
        <span class="commit-date">{{.AuthorDate | formatDateTime}}</span>
    </div>
    <div class="commit-message">{{.Message}}</div>
</div>

                {{end}}
            {{else}}
                {{range .Commits}}
                    <div class="commit-card" 
     data-author="{{.AuthorName}}" 
     data-date="{{.AuthorDate | formatDate}}">
    <div class="commit-header">
        <span class="commit-hash">{{.ShortHash}}</span>
        <span class="commit-author">{{.AuthorName}}</span>
        <span class="commit-date">{{.AuthorDate | formatDateTime}}</span>
    </div>
    <div class="commit-message">{{.Message}}</div>
</div>

                {{end}}
            {{end}}
        </div>

        <!-- Pagination -->
        <div class="pagination">
            <button class="btn" onclick="loadMore()" id="loadMoreBtn">
                <i class="fas fa-plus"></i> Load More Commits
            </button>
        </div>

        <!-- Footer -->
        <footer class="footer">
            <p>
                Generated with <i class="fas fa-heart" style="color: #e74c3c;"></i> 
                by Human Git History Tool
            </p>
            <p class="footer-links">
                <a href="#" onclick="scrollToTop()"><i class="fas fa-arrow-up"></i> Back to Top</a>
                <a href="https://github.com" target="_blank"><i class="fab fa-github"></i> View on GitHub</a>
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
            // Implementation depends on your data structure
        }

        // Filter by type
        function filterByType() {
            const type = document.getElementById('typeFilter').value;
            if (!type) {
                showAllCommits();
                return;
            }
            filterCommits('data-change-type', type);
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

        // Export to JSON
        function exportToJSON() {
            const data = {
                commits: {{.Commits | json}},
                stats: {{.Stats | json}},
                generatedAt: {{.GeneratedAt | formatDateTime}}
            };
            
            const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'git-history.json';
            a.click();
            URL.revokeObjectURL(url);
        }

        // Scroll to top
        function scrollToTop() {
            window.scrollTo({ top: 0, behavior: 'smooth' });
        }

        // Load more commits (simplified)
        function loadMore() {
            // This would typically make an API call
            alert('Load more functionality would fetch additional commits');
        }

        // Initialize charts
        document.addEventListener('DOMContentLoaded', function() {
            // Activity chart
            const ctx = document.getElementById('activityChart').getContext('2d');
            new Chart(ctx, {
                type: 'line',
                data: {
                    labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'],
                    datasets: [{
                        label: 'Commits',
                        data: [12, 19, 3, 5, 2, 3],
                        borderColor: 'rgb(75, 192, 192)',
                        tension: 0.1
                    }]
                }
            });
        });
    </script>
</body>
</html>