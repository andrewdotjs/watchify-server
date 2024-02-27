package responses

import "net/http"

type File struct {
	StatusCode int
	FileBuffer []byte
}

func (file File) ToClient(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(200)
	w.Write(file.FileBuffer)
}
