package models

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
