package types

type Series struct {
	Id           string `json:"id"`                      // uuid of the series
	Title        string `json:"title,omitempty"`         // title of the series
	Description  string `json:"description,omitempty"`   // description of the series
	EpisodeCount int    `json:"episode_count,omitempty"` // episode count of the series

	// General data that is useful for debugging.
	UploadDate   string `json:"upload_date,omitempty"`  // upload date of the series.
	LastModified string `json:"last_modifed,omitempty"` // last modified date of the series.
}
