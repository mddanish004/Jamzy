package music

type SearchResult struct {
	VideoID      string `json:"videoId"`
	Title        string `json:"title"`
	ChannelTitle string `json:"channelTitle"`
	ThumbnailURL string `json:"thumbnailUrl"`
	Duration     string `json:"duration"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

type VideoResponse struct {
	Result SearchResult `json:"result"`
}
