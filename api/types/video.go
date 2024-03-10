package types

type Video struct {
	// Essential video data
	Id              string `json:"id"`
	SeriesId        string `json:"series_id,omitempty"`
	Episode         int    `json:"episode,omitempty"`
	Title           string `json:"title,omitempty"`
	NextEpisode     string `json:"next_episode_id,omitempty"`
	PreviousEpisode string `json:"previous_episode_id,omitempty"`

	// General data that is useful for debugging.
	FileName      string `json:"file_name,omitempty"`
	FileExtension string `json:"file_extension,omitempty"`
	UploadDate    string `json:"upload_date,omitempty"`
	LastModified  string `json:"last_modified,omitempty"`
}
