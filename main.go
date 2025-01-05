package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv" // Import for loading .env file
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Source Structure
type Source struct {
	ID   interface{} `json:"id"`
	Name string      `json:"name"`
}

// Article Structure
type Article struct {
	Source      Source    `json:"source"`
	Author      string    `json:"author"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	URLToImage  string    `json:"urlToImage"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
}

// FormatPublishedDate formats the date
func (a *Article) FormatPublishedDate() string {
	year, month, day := a.PublishedAt.Date()
	return fmt.Sprintf("%v %d, %d", month, day, year)
}

// Results Structure
type Results struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

// Search Query Structure
type Search struct {
	SearchKey  string
	NextPage   int
	TotalPages int
	Results    Results
}

// IsLastPage checks if it's the last page
func (s *Search) IsLastPage() bool {
	return s.NextPage >= s.TotalPages
}

// CurrentPage gets the current page number
func (s *Search) CurrentPage() int {
	if s.NextPage == 1 {
		return s.NextPage
	}
	return s.NextPage - 1
}

// PreviousPage gets the previous page number
func (s *Search) PreviousPage() int {
	return s.CurrentPage() - 1
}

var tpl = template.Must(template.ParseFiles("index.html"))
var apiKey string

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
}

// Search Requests
func searchHandler(w http.ResponseWriter, r *http.Request) {
    u, err := url.Parse(r.URL.String())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("Internal server error"))
        return
    }

    params := u.Query()
    searchKey := params.Get("q")
    page := params.Get("page")
    if page == "" {
        page = "1"
    }

    fmt.Println("Search Query is: ", searchKey)
    search := &Search{}
    search.SearchKey = searchKey

    next, err := strconv.Atoi(page)
    if err != nil {
        http.Error(w, "Unexpected server error", http.StatusInternalServerError)
        return
    }

    search.NextPage = next
    pageSize := 20

    // Indonesian News API Query
    endpoint := fmt.Sprintf(
        "https://newsapi.org/v2/everything?q=%s&pageSize=%d&page=%d&apiKey=%s&language=id",
        url.QueryEscape(search.SearchKey), pageSize, search.NextPage, apiKey,
    )

    resp, err := http.Get(endpoint)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    err = json.NewDecoder(resp.Body).Decode(&search.Results)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    search.TotalPages = int(math.Ceil(float64(search.Results.TotalResults) / float64(pageSize)))
    err = tpl.Execute(w, search)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
    }

    if !search.IsLastPage() {
        search.NextPage++
    }
}
func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	apiKey = os.Getenv("NEWS_API_KEY")
	if apiKey == "" {
		log.Fatal("NEWS_API_KEY must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Static Files
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Server Handlers
	mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/", indexHandler)

	// Start the server
	fmt.Printf("Starting server on port %s...\n", port)
	http.ListenAndServe(":"+port, mux)
}
