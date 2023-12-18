package main

import (
	"os"
	"path/filepath"
	"strings"
	"log"
)


var (
	pagesByCategory map[string][]WikipediaPage
)

var invertedIndex map[int][]int 

func initializeIndex(basePath string) {
	wordToID = make(map[string]int)
	var idCounter int
	pagesByCategory = make(map[string][]WikipediaPage)
	invertedIndex = make(map[int][]int)

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
							for _, wordID := range page.WordID {
									invertedIndex[wordID] = append(invertedIndex[wordID], len(categoryPages)-1)
							}
					}
			}

			pagesByCategory[folder] = categoryPages
			linkFolderPath := filepath.Join(basePath, "Links", folder)
			processLinks(linkFolderPath, folder, categoryPages)
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
	pagesByCategory[category] = categoryPages
}
