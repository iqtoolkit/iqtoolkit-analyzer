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
body { font-family: -apple-system, BlinkMacSystemFont, 'Inter', system-ui, sans-serif; background: #f8fafc; color: #0f172a; line-height: 1.6; padding: 2rem; }
.container { max-width: 1200px; margin: 0 auto; }
header { display: flex; align-items: center; gap: 1rem; margin-bottom: 2rem; padding-bottom: 1.5rem; border-bottom: 1px solid #e2e8f0; }
.logo { display: flex; align-items: center; gap: 0.75rem; }
.logo-icon { width: 36px; height: 36px; }
.logo-text { font-size: 1.4rem; font-weight: 700; }
.logo-text .iq { color: #059669; }
.logo-text .rest { color: #64748b; }
.meta { color: #64748b; font-size: 0.85rem; margin-left: auto; text-align: right; }
.meta a { color: #059669; text-decoration: none; }
h2 { color: #059669; font-size: 1.2rem; margin: 2rem 0 0.75rem; padding-bottom: 0.4rem; border-bottom: 1px solid #e2e8f0; }
.version-box { background: #fff; border: 1px solid #e2e8f0; border-radius: 6px; padding: 0.75rem 1rem; font-family: monospace; font-size: 0.9rem; color: #475569; }
table { width: 100%; border-collapse: collapse; margin-bottom: 2rem; font-size: 0.8rem; }
th { background: #f1f5f9; color: #059669; padding: 0.5rem 0.6rem; text-align: left; border-bottom: 2px solid #e2e8f0; }
td { padding: 0.4rem 0.6rem; border-bottom: 1px solid #f1f5f9; color: #475569; }
tr:hover td { background: #f8fafc; }
.installed { color: #059669; font-weight: 600; }
.not-installed { color: #94a3b8; }
footer { margin-top: 3rem; padding-top: 1rem; border-top: 1px solid #e2e8f0; text-align: center; color: #64748b; font-size: 0.8rem; }
footer a { color: #059669; text-decoration: none; }
@media print {
  body { background: #fff; padding: 0; font-size: 9pt; }
  .container { max-width: 100%; }
  header { border-bottom: 1px solid #ccc; }
  h2 { border-bottom: 1px solid #ccc; page-break-after: avoid; }
  table { font-size: 7.5pt; page-break-inside: auto; }
  tr { page-break-inside: avoid; }
  th { background: #f0f0f0 !important; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
  tr:hover td { background: none; }
  footer { border-top: 1px solid #ccc; }
}
</style>
</head>
<body>
<div class="container">
<header>
	<div class="logo">
		<svg class="logo-icon" viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
			<polygon points="50,5 93,27.5 93,72.5 50,95 7,72.5 7,27.5" fill="none" stroke="#475569" stroke-width="6" stroke-linejoin="round"/>
			<line x1="50" y1="5" x2="50" y2="50" stroke="#475569" stroke-width="2"/>
			<line x1="93" y1="27.5" x2="50" y2="50" stroke="#475569" stroke-width="2"/>
			<line x1="93" y1="72.5" x2="50" y2="50" stroke="#475569" stroke-width="2"/>
			<line x1="50" y1="95" x2="50" y2="50" stroke="#475569" stroke-width="2"/>
			<line x1="7" y1="72.5" x2="50" y2="50" stroke="#475569" stroke-width="2"/>
			<line x1="7" y1="27.5" x2="50" y2="50" stroke="#475569" stroke-width="2"/>
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
