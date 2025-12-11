package main

import (
	"database/sql"
	"encoding/json"
	"flag"

	"go-trailer/models"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type config struct {
	DSN string `json:"dsn"`
}

type application struct {
	errorLog     *log.Logger
	infoLog      *log.Logger
	trailerModel *models.TrailerModel
}

// This struct will hold the data for a single page.
// We'll pass this to the 'page.html' template.
type Page struct {
	Title        string
	Body         string // The content of the page
	Foot         string //unused
	YouTubeEmbed []YouTubeVideo
	Head         string
	Year         int
}

// YouTubeVideo holds the data for a single YouTube video, including its vote count.
type YouTubeVideo struct {
	ID    string
	URL   string
	Votes int
}

// Global variable to cache all our templates
var templates *template.Template

// This regex is used to create a "slug" from a page title.
// e.g., "My New Page" -> "my-new-page"
var slugRegex = regexp.MustCompile("[^a-zA-Z0-9-]+")

func main() {
	file, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	// 1. Load configuration
	var cfg config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}
	flag.Parse() //Sanitizes the arg coming in just in case

	db, err := openDB(cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := &application{
		errorLog: log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		infoLog:  log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),

		// Initialize the model by passing the database connection!
		trailerModel: &models.TrailerModel{DB: db},
	}

	// Parse all templates in the 'templates' directory on startup.

	// template.Must() will panic if it can't parse, which is fine for startup.
	templates = template.Must(template.ParseGlob("templates/*.html"))

	// --- Register our HTTP handlers ---

	// 1. The Homepage
	http.HandleFunc("/", indexHandler)

	// 2. The dynamic page viewer.
	// This tells the router to send all requests starting with /page/ to this handler.
	http.HandleFunc("/page/", app.pageViewHandler)

	// 3. The API endpoint to create a new page:
	http.HandleFunc("/create", app.createPageHandler)

	// 4. A file server to serve our static CSS file
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 5. The API endpoint to save a YouTube link for a page:
	http.HandleFunc("/api/page/", app.youtubeSaveHandler)

	// 6. The API endpoint for upvoting/downvoting a YouTube video:
	http.HandleFunc("/api/vote/", app.youtubeVoteHandler)

	// Start the server
	log.Println("🚀 Starting server on http://localhost:8080")
	app.infoLog.Println("Server starting...")
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

// youtubeVoteHandler handles the POST request to upvote or downvote a YouTube video.
// The URL format is /api/vote/{slug}/{videoID}/{action}
func (app *application) youtubeVoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	slug := pathParts[2]
	videoID := pathParts[3]
	action := pathParts[4]

	if action != "upvote" && action != "downvote" {
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	// Read the votes file
	votesFilename := filepath.Join("pages", slug+".votes.json")
	votes := make(map[string]int)

	data, err := os.ReadFile(votesFilename)
	if err == nil {
		if err := json.Unmarshal(data, &votes); err != nil {
			log.Printf("Error unmarshalling votes: %v", err)
			http.Error(w, "Could not process votes", http.StatusInternalServerError)
			return
		}
	}

	// Update the vote count
	if action == "upvote" {
		votes[videoID]++
	} else {
		votes[videoID]--
	}

	// Write the updated votes back to the file
	updatedData, err := json.Marshal(votes)
	if err != nil {
		log.Printf("Error marshalling votes: %v", err)
		http.Error(w, "Could not save vote", http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(votesFilename, updatedData, 0644); err != nil {
		log.Printf("Error writing votes file: %v", err)
		http.Error(w, "Could not save vote", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Vote saved!"))
	log.Printf("Vote saved for video %s on page %s", videoID, slug)
}

// youtubeSaveHandler handles the POST request to save a YouTube link for a page.
// The slug is extracted from the URL, e.g., /api/page/my-page/save-youtube
func (app *application) youtubeSaveHandler(w http.ResponseWriter, r *http.Request) {
	// 1. We only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	// 2. Extract the page slug from the URL
	// The path will be /api/page/my-page-slug/save-youtube
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	slug := pathParts[3]

	// 3. Decode the JSON request body: {"youtube_url": "https://..."}
	var reqBody struct {
		URL string `json:"youtube_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// 4. Basic validation: is it a real YouTube link?
	// Our regex helper is perfect for this.
	embedURL, _ := extractYouTubeVideoInfo(reqBody.URL)
	if embedURL == "" {
		http.Error(w, "Invalid YouTube URL", http.StatusBadRequest)
		return
	}

	// 5. Get the movie ID from the slug
	movie, err := app.trailerModel.GetMovieBySlug(slug)
	if err != nil {
		log.Printf("Error getting movie by slug: %v", err)
		http.Error(w, "Could not find movie", http.StatusInternalServerError)
		return
	}

	// 6. Create a new trailer
	_, videoID := extractYouTubeVideoInfo(reqBody.URL)
	trailer := &models.Trailer{
		MovieID:   movie.ID,
		Title:     "User Submitted Trailer", // You might want to get this from the request
		YouTubeID: videoID,
	}

	// 7. Insert the new trailer into the database
	_, err = app.trailerModel.InsertTrailer(trailer)
	if err != nil {
		log.Printf("Error inserting trailer: %v", err)
		http.Error(w, "Could not save link", http.StatusInternalServerError)
		return
	}

	// 6. Send a success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("YouTube link saved!"))
	log.Printf("YouTube link saved for page: %s", slug)
}

// OpenDB() function wraps sql.open and returns the sql.DB connection pool
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn) //sql.open dosent create any connections, just inits a pool
	if err != nil {
		return nil, err

	} //ping to check if the connection is good
	if err = db.Ping(); err != nil {
		return nil, err

	}
	return db, nil
}
