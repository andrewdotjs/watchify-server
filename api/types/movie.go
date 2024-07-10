package types

type Movie struct {
 	Id           string `json:"id"`                    // uuid of the movie
  CoverId      string `json:"cover_id,omitempty"`
	Title        string `json:"title,omitempty"`       // title of the movie
	Description  string `json:"description,omitempty"` // description of the movie

	Cover map[string]any `json:"cover,omitempty"`
	// 	EXAMPLE:
	//  "cover": {
	//		"exists": true,
	//		"url": "example.com/api/v1/{series_id}/cover"
	//  }

	// General data that is useful for debugging.
	UploadDate   string `json:"upload_date,omitempty"`  // upload date of the movie.
	LastModified string `json:"last_modifed,omitempty"` // last modified date of the movie.
}
