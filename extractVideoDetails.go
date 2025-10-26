package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"download-youtube/models"
)

type YouTubeChannel struct {
	jsonFilePath        string
	evnVar              models.EnvVar
	currentVideoData    []models.Video
	downloadedVideoData []models.Video
}

// get gets all video data based on the channel ID. Will loop until it has recieved all of them or reached the maxResult
func (YT YouTubeChannel) getData() {
	var existingVideos []models.Video

	jsonFile, err := os.Open(YT.jsonFilePath)
	if err != nil {
		log.Print("Did not find any data for: ", YT.jsonFilePath)
		log.Print("Will generate new file with data")
	} else {
		log.Print("Reading Existing File")
		jsonByte, _ := io.ReadAll(jsonFile)
		json.Unmarshal(jsonByte, &existingVideos)
	}
	defer jsonFile.Close()

	newVideoData, err := YT.getVideos()
	if err != nil {
		log.Fatal(err)
	}

	// newVideoData, err := YT.getVideosDEBUG()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	extractedInfo := YT.extractInformation(newVideoData)

	videosToAdd := FindNewVideos(existingVideos, extractedInfo)
	existingVideos = append(existingVideos, videosToAdd...)

	marshalled, _ := json.Marshal(existingVideos)
	err = os.WriteFile(YT.jsonFilePath, marshalled, 0644)
	if err != nil {
		log.Print("Problem with writting JSON", err)
	}
	log.Printf("Successfully saved the JSON file to: %s", YT.jsonFilePath)
}

func (YT YouTubeChannel) getVideosDEBUG() []models.SearchResult {
	jsonFile, err := os.Open("TestData/YouTube-Data-Response.json")
	if err != nil {
		log.Print("Problem reading the debugFile", err)
	}

	var res models.APIResponse
	if err := json.NewDecoder(jsonFile).Decode(&res); err != nil {
		log.Print("Error decoding response:", err)
		return res.Items
	}

	return res.Items
}

const (
	baseURL           = "https://www.googleapis.com/youtube/v3"
	searchEndpoint    = "search"
	playlistEndpoint  = "playlistItems"
	defaultPart       = "snippet,id"
	defaultOrder      = "date"
	defaultMaxResults = 50
)

func (YT YouTubeChannel) getVideos() ([]models.SearchResult, error) {
	nextPageToken := ""
	totalFetched := 0

	client := &http.Client{}
	var videoData []models.SearchResult

	for {
		url, err := YT.buildURL(nextPageToken)
		if err != nil {
			log.Printf("Error building URL: %v", err)
			return videoData, err
		}

		resp, err := client.Get(url)
		if err != nil {
			log.Printf("Error fetching URL %s: %v", url, err)
			return videoData, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			err := fmt.Errorf("received status code %d for URL %s: %s", resp.StatusCode, url, body)
			log.Print(err)
			return videoData, err
		}

		var res models.APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			log.Printf("Error decoding response from %s: %v", url, err)
			return videoData, err
		}

		videoData = append(videoData, res.Items...)
		totalFetched++

		if res.NextPageToken == "" || totalFetched >= defaultMaxResults {
			break
		}

		nextPageToken = res.NextPageToken
		// Optional delay to avoid hitting rate limits
		time.Sleep(5 * time.Second)
	}

	log.Printf("Total videos fetched: %d\n", len(videoData))
	return videoData, nil
}

// buildYouTubeURL constructs the API URL for either search or playlistItems endpoint.
func (YT YouTubeChannel) buildURL(pageToken string) (string, error) {
	var endpoint, idParam, idValue string
	if YT.evnVar.ChannelID != "" {
		endpoint = searchEndpoint
		idParam = "channelId"
		idValue = YT.evnVar.ChannelID
	} else {
		endpoint = playlistEndpoint
		idParam = "playlistId"
		idValue = YT.evnVar.PlaylistID
	}

	return fmt.Sprintf("%s/%s?key=%s&%s=%s&part=%s&order=%s&maxResults=%d&pageToken=%s",
		baseURL, endpoint, YT.evnVar.ApiKey, idParam, idValue, defaultPart, defaultOrder, defaultMaxResults, pageToken), nil
}

