package types

import "reflect"

type Cover struct {
	Id         string `json:"id"`
	SeriesId   string `json:"series_id"`
	FileName   string `json:"file_name"`
	UploadDate string `json:"upload_date"`
}

func (c Cover) ToColumns() []interface{} {
	s := reflect.ValueOf(&c).Elem()
	numCols := s.NumField()
	columns := make([]interface{}, numCols)
	for i := 0; i < numCols; i++ {
		field := s.Field(i)
		columns[i] = field.Addr().Interface()
	}

	return columns
}
