package types

type User struct {
	Id string `json:"id"`

	Username  string `json:"username"`
	Password  string `json:"password"`   // This needs to be an encrypted hash.
	FirstName string `json:"first_name"` // Optional: allows the app to feel a bit more personal.

	CreationDate string `json:"creation_date"` // Needs to be a golang date converted to a string.
	LastModified string `json:"last_modified"` // Needs to be a golang date converted to a string.
}

// This will be the struct for data containing information on user profile images.
type Image struct {
	Id     string `json:"id"`
	UserId string `json:"user_id"` // Foreign key

	UploadDate string `json:"upload_date"` // Needs to be a golang date converted to a string.
}
