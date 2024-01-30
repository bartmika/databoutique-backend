package pdfbuilder

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func cleanup(dataDirectoryPath string) {
	filepath.Walk(dataDirectoryPath,
		func(path string, info os.FileInfo, err error) (reterr error) {
			if info.Mode().IsRegular() {
				if len(path) > 3 {
					if path[len(path)-4:] == ".pdf" {
						log.Println("deleting:", path)
						os.Remove(path)
					}
				}
			}
			return
		})
}

func splitText(text string, maxChars int) []string {
	var lines []string
	words := strings.Fields(text)

	currentLine := ""
	for _, word := range words {
		if len(currentLine)+len(word) <= maxChars {
			// Word fits within the current line
			if currentLine != "" {
				currentLine += " " + word
			} else {
				currentLine += word
			}
		} else {
			// Word needs to be moved to the next line
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	// Add the last line
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

func getElementAtIndex(slice []string, index int) (string, bool) {
	if index < 0 || index >= len(slice) {
		return "", false
	}
	return slice[index], true
}
