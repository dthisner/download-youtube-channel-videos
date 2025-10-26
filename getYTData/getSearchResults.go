package getYTData

import (
	"download-youtube/models"
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

func (YT YouTubeChannel) GetSearchResultVideos() ([]SearchResult, error) {
	nextPageToken := ""
	totalFetched := 0

	client := &http.Client{}
	var videoData []SearchResult

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

		var res APIResponse
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

func (YT YouTubeChannel) GetSearchResultVideosDEBUG() []SearchResult {
	jsonFile, err := os.Open("TestData/YouTube-Data-Response.json")
	if err != nil {
		log.Print("Problem reading the debugFile", err)
	}

	var res APIResponse
	if err := json.NewDecoder(jsonFile).Decode(&res); err != nil {
		log.Print("Error decoding response:", err)
		return res.Items
	}

	return res.Items
}

// extractInformation takes the response JSON and saves it to our Video Struct
func (YT YouTubeChannel) ExtractSearchResultInfo(videoData []SearchResult) []models.Video {
	var currentEpisode = 1
	var currentSeason = 1

	startYear, err := strconv.Atoi(YT.EnvVar.SeasonStartYear)
	if err != nil {
		log.Printf("invalid SEASON_START_YEAR %q: %v", YT.EnvVar.SeasonStartYear, err)
		return nil
	}

	var videosData []models.Video

	for _, item := range videoData {
		var video models.Video

		if item.ID.VideoID == "" {
			log.Print("Item does not have an Video ID")
			continue
		}

		video = YT.SetVideoPlaylistDetails(video, item.Snippet)

		video.URL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.ID.VideoID)
		video.ID = item.ID.VideoID

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

		video = YT.FilePathAndName(video)

		videosData = append(videosData, video)
		printData(video)
	}

	return videosData
}

func (YT YouTubeChannel) SetVideoPlaylistDetails(video models.Video, snippet SearchSnippet) models.Video {
	video.ChannelTitle = snippet.ChannelTitle
	video.Description = snippet.Description
	video.Title = strings.Replace(snippet.Title, "\u0026#39;", "", -1)
	video.PublishedAt = snippet.PublishedAt
	video.ThumbnailURL = getThumbUrl(snippet.Thumbnails)

	return video
}
