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
    "time"
)
type WikipediaPage struct {
    URL       string
    WordID    []int
    PageRank  float64
    OutLinks  []string
    Category  string 
}


var (
    wordToID map[string]int
    pages    []WikipediaPage
)

var (
    pagesByCategory map[string][]WikipediaPage
)


type SearchResult struct {
    URL           string  `json:"url"`
    ContentScore  float64 `json:"contentScore"`
    LocationScore float64 `json:"locationScore"`
    PageRankScore float64 `json:"pageRankScore"`
    TotalScore    float64 `json:"totalScore"`
}

// Function to initialize and start the HTTP server
func startServer() {
    http.HandleFunc("/search", searchHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// HTTP handler function for search queries
func searchHandler(w http.ResponseWriter, r *http.Request) {
    enableCors(&w) 

    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusOK)
        return
    }

    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    // Parse the request body to extract the search query
    var requestData struct {
        Word string `json:"word"`
    }
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&requestData); err != nil {
        http.Error(w, "Failed to decode request body", http.StatusBadRequest)
        return
    }

    query := requestData.Word
    results := performSearch(query)
    jsonResponse, err := json.Marshal(results)
    if err != nil {
        http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
        return
    }
 
    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonResponse)
}


func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*") // This is not safe for production
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}


func performSearch(query string) []SearchResult {
    queryWords := strings.Fields(query)
    contentScores := make([]float64, len(pages))
    locationScores := make([]float64, len(pages))

    for i, page := range pages {
        anyWordFound := false
        pageLocationScore := 0.0
        wordsNotFound := 0

        for _, wordID := range page.WordID {
            for _, qWord := range queryWords {
                if id, exists := wordToID[qWord]; exists && wordID == id {
                    contentScores[i]++
                }
            }
        }

        for _, qWord := range queryWords {
            wordFound := false
            for idx, wordID := range page.WordID {
                if id, exists := wordToID[qWord]; exists && wordID == id {
                    if !wordFound {
                        pageLocationScore += float64(idx) + 1
                        anyWordFound = true
                        wordFound = true
                        break
                    }
                }
            }
            if !wordFound {
                wordsNotFound++
            }
        }

        if anyWordFound && wordsNotFound == 0 {
            locationScores[i] = 1.0 / pageLocationScore
        } else {
            locationScores[i] = 0
        }
    }

    normalize(contentScores, false)
    normalize(locationScores, false)

    var searchResults []SearchResult
    for i, _ := range pages {
        if contentScores[i] > 0 {
            locationScore := locationScores[i] * 0.8
            pageRankScore := pages[i].PageRank * 0.5
            totalScore := contentScores[i] + locationScore + pageRankScore
            searchResults = append(searchResults, SearchResult{
                URL:           pages[i].URL,
                ContentScore:  contentScores[i],
                LocationScore: locationScore,
                PageRankScore: pageRankScore,
                TotalScore:    totalScore,
            })
        }
    }


    sort.Slice(searchResults, func(i, j int) bool {
        return searchResults[i].TotalScore > searchResults[j].TotalScore
    })

  

    return searchResults
}





func initializeIndex(basePath string) {
    wordToID = make(map[string]int)
    var idCounter int
    pagesByCategory = make(map[string][]WikipediaPage)

    wordsPath := filepath.Join(basePath, "Words")

    for _, folder := range []string{"Games", "Programming"} {
        var categoryPages []WikipediaPage
        wordFolderPath := filepath.Join(wordsPath, folder)

        wordFiles, err := os.ReadDir(wordFolderPath)
        if err != nil {
            log.Fatalf("Failed to read directory: %v", err)
        }

        for _, file := range wordFiles {
            fileName := file.Name()
            wordFilePath := filepath.Join(wordFolderPath, fileName)
            page := processFile(wordFilePath, &idCounter, folder)
            if page != nil {
                categoryPages = append(categoryPages, *page)
            }
        }

        pagesByCategory[folder] = categoryPages
        linkFolderPath := filepath.Join(basePath, "Links", folder)
        processLinks(linkFolderPath, folder, categoryPages)

        // Append pages from each category to the global pages slice
        pages = append(pages, categoryPages...)
    }
}
func processFile(filePath string, idCounter *int, category string) *WikipediaPage {
    content, err := os.ReadFile(filePath)
    if err != nil {
        log.Printf("Failed to read file: %v", err)
        return nil
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

    baseFileName := filepath.Base(filePath)

    return &WikipediaPage{
        URL:      baseFileName,
        WordID:   wordIDs,
        Category: category, 
    }
}


func processLinks(linkFolderPath, category string, categoryPages []WikipediaPage) {
    for i, page := range categoryPages {
        linkFileName := filepath.Base(page.URL)
        linkFilePath := filepath.Join(linkFolderPath, linkFileName)

        content, err := os.ReadFile(linkFilePath)
        if err != nil {
            log.Printf("Failed to read links file for %s: %v", linkFileName, err)
            continue
        }

        page.OutLinks = strings.Split(strings.TrimSpace(string(content)), "\n")
        categoryPages[i] = page

    }
    // Update the pagesByCategory map
    pagesByCategory[category] = categoryPages
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


func min(values []float64) float64 {
    minValue := math.MaxFloat64
    for _, v := range values {
        if v < minValue {
            minValue = v
        }
    }
    return minValue
}

func max(values []float64) float64 {
    maxValue := -math.MaxFloat64
    for _, v := range values {
        if v > maxValue {
            maxValue = v
        }
    }
    return maxValue
}


func calculatePageRank() {
    dampingFactor := 0.85
    numPages := float64(len(pages))

    // Initialize PageRank for each page
    for i := range pages {
        pages[i].PageRank = 1.0 / numPages
    }

    var newPageRanks = make([]float64, len(pages)) // Declaration moved outside of the loop

    // Calculate PageRank with the correct matching of OutLinks
    for iteration := 0; iteration < 20; iteration++ {
        for i, page := range pages {
            sum := 0.0
            for _, otherPage := range pages {
                // Ensure OutLinks match correctly
                if contains(otherPage.OutLinks, "/wiki/"+page.URL) {
                    sum += otherPage.PageRank / float64(len(otherPage.OutLinks))
                }
            }
            newPageRanks[i] = (1-dampingFactor)/numPages + dampingFactor*sum
        }
        for i := range pages {
            pages[i].PageRank = newPageRanks[i]
        }
    }

    // Normalize the PageRank scores after the iterations are complete
    maxPageRank := max(newPageRanks)
    for i := range pages {
        pages[i].PageRank /= maxPageRank // Normalization step
    }
}


// Helper function to check if a slice contains a specific string
func contains(slice []string, str string) bool {
    for _, s := range slice {
        if s == str {
            return true
        }
    }
    return false
}




func main() {
    relativePath := "./wikipedia"
    absolutePath, err := filepath.Abs(relativePath)
    if err != nil {
        log.Fatalf("Error getting absolute path: %v", err)
    }

    initializeIndex(absolutePath)

    startTime := time.Now()
    calculatePageRank()
    elapsedTime := time.Since(startTime)
    fmt.Printf("PageRank: %.2fs\n", elapsedTime.Seconds())


    fmt.Println("Listening on port: 8080")
 
   startServer()
}