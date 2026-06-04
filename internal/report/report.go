package report

import (
	"html/template"
	"io"
	"time"

	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/dbconn"
)

type Data struct {
	Version     string
	Settings    []dbconn.Setting
	Extensions  []dbconn.Extension
	GeneratedAt time.Time
}

var tmpl = template.Must(template.New("report").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>PostgreSQL Report — iqtoolkit-analyzer</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: -apple-system, BlinkMacSystemFont, 'Inter', system-ui, sans-serif; background: #030712; color: #e4ebff; line-height: 1.6; padding: 2rem; }
.container { max-width: 1200px; margin: 0 auto; }
header { display: flex; align-items: center; gap: 1rem; margin-bottom: 2rem; padding-bottom: 1.5rem; border-bottom: 1px solid #1f2b3f; }
.logo { display: flex; align-items: center; gap: 0.75rem; }
.logo-icon { width: 40px; height: 40px; }
.logo-text { font-size: 1.5rem; font-weight: 700; }
.logo-text .iq { color: #3fb366; }
.logo-text .rest { color: #8ea2c6; }
.meta { color: #8ea2c6; font-size: 0.85rem; margin-left: auto; text-align: right; }
.meta a { color: #3fb366; text-decoration: none; }
.meta a:hover { text-decoration: underline; }
h2 { color: #3fb366; font-size: 1.25rem; margin: 2rem 0 1rem; padding-bottom: 0.5rem; border-bottom: 1px solid #1f2b3f; }
.version-box { background: #0b1425; border: 1px solid #1f2b3f; border-radius: 8px; padding: 1rem 1.5rem; font-family: monospace; font-size: 0.9rem; color: #c0c9e5; }
table { width: 100%; border-collapse: collapse; margin-bottom: 2rem; font-size: 0.85rem; }
th { background: #0b1425; color: #3fb366; padding: 0.6rem 0.8rem; text-align: left; border-bottom: 2px solid #1f2b3f; position: sticky; top: 0; }
td { padding: 0.5rem 0.8rem; border-bottom: 1px solid #111b2d; color: #c0c9e5; }
tr:hover td { background: #0b1425; }
.installed { color: #6dffbd; font-weight: 600; }
.not-installed { color: #8ea2c6; }
footer { margin-top: 3rem; padding-top: 1.5rem; border-top: 1px solid #1f2b3f; text-align: center; color: #8ea2c6; font-size: 0.8rem; }
footer a { color: #3fb366; text-decoration: none; }
</style>
</head>
<body>
<div class="container">
<header>
	<div class="logo">
		<svg class="logo-icon" viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
			<polygon points="50,5 93,27.5 93,72.5 50,95 7,72.5 7,27.5" fill="none" stroke="#8ea2c6" stroke-width="6" stroke-linejoin="round"/>
			<line x1="50" y1="5" x2="50" y2="50" stroke="#8ea2c6" stroke-width="2"/>
			<line x1="93" y1="27.5" x2="50" y2="50" stroke="#8ea2c6" stroke-width="2"/>
			<line x1="93" y1="72.5" x2="50" y2="50" stroke="#8ea2c6" stroke-width="2"/>
			<line x1="50" y1="95" x2="50" y2="50" stroke="#8ea2c6" stroke-width="2"/>
			<line x1="7" y1="72.5" x2="50" y2="50" stroke="#8ea2c6" stroke-width="2"/>
			<line x1="7" y1="27.5" x2="50" y2="50" stroke="#8ea2c6" stroke-width="2"/>
		</svg>
		<span class="logo-text"><span class="iq">IQ</span><span class="rest">toolkit.ai</span></span>
	</div>
	<div class="meta">
		<div>Generated: {{.GeneratedAt.Format "2006-01-02 15:04:05 MST"}}</div>
		<div>by <a href="https://thepostgresguy.com">The Postgres Guy</a></div>
	</div>
</header>

<h2>PostgreSQL Version</h2>
<div class="version-box">{{.Version}}</div>

<h2>Settings (pg_settings)</h2>
<table>
<tr><th>Name</th><th>Value</th><th>Source</th></tr>
{{range .Settings}}<tr><td>{{.Name}}</td><td>{{.Value}}</td><td>{{.Source}}</td></tr>
{{end}}</table>

<h2>Extensions</h2>
<table>
<tr><th>Name</th><th>Default Version</th><th>Installed Version</th></tr>
{{range .Extensions}}<tr><td>{{.Name}}</td><td>{{.DefaultVersion}}</td><td>{{if .InstalledVersion}}<span class="installed">{{.InstalledVersion}}</span>{{else}}<span class="not-installed">—</span>{{end}}</td></tr>
{{end}}</table>

<footer>
	<p>iqtoolkit-analyzer — by <a href="https://thepostgresguy.com">The Postgres Guy</a> — <a href="https://iqtoolkit.ai">iqtoolkit.ai</a></p>
</footer>
</div>
</body>
</html>`))

func Generate(w io.Writer, data Data) error {
	return tmpl.Execute(w, data)
}
