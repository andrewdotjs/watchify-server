package types

type Cover struct {
	// Primary and Foreign keys
	Id      string `json:"id"` // (Primary)
	ParentId string `json:"parent_id"`
	UserId  string `json:"user_id"`

	// General data
	FileExtension string `json:"file_extension"`
	FileName      string `json:"file_name"`
	UploadDate    string `json:"upload_date"`
}
