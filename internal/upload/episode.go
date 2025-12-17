package upload

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/andrewdotjs/watchify-server/internal/types"
	"github.com/google/uuid"
)

// This takes a episode file and stores it in the file system and database.
//
// This function does not return anything.
func Episode(uploadedFile *multipart.FileHeader, episode *types.Episode, database *sql.DB, uploadDirectory *string) *responses.Error {
	var currentDateTime string = time.Now().Format("2006-01-02 15:04:05")
	var splitFileName []string = strings.Split(uploadedFile.Filename, ".")
	var uploadPath string
	var id string = uuid.NewString()
	var err error

	episode.Id = id
	episode.FileExtension = splitFileName[1]
	episode.FileName = fmt.Sprintf("%s, %s", id, splitFileName[1])
	episode.UploadDate = currentDateTime
	episode.LastModified = currentDateTime

	log.Print("SYS : Starting series episode upload sequence.")
	fmt.Print("SYS : Starting series episode upload sequence.\n")

	log.Print("SYS : Checking if upload directory is present.")
	fmt.Print("SYS : Checking if upload directory is present.\n")

	// Create the upload directory if it doesn't exist.
	if _, err = os.ReadDir(*uploadDirectory); os.IsNotExist(err) {
		log.Printf("SYS : Could not find upload directory at '%s'. Creating one.", *uploadDirectory)
		fmt.Printf("SYS : Could not find upload directory at '%s'. Creating one.\n", *uploadDirectory)

		if err = os.Mkdir(*uploadDirectory, os.FileMode(0777)); err != nil {
			var errorResponse responses.Error = responses.Error{
				Type:   "null",
				Title:  "Failure to upload series episode",
				Status: 500,
				Detail: "Failed to create a new upload directory when needed.",
				// Add URL endpoint as instance.
			}

			log.Printf("SYS : Failed to create a new upload directory when needed. %v", err)
			fmt.Printf("SYS : Failed to create a new upload directory when needed. %v\n", err)

			return &errorResponse
		}
	}

	// Retrieve episode number from file name.
	if number, err := strconv.Atoi(splitFileName[0]); err != nil {
		log.Printf("SYS : Error retrieving episode number from filename \"%s\", can this be converted to an integer? Setting this value to 0.", splitFileName[0])
		fmt.Printf("SYS : Error retrieving episode number from filename \"%s\", can this be converted to an integer? Setting this value to 0.\n", splitFileName[0])
		episode.EpisodeNumber = 0
	} else {
		episode.EpisodeNumber = number
	}

	log.Print("SYS : Executing insert statement into the database.")
	fmt.Print("SYS : Executing insert statement into the database.\n")

	// Insert the new episode's data into the series_episodes table.
	if _, err = database.Exec(`
	  INSERT INTO
			series_episodes
		VALUES
		  (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
		episode.Id,
		episode.ParentId,
		episode.EpisodeNumber,
		nil,
		nil,
		episode.FileName,
		episode.FileExtension,
		episode.UploadDate,
		episode.LastModified,
	); err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload series episode",
			Status: 500,
			Detail: "Failed to execute SQL insert statement.",
			// Add URL endpoint as instance.
		}

		log.Printf("SYS : Failed to execute SQL insert statement. %v", err)
		fmt.Printf("SYS : Failed to execute SQL insert statement. %v\n", err)

		return &errorResponse
	}

	log.Print("SYS : Creating file in storage system.")
	fmt.Print("SYS : Creating file in storage system.\n")

	// Create file that will soon contain uploaded file contents.
	uploadPath = path.Join(*uploadDirectory, episode.FileName)
	out, err := os.Create(uploadPath)
	if err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload series episode",
			Status: 500,
			Detail: "Failed to create file in storage system.",
			// Add URL endpoint as instance.
		}

		log.Printf("SYS : Failed to create file in storage system. %v", err)
		fmt.Printf("SYS : Failed to create file in storage system. %v\n", err)

		return &errorResponse
	}

	defer out.Close()

	log.Print("SYS : Opening uploaded file.")
	fmt.Print("SYS : Opening uploaded file.\n")

	// Open uploaded file so that it's ready to be copied.
	file, err := uploadedFile.Open()
	if err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload series episode",
			Status: 500,
			Detail: "Failed to open uploaded file.",
			// Add URL endpoint as instance.
		}

		log.Printf("SYS : Failed to open uploaded file. %v", err)
		fmt.Printf("SYS : Failed to open uploaded file. %v\n", err)

		return &errorResponse
	}

	log.Print("SYS : Copying file contents to newly created file in storage system.")
	fmt.Print("SYS : Copying file contents to newly created file in storage system.\n")

	// Copy the file to the destination and error handle.
	if _, err := io.Copy(out, file); err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload series episode",
			Status: 500,
			Detail: "Failed to copy file contents to newly created file.",
			// Add URL endpoint as instance.
		}

		log.Printf("SYS : Failed to copy file contents to newly created file. %v", err)
		fmt.Printf("SYS : Failed to copy file contents to newly created file. %v\n", err)

		return &errorResponse
	}

	return nil
}
