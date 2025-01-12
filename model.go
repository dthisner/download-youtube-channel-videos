package main

type Video struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	URL          string `json:"url"`
	ID           string `json:"id"`
	ThumbnailURL string `json:"thumbnailUrl"`
	PublishedAt  string `json:"publishedAt"`
	ChannelTitle string `json:"channelTitle"`
	Season       int    `json:"season"`
	Episode      int    `json:"episode"`
	Downloaded   bool   `json:"downloaded"`
	Filename     string `json:"filename"`
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
