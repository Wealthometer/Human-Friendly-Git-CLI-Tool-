<div class="stats-container">
    <h2><i class="fas fa-chart-bar"></i> Repository Statistics</h2>
    
    <!-- Overview Stats -->
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
            <div class="stat-icon files">
                <i class="fas fa-file-code"></i>
            </div>
            <div class="stat-content">
                <h3>{{.FilesChanged}}</h3>
                <p>Files Changed</p>
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

    <!-- Timeline -->
    <div class="stat-section">
        <h3><i class="fas fa-calendar-alt"></i> Timeline</h3>
        <div class="timeline">
            <div class="timeline-item">
                <span class="timeline-date">First Commit</span>
                <span class="timeline-value">{{.FirstCommit | formatDate}}</span>
                <span class="timeline-age">{{.FirstCommit | calculateAge}}</span>
            </div>
            <div class="timeline-item">
                <span class="timeline-date">Last Commit</span>
                <span class="timeline-value">{{.LastCommit | formatDate}}</span>
                <span class="timeline-age">{{.LastCommit | formatTimeAgo}}</span>
            </div>
        </div>
    </div>

    <!-- Author Contributions -->
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
                    <span class="stat lines">+{{$stats.Insertions}}/-{{$stats.Deletions}}</span>
                </div>
                <div class="author-progress">
                    <div class="progress-bar" style="width: {{percentage $stats.Commits $.TotalCommits}}%"></div>
                </div>
            </div>
            {{end}}
        </div>
    </div>

    <!-- Charts -->
    <div class="charts-grid">
        <div class="chart-container">
            <h4><i class="fas fa-chart-line"></i> Activity Over Time</h4>
            <canvas id="activityChart"></canvas>
        </div>
        <div class="chart-container">
            <h4><i class="fas fa-chart-pie"></i> Author Distribution</h4>
            <canvas id="authorChart"></canvas>
        </div>
    </div>
</div>