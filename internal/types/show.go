package types

type Show struct {
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
	UploadDate   string `json:"upload_date,omitempty"`   // upload date of the series.
	LastModified string `json:"last_modified,omitempty"` // last modified date of the series.
}
