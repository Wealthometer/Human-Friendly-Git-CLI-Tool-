<div class="commit-card" 
     data-author="{{.Commit.AuthorName}}"
     data-date="{{.Commit.AuthorDate | formatDate}}"
     data-change-type="{{if .Commit.FileChanges}}{{range .Commit.FileChanges}}{{.Status | slice 0 | upper}}{{end}}{{end}}">
    
    <!-- Commit Header -->
    <div class="commit-header">
        <div class="commit-hash">
            <i class="fas fa-code-commit"></i>
            <a href="#" class="hash-link" title="{{.Commit.Hash}}">
                {{.Commit.ShortHash | shortHash}}
            </a>
        </div>
        <div class="commit-meta">
            <span class="author">
                <i class="fas fa-user"></i>
                {{.Commit.AuthorName}}
            </span>
            <span class="date">
                <i class="fas fa-clock"></i>
                {{.Commit.AuthorDate | formatTimeAgo}}
                <small>({{.Commit.AuthorDate | formatDateTime}})</small>
            </span>
        </div>
    </div>

    <!-- Commit Message -->
    <div class="commit-message">
        <h3>{{.Commit.Message}}</h3>
        {{if .Commit.Body}}
        <div class="commit-body">
            <p>{{.Commit.Body}}</p>
        </div>
        {{end}}
    </div>

    <!-- File Changes -->
    {{if .Options.ShowFiles}}
    <div class="file-changes">
        <h4><i class="fas fa-file-alt"></i> Changed Files ({{len .Commit.FileChanges}})</h4>
        <div class="file-list">
            {{range .Commit.FileChanges}}
            <div class="file-item status-{{.Status | fileStatusColor}}">
                <span class="file-status">
                    <i class="fas {{.Status | fileStatusIcon}}"></i>
                    {{.Status | fileStatusText}}
                </span>
                <span class="file-path">{{.FilePath}}</span>
                {{if .OldPath}}
                <span class="file-rename">
                    <i class="fas fa-arrow-right"></i>
                    {{.OldPath}}
                </span>
                {{end}}
                {{if or .Insertions .Deletions}}
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

    <!-- Commit Stats -->
    {{if .Options.ShowStats}}
    <div class="commit-stats">
        <div class="stat-item">
            <i class="fas fa-file"></i>
            <span>{{.Commit.Stats.FilesChanged}} files</span>
        </div>
        <div class="stat-item positive">
            <i class="fas fa-plus-circle"></i>
            <span>+{{.Commit.Stats.Insertions}}</span>
        </div>
        <div class="stat-item negative">
            <i class="fas fa-minus-circle"></i>
            <span>-{{.Commit.Stats.Deletions}}</span>
        </div>
    </div>
    {{end}}

    <!-- Commit Footer -->
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
        <div class="commit-actions">
            <button class="btn-small" onclick="viewCommitDetails('{{.Commit.Hash}}')">
                <i class="fas fa-external-link-alt"></i> View Details
            </button>
            <button class="btn-small" onclick="copyHash('{{.Commit.Hash}}')">
                <i class="fas fa-copy"></i> Copy Hash
            </button>
        </div>
    </div>
</div>