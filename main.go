package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// This struct will hold the data for a single page.
// We'll pass this to the 'page.html' template.
type Page struct {
	Title string
	Body  string // The content of the page
	Foot  string
	Head  string
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

// pageViewHandler serves a single page (page.html)
func pageViewHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the page title (slug) from the URL
	// r.URL.Path will be "/page/my-new-page"
	slug := r.URL.Path[len("/page/"):]

	// Security: Use filepath.Base to prevent directory traversal attacks
	// e.g., prevents a request like /page/../../etc/passwd
	safeSlug := filepath.Base(slug)

	// Load the page content from the file
	filename := filepath.Join("pages", safeSlug+".txt")
	body, err := os.ReadFile(filename)
	if err != nil {
		// If the file doesn't exist, send a 404
		log.Printf("Page not found: %s", filename)
		http.NotFound(w, r)
		return
	}

	// Create a Page struct with the data
	pageData := &Page{
		Title: safeSlug,
		Body:  string(body),
	}

	// Execute the 'page.html' template
	err = templates.ExecuteTemplate(w, "page.html", pageData)
	if err != nil {
		log.Printf("Error executing page template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// createPageHandler handles the POST request to create a new page
func createPageHandler(w http.ResponseWriter, r *http.Request) {

	// We only accept POST requests here
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	// Decode the JSON request body: {"name": "My New Page"}
	var reqBody struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if err := charchecker(reqBody.Name); err != nil {
		http.Error(w, "Bad naming found, try again. Cannot use symbols, try words only", http.StatusBadRequest)
		return
	}

	if reqBody.Name == "" || reqBody.Name == " " {
		http.Error(w, "Page name is required", http.StatusBadRequest)
		return
	}

	// --- Create the page file ---

	// 1. Sanitize the name into a URL-friendly "slug"
	slug := strings.ToLower(reqBody.Name)
	slug = strings.ReplaceAll(slug, " ", "-")   // Replace spaces with hyphens
	slug = slugRegex.ReplaceAllString(slug, "") // Remove all other weird characters

	if slug == "" {
		slug = "untitled" // Fallback for empty/invalid names
	}

	// 2. Define the file path
	filename := filepath.Join("pages", slug+".txt")

	// 3. Check if file already exists. If so, just redirect to it.
	if _, err := os.Stat(filename); err == nil {
		log.Printf("Page already exists, redirecting: %s", slug)
		http.Redirect(w, r, "/page/"+slug, http.StatusFound)
		return
	}

	// 4. Create the new file with default content
	defaultBody := "This is your new page: **" + reqBody.Name + "**"
	err := os.WriteFile(filename, []byte(defaultBody), 0644) // 0644 = rw-r--r--
	if err != nil {
		log.Printf("Error writing new page file: %v", err)
		http.Error(w, "Could not save page", http.StatusInternalServerError)
		return
	}

	log.Printf("New page created: %s", filename)

	// 5. Redirect the user to their new page
	http.Redirect(w, r, "/page/"+slug, http.StatusSeeOther)
}
