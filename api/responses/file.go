package responses

import "net/http"

type File struct {
	StatusCode int
	FileBuffer []byte
}

// Takes a built File struct and converts it into JSON-compatible bytes
// using the "encoding/json" library then sends to client through provided
// ResponseWriter.
func (file File) ToClient(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Language", "en")
	w.WriteHeader(file.StatusCode)
	w.Write(file.FileBuffer)
}
