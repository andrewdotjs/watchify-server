package upload

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path"
	"strings"
	"time"

	"github.com/andrewdotjs/watchify-server/internal/responses"
	"github.com/andrewdotjs/watchify-server/internal/types"
	"github.com/google/uuid"
)

// This takes a movie cover file and stores it in the file system and database.
//
// This function returns an error response that needs to be sent to the client.
func Cover(uploadedFile *multipart.FileHeader, movieCover *types.Cover, database *sql.DB, uploadDirectory *string) *responses.Error {
	var id string = uuid.NewString()
	var uploadPath string

	movieCover.Id = id
	movieCover.FileExtension = strings.Split(uploadedFile.Filename, ".")[1]
	movieCover.FileName = fmt.Sprintf("%s.%s", id, strings.Split(uploadedFile.Filename, ".")[1])
	movieCover.UploadDate = time.Now().Format("2006-01-02 15:04:05")

	log.Print("SYS : Starting movie cover upload sequence.")
	fmt.Print("SYS : Starting movie cover upload sequence.\n")

	log.Print("SYS : Checking if upload directory is present.")
	fmt.Print("SYS : Checking if upload directory is present.\n")

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

			log.Printf("SYS : An interesting error occurred while checking if upload directory exists. %v", err)
			fmt.Printf("SYS : An interesting error occurred while checking if upload directory exists. %v\n", err)

			return &errorResponse
		}

		log.Printf("SYS : Could not find upload directory at '%s'. Creating one.", *uploadDirectory)
		fmt.Printf("SYS : Could not find upload directory at '%s'. Creating one.\n", *uploadDirectory)

		if err := os.Mkdir(*uploadDirectory, os.ModePerm); err != nil {
			var errorResponse responses.Error = responses.Error{
				Type:   "null",
				Title:  "Failure to upload movie cover",
				Status: 500,
				Detail: "Failed to create upload directory when needed.",
				// Add URL endpoint as instance.
			}

			log.Print("SYS : Failed to create upload directory when needed.", err)
			fmt.Print("SYS : Failed to create upload directory when needed.\n", err)

			return &errorResponse
		}
	}

	log.Print("SYS : Executing insert statement into the database.")
	fmt.Print("SYS : Executing insert statement into the database.\n")

	// Insert the new cover's data into the series_covers table/
	if _, err := database.Exec(`
	  INSERT INTO
			movie_covers
		VALUES
		  (?, ?, ?, ?, ?, ?)
		`,
		movieCover.Id,
		movieCover.ParentId,
		movieCover.UserId,
		movieCover.FileExtension,
		movieCover.FileName,
		movieCover.UploadDate,
	); err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload movie cover",
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
	fmt.Println(movieCover.FileName)
	uploadPath = path.Join(*uploadDirectory, movieCover.FileName)
	out, err := os.Create(uploadPath)
	if err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload series cover",
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

		log.Printf("SYS : Failed to open uploaded file. %v", err)
		fmt.Printf("SYS : Failed to open uploaded file. %v\n", err)

		return &errorResponse
	}

	log.Print("SYS : Copying file contents to newly created file in storage system.")
	fmt.Print("SYS : Copying file contents to newly created file in storage system.\n")

	// Copy the file contents to newly created file.
	if _, err := io.Copy(out, file); err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload cover",
			Status: 500,
			Detail: "Failed to copy file contents to newly created file.",
			// Add URL endpoint as instance.
		}

		log.Printf("SYS : Failed to copy file contents to newly created file in storage system. %v", err)
		fmt.Printf("SYS : Failed to copy file contents to newly created file in storage system. %v\n", err)

		return &errorResponse
	}

	return nil
}
