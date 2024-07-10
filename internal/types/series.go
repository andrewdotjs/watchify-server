package types

type Series struct {
	Id      string `json:"id"` // uuid of the series
	CoverId string `json:"cover_id"`

	Title        string `json:"title,omitempty"`       // title of the series
	Description  string `json:"description,omitempty"` // description of the series
	EpisodeCount int    `json:"-"`                     // episode count of the series
	Hidden       bool   `json:"hidden"`

	Episodes map[string]any `json:"episodes,omitempty"`
	// 	EXAMPLE:
	//  "episodes": {
	//		"count": 0,
	//		"url": "example.com/api/v1/{series_id}/episodes"
	//  }

	Cover map[string]any `json:"cover,omitempty"`
	// 	EXAMPLE:
	//  "cover": {
	//		"exists": true,
	//		"url": "example.com/api/v1/{series_id}/cover"
	//  }

	// General data that is useful for debugging.
	UploadDate   string `json:"upload_date,omitempty"`  // upload date of the series.
	LastModified string `json:"last_modifed,omitempty"` // last modified date of the series.
}

type Episode struct {
	// Essential video data
	Id       string `json:"id"`
	SeriesId string `json:"series_id,omitempty"`

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

type SeriesCover struct {
	// Essential cover data
	Id       string `json:"id"`
	SeriesId string `json:"series_id"`

	// General data that is useful for debugging.
	FileName      string `json:"file_name"`
	FileExtension string `json:"file_extension"`
	UploadDate    string `json:"upload_date"`
}
