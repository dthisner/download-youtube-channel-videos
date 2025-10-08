package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var JSON_FILE_NAME = "all_channel_videos.json"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	channelID := os.Getenv("YT_CHANNEL_ID")
	getYouTubeChannelVideos(channelID)

	// downloadAndOrganizeVideos()
}

func downloadAndOrganizeVideos() {
	jsonFile, err := os.Open(JSON_FILE_NAME)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	var videos []Video

	jsonByte, _ := io.ReadAll(jsonFile)
	json.Unmarshal(jsonByte, &videos)

	for i, video := range videos {
		if video.Downloaded {
			log.Printf("Video %s has already been downloaded", video.Title)
			continue
		}

		log.Printf("Downloading Video %s", video.Title)

		checkSeasonFolderExist(video.Season)
		generateEpisodeNfo(video)
		err = downloadImage(video)
		if err != nil {
			log.Print(err)
		} else {
			log.Printf("Successfully downloaded image to: %s", video.Filepath)
			videos[i].ImageSaved = true
		}

		err = downloadVideo(video)
		if err != nil {
			log.Print(err)
		} else {
			videos[i].Downloaded = true
		}

		videosJSON, _ := json.Marshal(videos)
		err = os.WriteFile(JSON_FILE_NAME, videosJSON, 0644)
		if err != nil {
			log.Print("Problem with writting JSON", err)
		}
		log.Print("Successfully downloaded, merged the video and updated the JSON file")
	}
}

// checkSeasonFolderExist creates the season folder if it's missing
func checkSeasonFolderExist(season string) error {
	var tvShowName = os.Getenv("YT_SHOW_NAME")
	folderPath := fmt.Sprintf("%s/Season %s", tvShowName, season)

	_, err := os.Stat(folderPath)

	if os.IsNotExist(err) {
		// Create the folder (and any necessary parent directories)
		err := os.MkdirAll(folderPath, 0755) // 0755 is the permission mode
		if err != nil {
			return fmt.Errorf("error creating folder: %s", err)

		}
		log.Print("Folder created successfully: ", folderPath)
	} else {
		log.Print("Folder already exists:", folderPath)
	}

	return nil
}

// getYouTubeChannelVideos gets all youtube videos based on the channel ID. Will loop until it has recieved all of them or reached the maxResult
func getYouTubeChannelVideos(channelID string) {
	apiKey := os.Getenv("YT_API_KEY")
	baseURL := "https://www.googleapis.com/youtube/v3/search"

	nextPageToken := ""
	maxResults := 500 // Set your limit (max = 500)
	totalFetched := 0

	var videos []Video

	for {
		url := fmt.Sprintf("%s?key=%s&channelId=%s&part=snippet,id&order=date&maxResults=20&pageToken=%s", baseURL, apiKey, channelID, nextPageToken)

		resp, err := http.Get(url)
		if err != nil {
			log.Print("Error:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Error: received status code %d\n", resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)
			log.Printf("Response body: %s\n", body)
			return
		}

		var res APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			log.Print("Error decoding response:", err)
			return
		}

		extractInformation(res.Items, &videos)
		totalFetched++

		if res.NextPageToken == "" || totalFetched >= maxResults {
			break
		}

		nextPageToken = res.NextPageToken
		// Optional delay to avoid hitting rate limits
		time.Sleep(5 * time.Second)
	}

	fmt.Printf("Total videos fetched: %d\n", len(videos))

	videosJSON, _ := json.Marshal(videos)
	err := os.WriteFile(JSON_FILE_NAME+"updated", videosJSON, 0644)
	if err != nil {
		log.Print("Problem with writting JSON", err)
	}

	log.Print("Successfully saved the JSON file")
}

// extractInformation takes the response JSON and saves it to our Video Struct
func extractInformation(res []SearchResult, videos *[]Video) {
	var currentEpisode = 1
	var currentSeason = 1
	var tvShowName = os.Getenv("YT_SHOW_NAME")

	yearToSeason := map[string]int{
		"2016": 1, "2017": 2, "2018": 3, "2019": 4,
		"2020": 5, "2021": 6, "2022": 7, "2023": 8,
		"2024": 9, "2025": 10, "2026": 11,
	}

	for _, item := range res {
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

		year := video.PublishedAt[0:4]
		if season, exists := yearToSeason[year]; exists {
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

		} else {
			video.Season = "00" // Default season or you could skip it
		}

		*videos = append(*videos, video)
		printData(video)
	}
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
