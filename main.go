package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// This struct will hold the data for a single page.
// We'll pass this to the 'page.html' template.
type Page struct {
	Title        string
	Body         string // The content of the page
	Foot         string //unused
	YouTubeEmbed string
	Head         string
	Year         int
}

// Global variable to cache all our templates
var templates *template.Template

// This regex is used to create a "slug" from a page title.
// e.g., "My New Page" -> "my-new-page"
var slugRegex = regexp.MustCompile("[^a-zA-Z0-9-]+")

func main() {
	// Parse all templates in the 'templates' directory on startup.
	// template.Must() will panic if it can't parse, which is fine for startup.
	templates = template.Must(template.ParseGlob("templates/*.html"))

	// --- Register our HTTP handlers ---

	// 1. The Homepage:
	http.HandleFunc("/", indexHandler)

	// 2. The dynamic page viewer. Note the trailing slash!
	// This tells the router to send all requests starting with /page/ to this handler.
	http.HandleFunc("/page/", pageViewHandler)

	// 3. The API endpoint to create a new page:
	http.HandleFunc("/create", createPageHandler)

	// Start the server
	log.Println("🚀 Starting server on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// --- Handler Functions ---

// indexHandler serves the homepage (index.html)
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// We need to get a list of all pages to display
	files, err := os.ReadDir("pages")
	if err != nil {
		log.Printf("Error reading pages directory: %v", err)
		http.Error(w, "Could not list pages", http.StatusInternalServerError)
		return
	}

	var pageNames []string
	for _, file := range files {
		// Only list text files and trim the .txt extension
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".txt") {
			name := strings.TrimSuffix(file.Name(), ".txt")
			pageNames = append(pageNames, name)
		}
	}

	// Execute the 'index.html' template, passing in the list of page names
	err = templates.ExecuteTemplate(w, "index.html", pageNames)
	if err != nil {
		log.Printf("Error executing index template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
