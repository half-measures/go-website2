package main

//Holds the page creation POST to generate a new text for a page template
//Also has how we display our pages

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// createPageHandler handles the POST request to create a new page for the pages folder
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
		http.Error(w, "Bad name found, try again. Cannot use symbols, try words only.", http.StatusBadRequest)
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
	defaultBody := "This is the new page for **" + reqBody.Name + "**"
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

	// --- Render the page ---

	// 1. Read the optional YouTube link file
	youtubeFilename := filepath.Join("pages", safeSlug+".youtube.txt")
	youtubeURLs, err := os.ReadFile(youtubeFilename)
	var embedURLs []string
	if err == nil { // File exists
		// Split the file content by newline to get individual URLs
		urls := strings.Split(string(youtubeURLs), "\n")
		for _, url := range urls {
			if url != "" { // Ignore empty lines
				embedURLs = append(embedURLs, processYouTubeURL(url))
			}
		}
	}

	// 2. Create a Page struct with the data
	pageData := &Page{
		Title:        safeSlug,
		Body:         string(body),
		YouTubeEmbed: embedURLs, // Will be nil if no links are found
		Year:         time.Now().Year(),
	}

	// Execute the 'page.html' template
	err = templates.ExecuteTemplate(w, "page.html", pageData)
	if err != nil {
		log.Printf("Error executing page template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
