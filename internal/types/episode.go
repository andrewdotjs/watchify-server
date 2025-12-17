package types

type Episode struct {
	// Essential video data
	Id       string `json:"id"`
	ParentId string `json:"series_id,omitempty"`

	EpisodeNumber int    `json:"episode_number,omitempty"`
	Title         string `json:"title,omitempty"`
	Description   string `json:"description,omitempty"`

	// Urls for easier app navigation
	NextEpisode     map[string]string `json:"next_episode,omitempty"`
	PreviousEpisode map[string]string `json:"previous_episode,omitempty"`

	// General data that is useful for debugging.
	FileExtension string `json:"file_extension,omitempty"`
	FileName      string `json:"file_name"`
	UploadDate    string `json:"upload_date,omitempty"`
	LastModified  string `json:"last_modified,omitempty"`
}
