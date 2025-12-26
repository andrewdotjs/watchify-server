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

func Movie(
  uploadedFile *multipart.FileHeader,
  movie *types.Movie,
  database *sql.DB,
  uploadDirectory *string,
  log *logger.Logger,
  functionId *string,
) (*string, *responses.Error) {
	var currentTime string = time.Now().Format("01-02-2006 15:04:05")
	var uploadPath string
	var err error

	log.Info(*functionId, "Commencing movie upload")

	movie.Id = uuid.New().String()
	movie.FileExtension = strings.Split(uploadedFile.Filename, ".")[1]
	movie.FileName = fmt.Sprintf("%v.%v", movie.Id, movie.FileExtension)
	movie.UploadDate = currentTime
	movie.LastModified = currentTime

	log.Info(*functionId, "Verifying upload directory")

	// Create the upload directory if it doesn't exist.
	if _, err = os.ReadDir(*uploadDirectory); os.IsNotExist(err) {
	  log.Error(*functionId, fmt.Sprintf("Could not find upload directory at %s. Creating one", *uploadDirectory))

		if err = os.Mkdir(*uploadDirectory, os.FileMode(0777)); err != nil {
			var errorResponse responses.Error = responses.Error{
				Type:   "null",
				Title:  "Failure to upload movie",
				Status: 500,
				Detail: "Failed to create upload directory when needed.",
				// Add URL endpoint as instance.
			}

			log.Error(*functionId, fmt.Sprintf("Failed to create upload directory when needed. %v", err))
			return nil, &errorResponse
		}
	}

	log.Info(*functionId, "Attempting to insert movie information into the database.")

	// Insert the new episode's data into the series_episodes table.
	if _, err = database.Exec(
	  `
			INSERT INTO
			  movies
			VALUES
			  (?, ?, ?, ?, ?, ?, ?, ?)
		`,
		movie.Id,
		movie.Title,
		movie.Description,
		movie.Hidden,
		movie.FileExtension,
		movie.FileName,
		movie.UploadDate,
		movie.LastModified,
	); err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload movie",
			Status: 500,
			Detail: "Failed to execute SQL insert statement.",
			// Add URL endpoint as instance.
		}

		log.Error(*functionId, fmt.Sprintf("Insert failed. %v", err))
		return nil, &errorResponse
	}

	log.Info(*functionId, "Creating file in file system")

	// Create file that will soon contain uploaded file contents.
	uploadPath = path.Join(*uploadDirectory, movie.FileName)
	out, err := os.Create(uploadPath)
	if err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload movie",
			Status: 500,
			Detail: "Failed to create file in storage system.",
			// Add URL endpoint as instance.
		}

		log.Error(*functionId, fmt.Sprintf("Failed to create file in file system. %v", err))
		return nil, &errorResponse
	}

	defer out.Close()

	log.Info(*functionId, "Opening uploaded file")

	// Open uploaded file so that it's ready to be copied.
	file, err := uploadedFile.Open()
	if err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload movie",
			Status: 500,
			Detail: "Failed to open uploaded file.",
			// Add URL endpoint as instance.
		}

		log.Error(*functionId, fmt.Sprintf("Failed to open uploaded file. %v", err))
		return nil, &errorResponse
	}

	log.Info(*functionId, "Attempting to store uploaded file into file system")

	// Copy the file to the destination and error handle.
	if _, err := io.Copy(out, file); err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload movie",
			Status: 500,
			Detail: "Failed to copy file contents to newly created file.",
			// Add URL endpoint as instance.
		}

		log.Error(*functionId, fmt.Sprintf("Failed to store uploaded file into file system. %v", err))
		return nil, &errorResponse
	}

	return &movie.Id, nil
}
