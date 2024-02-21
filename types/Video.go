package types

import "reflect"

type Video struct {
	Id         string `json:"id,omitempty"`
	SeriesId   string `json:"series_id,omitempty"`
	Episode    int    `json:"episode_number,omitempty"`
	Title      string `json:"title,omitempty"`
	FileName   string `json:"file_name"`
	UploadDate string `json:"upload_date"`
}

func (v Video) ToColumns() []interface{} {
	s := reflect.ValueOf(&v).Elem()
	numCols := s.NumField()
	columns := make([]interface{}, numCols)
	for i := 0; i < numCols; i++ {
		field := s.Field(i)
		columns[i] = field.Addr().Interface()
	}

	return columns
}
