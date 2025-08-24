package web

import (
	// Remove the "embed" import since we're not using it
	"html/template"
	"net/http"
	"path/filepath"
)

// Templates holds all HTML templates
var Templates *template.Template

// Initialize sets up the web interface
func Initialize() error {
	// Parse templates
	var err error
	Templates, err = template.ParseGlob(filepath.Join("server", "web", "templates", "*.html"))
	if err != nil {
		return err
	}

	// Set up static file server
	http.Handle("/static/", http.StripPrefix("/static/", 
		http.FileServer(http.Dir(filepath.Join("server", "web", "static")))))
	
	// Set up routes
	http.HandleFunc("/dashboard", handleDashboard)
	
	return nil
}

// handleDashboard serves the main dashboard page
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	Templates.ExecuteTemplate(w, "dashboard.html", nil)
}