// normalizeTitle standardizes titles for consistent comparison
func normalizeTitle(title string) string {
	return strings.TrimSpace(strings.ToLower(title))
}

// FindNewVideos compares new videos against existing ones and returns new ones
func FindNewVideos(existing, newVideos []models.Video) []models.Video {
	// Build a map of existing titles for O(1) lookups
	titleMap := make(map[string]struct{})
	for _, video := range existing {
		titleMap[normalizeTitle(video.Title)] = struct{}{}
	}

	// Collect new videos that don't exist in the map
	var videosToAdd []models.Video
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
func (YT YouTubeChannel) extractInformation(videoData []models.SearchResult) []models.Video {
	var currentEpisode = 1
	var currentSeason = 1
	var tvShowName = YT.evnVar.ChannelName

	startYear, err := strconv.Atoi(YT.evnVar.SeasonStartYear)
	if err != nil {
		log.Printf("invalid SEASON_START_YEAR %q: %v", YT.evnVar.SeasonStartYear, err)
		return nil
	}

	var videosData []models.Video

	for _, item := range videoData {
		var video models.Video

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

		// Splitting between getting ALL videos from a channel VS getting something that have seasons and episodes in the name
		if strings.Contains(normalizeTitle(video.Title), "episode") && strings.Contains(normalizeTitle(video.Title), "season") {
			video.Title, video.Episode, video.Season, err = extractEpisodeInfo(video.Title)
			if err != nil {
				log.Print("Problem with extracting episode info", err)
			}
		} else {
			season := yearPub - startYear + 1
			video.Season = fmt.Sprintf("%02d", season)

			log.Printf("Season: %d CurrentSeasson: %d", season, currentSeason)
			if season == currentSeason {
				currentEpisode++
			} else {
				currentEpisode = 1
				currentSeason = season
			}

			video.Episode = fmt.Sprintf("%02d", currentEpisode)
		}

		video.Filepath = fmt.Sprintf("%s/Season %s/S%sE%s - %s",
			tvShowName, video.Season, video.Season, video.Episode, video.Title)

		video.Filename = fmt.Sprintf("S%sE%s - %s",
			video.Season, video.Episode, video.Title)

		videosData = append(videosData, video)
		printData(video)
	}

	return videosData
}

func extractEpisodeInfo(input string) (string, string, string, error) {
	// Define the regex pattern to match the title, episode, and season (case-insensitive for EPISODE and Season)
	// Assumes format: "Title | ... EPISODE <number> | Season <number>"
	pattern := `^(.*?)\s*\|\s*.*[Ee][Pp][Ii][Ss][Oo][Dd][Ee]\s+(\d+)\s*\|\s*[Ss][Ee][Aa][Ss][Oo][Nn]\s+(\d+)`
	re := regexp.MustCompile(pattern)

	// Find matches
	matches := re.FindStringSubmatch(input)
	if len(matches) != 4 {
		return "", "", "", fmt.Errorf("invalid format: %s", input)
	}

	title := strings.TrimSpace(matches[1])
	episodeStr := matches[2]
	seasonStr := matches[3]

	// Add leading zero to season and episode if needed
	season := episodeStr
	if len(seasonStr) == 1 {
		season = "0" + seasonStr
	}
	episode := episodeStr
	if len(episodeStr) == 1 {
		episode = "0" + episodeStr
	}

	return title, season, episode, nil
}

// getThumbUrl looks for the biggest thumbnail and saves that as the best option for Thumbnail URL
func getThumbUrl(thumbnails map[string]models.Thumbnail) string {
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
func printData(video models.Video) {
	log.Print("Title ", video.Title)
	log.Print("URL: ", video.URL)
	log.Print("Published At: ", video.PublishedAt)
	log.Print("Channel Title: ", video.ChannelTitle)
	log.Print("Video Thumnail URL: ", video.ThumbnailURL)
	log.Print()
}
