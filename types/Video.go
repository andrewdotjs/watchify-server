package types

type Video struct {
	Id       string `json:"id"`
	SeriesId string `json:"series_id",omitempty`
	Episode  int    `json:"episode_number",omitempty`
	Title    string `json:"title",omitempty`
}
