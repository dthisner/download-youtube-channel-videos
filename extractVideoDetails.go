package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// get gets all video data based on the channel ID. Will loop until it has recieved all of them or reached the maxResult
func updateYTChannelData() {
	var existingVideos []Video

	jsonFile, err := os.Open(JSON_FILE_NAME)
	if err != nil {
		log.Print("Did not find any data for: ", JSON_FILE_NAME)
		log.Print("Will download all new data")
	} else {
		log.Print("Reading Existing File")
		jsonByte, _ := io.ReadAll(jsonFile)
		json.Unmarshal(jsonByte, &existingVideos)
	}
	defer jsonFile.Close()

	newVideoData := getYTChannelVideos()
	// newVideoData := getYTChannelVideosDEBUG()

	extractedInfo := extractInformation(newVideoData)

	videosToAdd := FindNewVideos(existingVideos, extractedInfo)
	existingVideos = append(existingVideos, videosToAdd...)

	marshalled, _ := json.Marshal(existingVideos)
	err = os.WriteFile(JSON_FILE_NAME, marshalled, 0644)
	if err != nil {
		log.Print("Problem with writting JSON", err)
	}
	log.Printf("Successfully saved the JSON file to: %s", JSON_FILE_NAME)
}

func getYTChannelVideosDEBUG() []SearchResult {
	jsonFile, err := os.Open("TestData/YouTube-Data-Response.json")
	if err != nil {
		log.Print("Problem reading the debugFile", err)

	}

	var res APIResponse
	if err := json.NewDecoder(jsonFile).Decode(&res); err != nil {
		log.Print("Error decoding response:", err)
		return res.Items
	}

	fmt.Printf("Total videos fetched: %d\n", len(res.Items))
	return res.Items
}

func getYTChannelVideos() []SearchResult {
	apiKey := os.Getenv("YT_API_KEY")
	channelID := os.Getenv("YT_CHANNEL_ID")
	baseURL := "https://www.googleapis.com/youtube/v3/search"

	nextPageToken := ""
	maxResults := 50 // Set your limit (max = 50)
	totalFetched := 0

	var videoData []SearchResult

	for {
		url := fmt.Sprintf("%s?key=%s&channelId=%s&part=snippet,id&order=date&maxResults=20&pageToken=%s", baseURL, apiKey, channelID, nextPageToken)

		resp, err := http.Get(url)
		if err != nil {
			log.Print("Error:", err)
			return videoData
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Error: received status code %d\n", resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)
			log.Printf("Response body: %s\n", body)
			return videoData
		}

		var res APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			log.Print("Error decoding response:", err)
			return videoData
		}

		videoData = append(videoData, res.Items...)
		totalFetched++

		if res.NextPageToken == "" || totalFetched >= maxResults {
			break
		}

		nextPageToken = res.NextPageToken
		// Optional delay to avoid hitting rate limits
		time.Sleep(5 * time.Second)
	}

	log.Printf("Total videos fetched: %d\n", len(videoData))
	return videoData
}

// normalizeTitle standardizes titles for consistent comparison
func normalizeTitle(title string) string {
	return strings.TrimSpace(strings.ToLower(title))
}

// FindNewVideos compares new videos against existing ones and returns new ones
func FindNewVideos(existing, newVideos []Video) []Video {
	// Build a map of existing titles for O(1) lookups
	titleMap := make(map[string]struct{})
	for _, video := range existing {
		titleMap[normalizeTitle(video.Title)] = struct{}{}
	}

	// Collect new videos that don't exist in the map
	var videosToAdd []Video
	for _, video := range newVideos {
		log.Printf("Does %s exist already?", video.Title)
		if _, exists := titleMap[normalizeTitle(video.Title)]; !exists {
			log.Printf("%s Does NOT exist, lets add it!", video.Title)
			videosToAdd = append(videosToAdd, video)
		}
	}
	return videosToAdd
}

// extractInformation takes the response JSON and saves it to our Video Struct
func extractInformation(videoData []SearchResult) []Video {
	var currentEpisode = 1
	var currentSeason = 1
	var tvShowName = os.Getenv("YT_SHOW_NAME")

	startYear, err := strconv.Atoi(os.Getenv("SEASON_START_YEAR"))
	if err != nil {
		log.Printf("invalid SEASON_START_YEAR %q: %v", os.Getenv("SEASON_START_YEAR"), err)
		return nil
	}

	var videosData []Video

	for _, item := range videoData {
		var video Video

		if item.ID.VideoID == "" {
			log.Print("Item does not have an Video ID")
			continue
		}

		video.ThumbnailURL = getThumbUrl(item.Snippet.Thumbnails)
		video.Title = strings.Replace(item.Snippet.Title, "\u0026#39;", "", -1)
		video.PublishedAt = item.Snippet.PublishedAt
		video.URL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.ID.VideoID)
		video.ID = item.ID.VideoID
		video.ChannelTitle = item.Snippet.ChannelTitle
		video.Description = item.Snippet.Description

		yearPub, err := strconv.Atoi(video.PublishedAt[0:4])
		if err != nil {
			log.Printf("invalid yearPublished %q: %v", video.PublishedAt[0:4], err)
			continue
		}

		season := yearPub - startYear + 1
		video.Season = fmt.Sprintf("%02d", season)

		log.Printf("Season: %d CurrentSeasson: %d", season, currentSeason)
		if season == currentSeason {
			currentEpisode++
			video.Episode = fmt.Sprintf("%02d", currentEpisode)
		} else {
			currentEpisode = 1
			video.Episode = fmt.Sprintf("%02d", currentEpisode)
			currentSeason = season
		}

		video.Filepath = fmt.Sprintf("%s/Season %s/S%sE%s - %s", tvShowName, video.Season, video.Season, video.Episode, video.Title)
		video.Filename = fmt.Sprintf("S%sE%s - %s", video.Season, video.Episode, video.Title)

		videosData = append(videosData, video)
		printData(video)
	}

	return videosData
}

// getThumbUrl looks for the biggest thumbnail and saves that as the best option for Thumbnail URL
func getThumbUrl(thumbnails map[string]Thumbnail) string {
	var biggestSize = 0
	var thumbnailURL string
	for _, thumb := range thumbnails {
		if thumb.Height > biggestSize {
			biggestSize = thumb.Height
			thumbnailURL = thumb.URL
			log.Printf("biggestSize is now: %d with URL: %s", biggestSize, thumbnailURL)
		}
	}
	return thumbnailURL
}

// printData Just prints some of the selected values
func printData(video Video) {
	log.Print("Title ", video.Title)
	log.Print("URL: ", video.URL)
	log.Print("Published At: ", video.PublishedAt)
	log.Print("Channel Title: ", video.ChannelTitle)
	log.Print("Video Thumnail URL: ", video.ThumbnailURL)
	log.Print()
}
