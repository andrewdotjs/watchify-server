package types

type Video struct {
	// Essential video data
	Id            string `json:"id"`
	SeriesId      string `json:"series_id,omitempty"`
	EpisodeNumber int    `json:"episode_number,omitempty"`
	Title         string `json:"title,omitempty"`

	// Urls for easier app navigation
	NextEpisode     map[string]string `json:"next_episode,omitempty"`
	PreviousEpisode map[string]string `json:"previous_episode,omitempty"`

	// General data that is useful for debugging.
	FileExtension string `json:"file_extension,omitempty"`
	UploadDate    string `json:"upload_date,omitempty"`
	LastModified  string `json:"last_modified,omitempty"`
}
