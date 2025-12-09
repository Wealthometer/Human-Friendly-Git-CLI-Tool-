<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Changelog - {{.Title}}</title>
    <link rel="stylesheet" href="assets/style.css">
    <style>
        .changelog {
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
        }
        .changelog-version {
            margin: 40px 0;
            padding: 20px;
            background: var(--card-bg);
            border-radius: 8px;
        }
        .changelog-date {
            color: #666;
            font-size: 0.9em;
        }
        .change-type {
            margin: 20px 0;
        }
        .change-type h4 {
            border-bottom: 2px solid;
            padding-bottom: 5px;
            margin-bottom: 10px;
        }
        .change-added h4 { color: var(--secondary-color); border-color: var(--secondary-color); }
        .change-changed h4 { color: var(--warning-color); border-color: var(--warning-color); }
        .change-fixed h4 { color: var(--danger-color); border-color: var(--danger-color); }
        .change-item {
            margin: 5px 0;
            padding-left: 20px;
        }
    </style>
</head>
<body>
    <div class="changelog">
        <header>
            <h1>{{.Title}} - Changelog</h1>
            <p class="subtitle">{{.Description}}</p>
            <p>Generated: {{.GeneratedAt | formatDate}}</p>
        </header>

        {{$currentDate := ""}}
        {{range .Commits}}
            {{$commitDate := .AuthorDate | formatDate}}
            {{if ne $commitDate $currentDate}}
                {{$currentDate = $commitDate}}
                <div class="changelog-date">
                    <h2>{{$currentDate}}</h2>
                </div>
            {{end}}
            
            <div class="changelog-version">
                <h3>{{.Message}}</h3>
                <p class="meta">
                    <strong>{{.AuthorName}}</strong> - {{.AuthorDate | formatTimeAgo}}
                </p>
                
                {{if .Body}}
                <div class="description">
                    {{.Body}}
                </div>
                {{end}}
                
                {{if .Options.ShowFiles}}
                <div class="changes">
                    {{range .FileChanges}}
                    <div class="change-type change-{{.Status | lower}}">
                        <h4>{{.Status}}</h4>
                        <div class="change-item">
                            <code>{{.FilePath}}</code>
                            {{if .Insertions .Deletions}}
                            <span class="stats">(+{{.Insertions}}/-{{.Deletions}})</span>
                            {{end}}
                        </div>
                    </div>
                    {{end}}
                </div>
                {{end}}
            </div>
        {{end}}
        
        <footer>
            <p>Generated with Human Git History Tool</p>
        </footer>
    </div>
</body>
</html>