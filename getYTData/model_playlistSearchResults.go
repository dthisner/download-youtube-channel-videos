package getYTData

// PlaylistItemListResponse represents the top-level response from the playlistItems endpoint.
type PlaylistItemListResponse struct {
	Kind          string         `json:"kind"`
	Etag          string         `json:"etag"`
	NextPageToken string         `json:"nextPageToken,omitempty"`
	PageInfo      PageInfo       `json:"pageInfo"`
	Items         []PlaylistItem `json:"items"`
}

// PageInfo contains pagination metadata.
type PageInfo struct {
	TotalResults   int `json:"totalResults"`
	ResultsPerPage int `json:"resultsPerPage"`
}

// PlaylistItem represents a single item in the playlist.
type PlaylistItem struct {
	Kind       string  `json:"kind"`
	Etag       string  `json:"etag"`
	PlaylistID string  `json:"id"` // Playlist item ID (string)
	Snippet    Snippet `json:"snippet"`
}

// Snippet contains details about the playlist item.
type Snippet struct {
	PublishedAt            string     `json:"publishedAt"`
	ChannelID              string     `json:"channelId"`
	Title                  string     `json:"title"`
	Description            string     `json:"description"`
	Thumbnails             Thumbnails `json:"thumbnails"`
	ChannelTitle           string     `json:"channelTitle"`
	PlaylistID             string     `json:"playlistId"`
	Position               int        `json:"position"`
	ResourceID             ResourceID `json:"resourceId"`
	VideoOwnerChannelTitle string     `json:"videoOwnerChannelTitle"`
	VideoOwnerChannelID    string     `json:"videoOwnerChannelId"`
}

// ResourceID contains the video ID.
type ResourceID struct {
	Kind    string `json:"kind"`
	VideoID string `json:"videoId"`
}

// Thumbnails contains thumbnail images for the video.
type Thumbnails struct {
	Default  Thumbnail `json:"default,omitempty"`
	Medium   Thumbnail `json:"medium,omitempty"`
	High     Thumbnail `json:"high,omitempty"`
	Standard Thumbnail `json:"standard,omitempty"`
	Maxres   Thumbnail `json:"maxres,omitempty"`
}
