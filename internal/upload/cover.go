package upload

import (
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"strings"
	"time"

	"github.com/andrewdotjs/watchify-server/internal/logger"
	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/andrewdotjs/watchify-server/internal/types"
	"github.com/google/uuid"
)

// This takes a movie cover file and stores it in the file system and database.
//
// This function returns an error response that needs to be sent to the client.
func Cover(
  uploadedFile *multipart.FileHeader,
  cover *types.Cover,
  database *sql.DB,
  uploadDirectory *string,
  log *logger.Logger,
  functionId *string,
) *responses.Error {
	var id string = uuid.NewString()
	var uploadPath string

	cover.Id = id
	cover.FileExtension = strings.Split(uploadedFile.Filename, ".")[1]
	cover.FileName = fmt.Sprintf("%s.%s", id, strings.Split(uploadedFile.Filename, ".")[1])
	cover.UploadDate = time.Now().Format("2006-01-02 15:04:05")

	log.Info(*functionId, "Starting cover upload")
	log.Info(*functionId, "Verifying upload directory")

	// Create the upload directory if it doesn't exist
	if _, err := os.Stat(*uploadDirectory); err != nil {
		if !os.IsNotExist(err) {
			var errorResponse responses.Error = responses.Error{
				Type:   "null",
				Title:  "Failure to upload movie cover",
				Status: 500,
				Detail: "An interesting error occurred while checking if upload directory exists.",
				// Add URL endpoint as instance.
			}

			log.Error(*functionId, fmt.Sprintf("An unknown error occurred during verification. %v", err))
			return &errorResponse
		}

		log.Info(*functionId, fmt.Sprintf("Could not find upload directory at %s. Creating one", *uploadDirectory))

		if err := os.Mkdir(*uploadDirectory, os.ModePerm); err != nil {
			var errorResponse responses.Error = responses.Error{
				Type:   "null",
				Title:  "Failure to upload movie cover",
				Status: 500,
				Detail: "Failed to create upload directory when needed.",
				// Add URL endpoint as instance.
			}

			log.Error(*functionId, fmt.Sprintf("Failed to create upload directory when needed. %v", err))
			return &errorResponse
		}
	}

	log.Info(*functionId, "Attempting to insert cover information into the database")

	// Insert the new cover's data into the series_covers table/
	if _, err := database.Exec(`
	  INSERT INTO
			covers
		VALUES
		  (?, ?, ?, ?, ?)
		`,
		cover.Id,
		cover.ParentId,
		cover.FileExtension,
		cover.FileName,
		cover.UploadDate,
	); err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload movie cover",
			Status: 500,
			Detail: "Failed to execute SQL insert statement.",
			// Add URL endpoint as instance.
		}

		log.Error(*functionId, fmt.Sprintf("Failed to upload cover information to the database. %v", err))
		return &errorResponse
	}

	log.Info(*functionId, "Creating file in file system")

	// Create file that will soon contain uploaded file contents.
	fmt.Println(cover.FileName)
	uploadPath = path.Join(*uploadDirectory, cover.FileName)
	out, err := os.Create(uploadPath)
	if err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload series cover",
			Status: 500,
			Detail: "Failed to create file in storage system.",
			// Add URL endpoint as instance.
		}

		log.Error(*functionId, fmt.Sprintf("Failed to create file in file system. %v", err))
		return &errorResponse
	}

	defer out.Close()

	log.Info(*functionId, "Attempting to open uploaded file")

	// Opens the file header to return the actual file.
	file, err := uploadedFile.Open()
	if err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload movie cover",
			Status: 500,
			Detail: "Failed to open uploaded file.",
			// Add URL endpoint as instance.
		}

		log.Error(*functionId, fmt.Sprintf("Failed to open uploaded file. %v", err))
		return &errorResponse
	}

	log.Info(*functionId, "Attempting to store file in file system.")

	// Copy the file contents to newly created file.
	if _, err := io.Copy(out, file); err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload cover",
			Status: 500,
			Detail: "Failed to copy file contents to newly created file.",
			// Add URL endpoint as instance.
		}

		log.Error(*functionId, fmt.Sprintf("Failed to store file in file system. %v", err))
		return &errorResponse
	}

	return nil
}
