package types

type MovieComment struct {
	Id      string `json:"id"`
	MovieId string `json:"movie_id"`
	UserId  string `json:"user_id"`

	Message  string `json:"message"`
	Likes    int    `json:"likes"`
	Dislikes int    `json:"dislikes"`

	// General data that is useful for debugging.
	CreationDate string `json:"creation_date,omitempty"` // upload date of the movie.
	LastModified string `json:"last_modifed,omitempty"`  // last modified date of the movie.
}

type MovieCover struct {
	// Primary and Foreign keys
	Id      string `json:"id"` // (Primary)
	MovieId string `json:"movie_id"`
	UserId  string `json:"user_id"`

	// General data
	FileExtension string `json:"file_extension"`
	FileName      string `json:"file_name"`
	UploadDate    string `json:"upload_date"`
}

type Movie struct {
	Id string `json:"id"` // uuid of the movie

	Title       string `json:"title,omitempty"`       // title of the movie
	Description string `json:"description,omitempty"` // description of the movie
	Hidden      bool   `json:"hidden"`

	Cover map[string]any `json:"cover,omitempty"`
	// 	EXAMPLE:
	//  "cover": {
	//		"exists": true,
	//		"url": "example.com/api/v1/{series_id}/cover"
	//  }

	// General data.
	FileExtension string `json:"file_extension"`
	FileName      string `json:"file_name"`
	UploadDate    string `json:"upload_date,omitempty"`  // upload date of the movie.
	LastModified  string `json:"last_modifed,omitempty"` // last modified date of the movie.
}
