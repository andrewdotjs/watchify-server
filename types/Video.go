package types

type Video struct {
	Id         string `json:"id,omitempty"`
	SeriesId   string `json:"series_id,omitempty"`
	Episode    int    `json:"episode_number,omitempty"`
	Title      string `json:"title,omitempty"`
	FileName   string `json:"file_name,omitempty"`
	UploadDate string `json:"upload_date,omitempty"`
}
