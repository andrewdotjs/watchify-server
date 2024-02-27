package types

type Cover struct {
	// Essential cover data
	Id       string `json:"id"`
	SeriesId string `json:"series_id"`

	// General data that is useful for debugging.
	FileName      string `json:"file_name,omitempty"`
	FileExtension string `json:"file_extension,omitempty"`
	UploadDate    string `json:"upload_date"`
}
