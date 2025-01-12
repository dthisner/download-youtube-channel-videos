package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kkdai/youtube/v2"
	"golang.org/x/exp/maps"
)

type Video struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnailUrl"`
	PublishedAt  string `json:"publishedAt"`
	ChannelTitle string `json:"channelTitle"`
	Season       int    `json:"season"`
	Episode      int    `json:"episode"`
	Downloaded   bool   `json:"downloaded"`
}

type SearchResult struct {
	Kind string `json:"kind"`
	ETag string `json:"etag"`
	ID   struct {
		Kind       string `json:"kind"`
		VideoID    string `json:"videoId,omitempty"`
		ChannelID  string `json:"channelId,omitempty"`
		PlaylistID string `json:"playlistId,omitempty"`
	} `json:"id"`
	Snippet struct {
		PublishedAt          string               `json:"publishedAt"`
		ChannelID            string               `json:"channelId"`
		Title                string               `json:"title"`
		Description          string               `json:"description"`
		Thumbnails           map[string]Thumbnail `json:"thumbnails"`
		ChannelTitle         string               `json:"channelTitle"`
		LiveBroadcastContent string               `json:"liveBroadcastContent"`
	} `json:"snippet"`
}

type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type APIResponse struct {
	Kind          string         `json:"kind"`
	ETag          string         `json:"etag"`
	Items         []SearchResult `json:"items"`
	NextPageToken string         `json:"nextPageToken,omitempty"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// https://www.youtube.com/channel/UC1KmNKYC1l0stjctkGswl6g
	// getYouTubeChannelVideos("UC1KmNKYC1l0stjctkGswl6g")

	jsonFile, err := os.Open("all_channel_videos.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	var videos []Video

	jsonByte, _ := io.ReadAll(jsonFile)
	json.Unmarshal(jsonByte, &videos)

	for i, video := range videos {
		log.Print("Video Title", video.Title)

		if video.Downloaded {
			log.Printf("Video %s has already been downloaded", video.Title)
			continue
		}

		err = downloadVideo(video.URL)
		if err != nil {
			log.Print(err)
		} else {
			videos[i].Downloaded = true
		}

		break
	}

	videosJSON, _ := json.Marshal(videos)
	err = os.WriteFile("channel_videos_seasons.json", videosJSON, 0644)
	if err != nil {
		log.Print("Problem with writting JSON", err)
	}
}

func downloadVideo(videoURL string) error {
	client := youtube.Client{}

	video, err := client.GetVideo(videoURL)
	if err != nil {
		return fmt.Errorf("error fetching video info: %v", err)
	}

	var videoFormat, audioFormat *youtube.Format
	for _, format := range video.Formats {
		if format.QualityLabel == "720p" && format.AudioChannels == 0 {
			videoFormat = &format
		}
		if format.AudioChannels > 0 && format.AudioQuality == "AUDIO_QUALITY_MEDIUM" {
			audioFormat = &format
		}
		if videoFormat != nil && audioFormat != nil {
			break
		}
	}

	if videoFormat == nil || audioFormat == nil {
		log.Fatal("Suitable video or audio format not found")
	}

	videoFileName := strings.ReplaceAll(video.Title, " ", "_") + "_video.mp4"
	err = downloadStream(client, video, videoFormat, videoFileName)
	if err != nil {
		return err
	}
	audioFileName := strings.ReplaceAll(video.Title, " ", "_") + "_audio.mp4"
	err = downloadStream(client, video, audioFormat, audioFileName)
	if err != nil {
		return err
	}

	err = mergeAudioVideo(video.Title, videoFileName, audioFileName)
	if err != nil {
		return err
	}

	return nil
}

func mergeAudioVideo(title, videoFileName, audioFileName string) error {
	mergedFileName := strings.ReplaceAll(title, " ", "_") + "_merged.mp4"
	ffmpegCmd := exec.Command("ffmpeg", "-i", videoFileName, "-i", audioFileName, "-c:v", "copy", "-c:a", "aac", "-strict", "experimental", mergedFileName)

	log.Print("Merging audio and video...")
	ffmpegCmd.Stdout = os.Stdout
	ffmpegCmd.Stderr = os.Stderr

	if err := ffmpegCmd.Run(); err != nil {
		return fmt.Errorf("error merging audio and video: %s", err)
	}

	log.Print("Merging complete:", mergedFileName)

	return nil
}

func downloadStream(client youtube.Client, video *youtube.Video, format *youtube.Format, filename string) error {
	stream, _, err := client.GetStream(video, format)
	if err != nil {
		return fmt.Errorf("WithAudioChannels - %s", err)
	}
	defer stream.Close()

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create - %s", err)
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		return fmt.Errorf("copy - %s", err)
	}

	fmt.Printf("Stream downloaded successfully: %s\n", filename)

	return nil
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

		var apiResponse APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			log.Print("Error decoding response:", err)
			return
		}

		yearToSeason := map[string]int{
			"2016": 1, "2017": 2, "2018": 3, "2019": 4,
			"2020": 5, "2021": 6, "2022": 7, "2023": 8,
			"2024": 9, "2025": 10, "2026": 11,
		}

		var currentEpisode = 1
		var currentSeason int

		for _, item := range apiResponse.Items {
			var video Video

			if item.ID.VideoID == "" {
				log.Print("Item does not have an Video ID")
				continue
			}

			keys := maps.Keys(item.Snippet.Thumbnails)
			video.ThumbnailURL = item.Snippet.Thumbnails[keys[0]].URL
			video.Title = item.Snippet.Title
			video.PublishedAt = item.Snippet.PublishedAt
			video.URL = item.ID.VideoID
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

			videos = append(videos, video)
			printData(video)
		}

		totalFetched++
		if apiResponse.NextPageToken == "" || totalFetched >= maxResults {
			break
		}

		nextPageToken = apiResponse.NextPageToken
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

func printData(video Video) {
	log.Print("Title:", video.Title)
	log.Print("URL:", video.URL)
	log.Print("Published At:", video.PublishedAt)
	log.Print("Channel Title:", video.ChannelTitle)
	log.Print("Video Thumnail URL:", video.ThumbnailURL)
	log.Print()
}
