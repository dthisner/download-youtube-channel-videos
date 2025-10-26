package getYTData

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"download-youtube/models"
)

type YouTubeChannel struct {
	JsonFilePath        string
	EnvVar              models.EnvVar
	CurrentVideoData    []models.Video
	DownloadedVideoData []models.Video
}

// get gets all video data based on the channel ID. Will loop until it has recieved all of them or reached the maxResult
func (YT YouTubeChannel) GetData() {
	var existingVideos []models.Video
	var extractedInfo []models.Video

	jsonFile, err := os.Open(YT.JsonFilePath)
	if err != nil {
		log.Print("Did not find any data for: ", YT.JsonFilePath)
		log.Print("Will generate new file with data")
	} else {
		log.Print("Reading Existing File")
		jsonByte, _ := io.ReadAll(jsonFile)
		json.Unmarshal(jsonByte, &existingVideos)
	}
	defer jsonFile.Close()

	if YT.EnvVar.ChannelID != "" {
		newVideoData, err := YT.GetSearchResultVideos()
		if err != nil {
			log.Fatal(err)
		}

		// newVideoData, err := YT.getVideosDEBUG()
		// if err != nil {
		// 	log.Fatal(err)
		// }

		extractedInfo = YT.ExtractSearchResultInfo(newVideoData)

	} else if YT.EnvVar.PlaylistID != "" {
		newVideoData, err := YT.GetPlaylistSearchResultVideos()
		if err != nil {
			log.Fatal(err)
		}

		extractedInfo = YT.ExtractPlaylistSearchResultInfo(newVideoData)

	} else {
		log.Fatal("Neither ChannelID or Playlist ID has values")
	}

	videosToAdd := FindNewVideos(existingVideos, extractedInfo)
	existingVideos = append(existingVideos, videosToAdd...)

	marshalled, _ := json.Marshal(existingVideos)
	err = os.WriteFile(YT.JsonFilePath, marshalled, 0644)
	if err != nil {
		log.Print("Problem with writting JSON", err)
	}
	log.Printf("Successfully saved the JSON file to: %s", YT.JsonFilePath)
}

const (
	baseURL           = "https://www.googleapis.com/youtube/v3"
	searchEndpoint    = "search"
	playlistEndpoint  = "playlistItems"
	defaultPart       = "snippet,id"
	defaultOrder      = "date"
	defaultMaxResults = 50
)

// buildYouTubeURL constructs the API URL for either search or playlistItems endpoint.
func (YT YouTubeChannel) buildURL(pageToken string) (string, error) {
	var endpoint, idParam, idValue string
	if YT.EnvVar.ChannelID != "" {
		endpoint = searchEndpoint
		idParam = "channelId"
		idValue = YT.EnvVar.ChannelID
	} else {
		endpoint = playlistEndpoint
		idParam = "playlistId"
		idValue = YT.EnvVar.PlaylistID
	}

	return fmt.Sprintf("%s/%s?key=%s&%s=%s&part=%s&order=%s&maxResults=%d&pageToken=%s",
		baseURL, endpoint, YT.EnvVar.ApiKey, idParam, idValue, defaultPart, defaultOrder, defaultMaxResults, pageToken), nil
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

// normalizeTitle normalizes special characters in the title
func normalizeYouTubeTitle(input string) string {
	// Replace & with "and"
	input = strings.ReplaceAll(input, "&", "and")
	input = strings.ReplaceAll(input, " l ", "|")

	return input
}

func extractEpisodeInfo(input string) (string, string, string, error) {
	log.Print("Input: ", input)
	normalizedInput := normalizeYouTubeTitle(input)

	log.Print("normalizedInput: ", normalizedInput)
	// Define the regex pattern to match the title, episode, and season (case-insensitive for EPISODE and Season)
	// Assumes format: "Title | ... EPISODE <number> | Season <number>"
	pattern := `^(.*?)\s*\|\s*.*[Ee][Pp][Ii][Ss][Oo][Dd][Ee]\s+(\d+)\s*\|\s*[Ss][Ee][Aa][Ss][Oo][Nn]\s+(\d+)`
	re := regexp.MustCompile(pattern)

	// Find matches
	matches := re.FindStringSubmatch(normalizedInput)
	if len(matches) != 4 {
		return "", "", "", fmt.Errorf("invalid format: %s", input)
	}

	title := strings.TrimSpace(matches[1])
	episodeStr := matches[2]
	seasonStr := matches[3]

	// Add leading zero to season and episode if needed
	season := seasonStr
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
func printData(video models.Video) {
	log.Print("Title ", video.Title)
	log.Print("URL: ", video.URL)
	log.Print("Published At: ", video.PublishedAt)
	log.Print("Channel Title: ", video.ChannelTitle)
	log.Print("Video Thumnail URL: ", video.ThumbnailURL)
	log.Print()
}

func (YT YouTubeChannel) FilePathAndName(video models.Video) models.Video {
	video.Filepath = fmt.Sprintf("%s/Season %s/S%sE%s - %s",
		YT.EnvVar.ChannelName, video.Season, video.Season, video.Episode, video.Title)

	video.Filename = fmt.Sprintf("S%sE%s - %s",
		video.Season, video.Episode, video.Title)

	return video
}
