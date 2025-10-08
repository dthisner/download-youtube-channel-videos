package main

import (
	"fmt"
	"strings"
)

type Video struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	URL          string `json:"url"`
	ID           string `json:"id"`
	ThumbnailURL string `json:"thumbnailUrl"`
	PublishedAt  string `json:"publishedAt"`
	ChannelTitle string `json:"channelTitle"`
	Season       string `json:"season"`
	Episode      string `json:"episode"`
	Downloaded   bool   `json:"downloaded"`
	ImageSaved   bool   `json:"imageSaved"`
	Filename     string `json:"filename"`
	Filepath     string `json:"filepath"`
	Error        string `json:"error"`
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

type EnvVar struct {
	ApiKey          string
	ChannelID       string
	PlaylistID      string
	ChannelName     string
	SeasonStartYear string
}

func (e EnvVar) Validate() error {
	var missingFields []string

	if e.ApiKey == "" {
		missingFields = append(missingFields, "YT_API_KEY")
	}
	if e.ChannelID == "" && e.PlaylistID == "" {
		missingFields = append(missingFields, "YT_CHANNEL_ID or YT_PLAYLIST_ID")
	}
	if e.ChannelName == "" {
		missingFields = append(missingFields, "YT_CHANNEL_NAME")
	}
	if e.SeasonStartYear == "" {
		missingFields = append(missingFields, "SEASON_START_YEAR")
	}

	// If there are missing fields, return a combined error
	if len(missingFields) > 0 {
		return fmt.Errorf("missing environment variables: %s", strings.Join(missingFields, ", "))
	}

	return nil
}
