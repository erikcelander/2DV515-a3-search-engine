package backend

import (
    "encoding/json"
    "log"
    "net/http"
    "strings"
)

// WikipediaPage structure to hold page data
type WikipediaPage struct {
    URL   string
    Words []int
}

// Search index and mapping
var (
    wordToID map[string]int
    pages    []WikipediaPage
)

// SearchResult structure for JSON response
type SearchResult struct {
    URL       string  `json:"url"`
    RankScore float64 `json:"rankScore"`
}

// Function to initialize and start the HTTP server
func startServer() {
    http.HandleFunc("/search", searchHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// HTTP handler function for search queries
func searchHandler(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("word")
    results := performSearch(query)
    jsonResponse, err := json.Marshal(results)
    if err != nil {
        http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonResponse)
}

// Function to perform the search and calculate word frequency scores
func performSearch(query string) []SearchResult {
    queryID, exists := wordToID[query]
    if !exists {
        return []SearchResult{} // Return empty if query word not found
    }
    
    var searchResults []SearchResult
    for _, page := range pages {
        score := calculateWordFrequencyScore(page, queryID)
        if score > 0 {
            searchResults = append(searchResults, SearchResult{
                URL:       page.URL,
                RankScore: normalizeScore(score),
            })
        }
    }

    // Sort and select top 5 results
    sortSearchResults(&searchResults)
    if len(searchResults) > 5 {
        searchResults = searchResults[:5]
    }

    return searchResults
}

// Placeholder functions for score calculation and normalization (to be implemented)
func calculateWordFrequencyScore(page WikipediaPage, queryID int) float64 {
    // TODO: Implement the word frequency scoring logic
    return 0.0
}

func normalizeScore(score float64) float64 {
    // TODO: Implement the score normalization logic
    return score
}

func sortSearchResults(results *[]SearchResult) {
    // TODO: Implement sorting logic for search results
}

func main() {
    startServer()
}
