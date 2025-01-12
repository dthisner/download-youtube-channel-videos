package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/exp/maps"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	channelID := os.Getenv("YT_CHANNEL_ID")
	getYouTubeChannelVideos(channelID)

	// jsonFile, err := os.Open("all_channel_videos.json")
	// if err != nil {
	// 	panic(err)
	// }
	// defer jsonFile.Close()

	// var videos []Video

	// jsonByte, _ := io.ReadAll(jsonFile)
	// json.Unmarshal(jsonByte, &videos)

	// for i, video := range videos {
	// 	log.Print("Video Title", video.Title)

	// 	if video.Downloaded {
	// 		log.Printf("Video %s has already been downloaded", video.Title)
	// 		continue
	// 	}

	// 	err = downloadVideo(video.URL)
	// 	if err != nil {
	// 		log.Print(err)
	// 	} else {
	// 		videos[i].Downloaded = true
	// 	}

	// 	break
	// }

	// videosJSON, _ := json.Marshal(videos)
	// err = os.WriteFile("channel_videos_seasons.json", videosJSON, 0644)
	// if err != nil {
	// 	log.Print("Problem with writting JSON", err)
	// }
}

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

		extractInformation(res, &videos)
		totalFetched++

		if res.NextPageToken == "" || totalFetched >= maxResults {
			break
		}

		nextPageToken = res.NextPageToken
		// Optional delay to avoid hitting rate limits
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("Total videos fetched: %d\n", len(videos))

	videosJSON, _ := json.Marshal(videos)
	err := os.WriteFile("all_channel_videos.json", videosJSON, 0644)
	if err != nil {
		log.Print("Problem with writting JSON", err)
	}

	log.Print("Successfully saved the JSON file")
}

func extractInformation(res APIResponse, videos *[]Video) {
	var currentEpisode = 1
	var currentSeason int

	yearToSeason := map[string]int{
		"2016": 1, "2017": 2, "2018": 3, "2019": 4,
		"2020": 5, "2021": 6, "2022": 7, "2023": 8,
		"2024": 9, "2025": 10, "2026": 11,
	}

	for _, item := range res.Items {
		var video Video

		if item.ID.VideoID == "" {
			log.Print("Item does not have an Video ID")
			continue
		}

		keys := maps.Keys(item.Snippet.Thumbnails)
		video.ThumbnailURL = item.Snippet.Thumbnails[keys[0]].URL
		video.Title = item.Snippet.Title
		video.PublishedAt = item.Snippet.PublishedAt
		video.URL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.ID.VideoID)
		video.ID = item.ID.VideoID
		video.ChannelTitle = item.Snippet.ChannelTitle
		video.Description = item.Snippet.Description

		year := video.PublishedAt[0:4]
		if season, exists := yearToSeason[year]; exists {
			video.Season = season

			log.Printf("Season: %d CurrentSeasson: %d", season, currentSeason)
			if season == currentSeason {
				currentEpisode++
				video.Episode = currentEpisode
			} else {
				currentEpisode = 1
				video.Episode = currentEpisode
				currentSeason = season
			}

		} else {
			video.Season = 0 // Default season or you could skip it
		}

		*videos = append(*videos, video)
		printData(video)
	}
}

func printData(video Video) {
	log.Print("Title:", video.Title)
	log.Print("URL:", video.URL)
	log.Print("Published At:", video.PublishedAt)
	log.Print("Channel Title:", video.ChannelTitle)
	log.Print("Video Thumnail URL:", video.ThumbnailURL)
	log.Print()
}
