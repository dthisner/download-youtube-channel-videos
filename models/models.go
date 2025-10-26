package models

import (
	"fmt"
	"strings"
)

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

type NFOEpisodeDetails struct {
	CreationDate     string
	Version          string
	Title            string
	OriginalTitle    string
	ShowTitle        string
	Season           string
	Episode          string
	DisplaySeason    string
	DisplayEpisode   string
	ID               string
	Ratings          string
	UserRating       string
	Plot             string
	Runtime          string
	MPAA             string
	Premiered        string
	Aired            string
	Watched          string
	PlayCount        string
	Trailer          string
	DateAdded        string
	EpBookmark       string
	Code             string
	VideoCodec       string
	VideoAspect      string
	VideoWidth       string
	VideoHeight      string
	VideoDuration    string
	StereoMode       string
	Source           string
	OriginalFilename string
	UserNote         string
	GroupEpisode     string
	GroupID          string
	GroupName        string
	GroupSeason      string
}
