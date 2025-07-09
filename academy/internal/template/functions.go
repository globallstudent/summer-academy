package template

import (
"html/template"
"time"

"github.com/microcosm-cc/bluemonday"
"github.com/russross/blackfriday/v2"
)

// Functions returns a map of template functions
func Functions() template.FuncMap {
	return template.FuncMap{
		"markdownToHTML": MarkdownToHTML,
		"add":            add,
		"formatTime":     formatTime,
	}
}

// MarkdownToHTML converts markdown text to HTML
func MarkdownToHTML(md string) template.HTML {
	// Convert markdown to HTML
	unsafe := blackfriday.Run([]byte(md))
	
	// Sanitize HTML to prevent XSS
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	
	return template.HTML(html)
}

// add adds two integers
func add(a, b int) int {
	return a + b
}

// formatTime formats a time.Time value
func formatTime(t time.Time) string {
	return t.Format("Jan 2, 2006 15:04")
}
