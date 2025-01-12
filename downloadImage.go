package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// downloadImage downloads an image from the given URL and saves it to the specified file
func downloadImage(video Video) error {
	log.Printf("Downloading Thumbnail to: %s", video.Title)

	response, err := http.Get(video.ThumbnailURL)
	if err != nil {
		return fmt.Errorf("failed to fetch the image: %v", err)
	}
	defer response.Body.Close()

	filePath := fmt.Sprintf("%s-thumb.jpg", video.Filepath)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("failed to save the image: %v", err)
	}

	return nil
}
