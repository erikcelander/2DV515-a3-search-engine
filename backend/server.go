package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "sort"
    "fmt"
    "math"
)
type WikipediaPage struct {
    URL    string
    WordID []int // List of word IDs
}

var (
    wordToID map[string]int
    pages    []WikipediaPage
)

type SearchResult struct {
    URL       string  `json:"url"`
    ContentScore float64     `json:"frequency"`
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
func performSearch(query string) []SearchResult {
    queryID, exists := wordToID[query]
    if !exists {
        return nil // Query word not found in the index
    }

    scores := make([]float64, len(pages))
    for i, page := range pages {
        for _, wordID := range page.WordID {
            if wordID == queryID {
                scores[i]++
            }
        }
    }

    // Normalize the scores (assuming higher frequency is better, so smallIsBetter is false)
    normalize(scores, false)

    var searchResults []SearchResult
    for i, score := range scores {
        if score > 0 {
            searchResults = append(searchResults, SearchResult{
                URL:          pages[i].URL,
                ContentScore: score, // Directly use the normalized score
            })
        }
    }

    // Sort the results based on normalized frequency
    sort.Slice(searchResults, func(i, j int) bool {
        return searchResults[i].ContentScore > searchResults[j].ContentScore
    })

    return searchResults
}


func initializeIndex(basePath string) {
    wordToID = make(map[string]int)
    var idCounter int

    wordsPath := filepath.Join(basePath, "Words")

    for _, folder := range []string{"Games", "Programming"} {
        folderPath := filepath.Join(wordsPath, folder)
        files, err := os.ReadDir(folderPath)
        if err != nil {
            log.Fatalf("Failed to read directory: %v", err)
        }

        for _, file := range files {
            fileName := file.Name()
            filePath := filepath.Join(folderPath, fileName)
            processFile(filePath, &idCounter)
        }
    }
}

func processFile(filePath string, idCounter *int) {
    content, err := os.ReadFile(filePath)
    if err != nil {
        log.Printf("Failed to read file: %v", err)
        return
    }

    words := strings.Fields(string(content))
    var wordIDs []int
    for _, word := range words {
        wordID, exists := wordToID[word]
        if !exists {
            wordID = *idCounter
            wordToID[word] = wordID
            (*idCounter)++
        }
        wordIDs = append(wordIDs, wordID)
    }

    pages = append(pages, WikipediaPage{
        URL:    filePath,
        WordID: wordIDs,
    })
}

func normalize(scores []float64, smallIsBetter bool) {
    if smallIsBetter {
        minVal := min(scores)
        for i := range scores {
            scores[i] = minVal / math.Max(scores[i], 0.00001)
        }
    } else {
        maxVal := max(scores)
        maxVal = math.Max(maxVal, 0.00001)
        for i := range scores {
            scores[i] = scores[i] / maxVal
        }
    }
}

// min finds the minimum value in a float64 slice
func min(values []float64) float64 {
    minValue := math.MaxFloat64
    for _, v := range values {
        if v < minValue {
            minValue = v
        }
    }
    return minValue
}

// max finds the maximum value in a float64 slice
func max(values []float64) float64 {
    maxValue := -math.MaxFloat64
    for _, v := range values {
        if v > maxValue {
            maxValue = v
        }
    }
    return maxValue
}

func main() {
    relativePath := "./wikipedia"
    absolutePath, err := filepath.Abs(relativePath)
    if err != nil {
        log.Fatalf("Error getting absolute path: %v", err)
    }

    fmt.Println("Using path:", absolutePath)
    initializeIndex(absolutePath)

    // Test queries
    testQueries := []string{"java"} // Single-word queries
    for _, query := range testQueries {
        results := performSearch(query)
        fmt.Printf("Results for '%s':\n", query)
        for i, result := range results {
            fmt.Printf("%d. URL: %s, Frequency: %.2f\n", i+1, result.URL, result.ContentScore)
        }
        fmt.Println() // Newline for better separation
    }


    // Uncomment the below line to start the HTTP server for actual deployment
    // startServer()
}
