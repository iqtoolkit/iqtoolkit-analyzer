package report

import (
	"html/template"
	"io"
	"time"

	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/dbconn"
)

type Data struct {
	Version    string
	Settings   []dbconn.Setting
	Extensions []dbconn.Extension
	GeneratedAt time.Time
}

var tmpl = template.Must(template.New("report").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>PostgreSQL Report</title>
<style>
body { font-family: system-ui, sans-serif; margin: 2rem; }
h1, h2 { color: #336791; }
table { border-collapse: collapse; width: 100%; margin-bottom: 2rem; }
th, td { border: 1px solid #ddd; padding: 0.5rem; text-align: left; }
th { background: #336791; color: white; }
tr:nth-child(even) { background: #f9f9f9; }
.installed { color: green; font-weight: bold; }
.meta { color: #666; font-size: 0.9rem; }
</style>
</head>
<body>
<h1>PostgreSQL Report</h1>
<p class="meta">Generated: {{.GeneratedAt.Format "2006-01-02 15:04:05 MST"}}</p>

<h2>Version</h2>
<p>{{.Version}}</p>

<h2>Settings (pg_settings)</h2>
<table>
<tr><th>Name</th><th>Value</th><th>Source</th></tr>
{{range .Settings}}<tr><td>{{.Name}}</td><td>{{.Value}}</td><td>{{.Source}}</td></tr>
{{end}}</table>

<h2>Extensions</h2>
<table>
<tr><th>Name</th><th>Default Version</th><th>Installed Version</th></tr>
{{range .Extensions}}<tr><td>{{.Name}}</td><td>{{.DefaultVersion}}</td><td>{{if .InstalledVersion}}<span class="installed">{{.InstalledVersion}}</span>{{else}}&mdash;{{end}}</td></tr>
{{end}}</table>
</body>
</html>`))

func Generate(w io.Writer, data Data) error {
	return tmpl.Execute(w, data)
}
