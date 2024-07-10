package types

type Cover struct {
	// Essential cover data
	Id       string `json:"id"`

	// General data that is useful for debugging.
	FileExtension string `json:"file_extension,omitempty"`
	UploadDate    string `json:"upload_date"`
}
