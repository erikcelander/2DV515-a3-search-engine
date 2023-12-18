package main

import (
	"math"
	"runtime"
	"sort"
	"strings"
	"sync"
)

var (
	wordToID map[string]int
	pages    []WikipediaPage
)

type WikipediaPage struct {
	URL       string
	WordID    []int
	PageRank  float64
	OutLinks  []string
	Category  string
}
type SearchResult struct {
	URL           string  `json:"url"`
	ContentScore  float64 `json:"contentScore"`
	LocationScore float64 `json:"locationScore"`
	PageRankScore float64 `json:"pageRankScore"`
	TotalScore    float64 `json:"totalScore"`
}

var searchResultPool = sync.Pool{
	New: func() interface{} {
		return new(SearchResult)
	},
}

func performSearch(query string) []SearchResult {
	queryWords := strings.Fields(query)
	contentScores := make([]float64, len(pages))
	locationScores := make([]float64, len(pages))

	numWorkers := runtime.NumCPU()
	workChan := make(chan int, len(pages))
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func() {
					defer wg.Done()
					for pageIndex := range workChan {
							page := pages[pageIndex]
							pageLocationScore := 0.0

							// content/word frequency score
							for _, wordID := range page.WordID {
									for _, qWord := range queryWords {
											if id, exists := wordToID[qWord]; exists && wordID == id {
													contentScores[pageIndex]++
													break
											}
									}
							}

							// location score
							for _, qWord := range queryWords {
									wordFound := false
									for idx, wordID := range page.WordID {
											if id, exists := wordToID[qWord]; exists && wordID == id {
													pageLocationScore += float64(idx) + 1
													wordFound = true
													break
											}
									}
									if !wordFound {
											pageLocationScore += 100000 
									}
							}
							locationScores[pageIndex] = pageLocationScore
					}
			}()
	}

	for i := 0; i < len(pages); i++ {
			workChan <- i
	}
	close(workChan)
	wg.Wait()

	normalize(locationScores, true)
	normalize(contentScores, false)

	var searchResults []SearchResult
	for i := range pages {
			if contentScores[i] > 0 {
					result := searchResultPool.Get().(*SearchResult)
					result.URL = pages[i].URL
					result.ContentScore = contentScores[i]
					result.LocationScore = locationScores[i] * 0.8
					result.PageRankScore = pages[i].PageRank * 0.5
					result.TotalScore = result.ContentScore + result.LocationScore + result.PageRankScore
					searchResults = append(searchResults, *result)
					searchResultPool.Put(result)
			}
	}

	sort.Slice(searchResults, func(i, j int) bool {
			return searchResults[i].TotalScore > searchResults[j].TotalScore
	})
	return searchResults
}


func calculatePageRank() {
	dampingFactor := 0.85
	numPages := float64(len(pages))

	for i := range pages {
			pages[i].PageRank = 1.0 / numPages
	}

	var newPageRanks = make([]float64, len(pages))

	for iteration := 0; iteration < 20; iteration++ {
			for i, page := range pages {
					sum := 0.0
					for _, otherPage := range pages {
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

	maxPageRank := max(newPageRanks)
	for i := range pages {
			pages[i].PageRank /= maxPageRank
	}
}

func normalize(scores []float64, smallIsBetter bool) {
	if smallIsBetter {
    minVal := min(scores)
    for i := range scores {
        if scores[i] > 0.00001 {
            scores[i] = minVal / scores[i]
        } else {
            scores[i] = minVal / 0.00001
        }
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

func contains(slice []string, str string) bool {
	for _, s := range slice {
			if s == str {
					return true
			}
	}
	return false
}