package types

type Cover struct {
	Id         string `json:"id"`
	SeriesId   string `json:"series_id"`
	FileName   string `json:"file_name"`
	UploadDate string `json:"upload_date"`
}
