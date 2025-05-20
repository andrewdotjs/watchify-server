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

func Movie(uploadedFile *multipart.FileHeader, movieStruct *types.Movie, database *sql.DB, uploadDirectory *string) (*string, *responses.Error) {
	var currentTime string = time.Now().Format("01-02-2006 15:04:05")
	var uploadPath string
	var err error

	log.Print("SYS : Starting movie upload sequence.")
	fmt.Print("SYS : Starting movie upload sequence.\n")

	movieStruct.Id = uuid.New().String()
	movieStruct.FileExtension = strings.Split(uploadedFile.Filename, ".")[1]
	movieStruct.FileName = fmt.Sprintf("%v.%v", movieStruct.Id, movieStruct.FileExtension)
	movieStruct.UploadDate = currentTime
	movieStruct.LastModified = currentTime

	log.Print("SYS : Checking if upload directory is present.")
	fmt.Print("SYS : Checking if upload directory is present.\n")

	// Create the upload directory if it doesn't exist.
	if _, err = os.ReadDir(*uploadDirectory); os.IsNotExist(err) {
		log.Printf("SYS : Could not find upload directory at '%s'. Creating one.", *uploadDirectory)
		fmt.Printf("SYS : Could not find upload directory at '%s'. Creating one.\n", *uploadDirectory)

		if err = os.Mkdir(*uploadDirectory, os.FileMode(0777)); err != nil {
			var errorResponse responses.Error = responses.Error{
				Type:   "null",
				Title:  "Failure to upload movie",
				Status: 500,
				Detail: "Failed to create upload directory when needed.",
				// Add URL endpoint as instance.
			}

			log.Print("SYS : Failed to create upload directory when needed.", err)
			fmt.Print("SYS : Failed to create upload directory when needed.\n", err)

			return nil, &errorResponse
		}
	}

	log.Print("SYS : Executing insert statement into the database.")
	fmt.Print("SYS : Executing insert statement into the database.\n")

	// Insert the new episode's data into the series_episodes table.
	if _, err = database.Exec("INSERT INTO movies VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		movieStruct.Id,
		movieStruct.Title,
		movieStruct.Description,
		movieStruct.Hidden,
		movieStruct.FileExtension,
		movieStruct.FileName,
		movieStruct.UploadDate,
		movieStruct.LastModified,
	); err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload movie",
			Status: 500,
			Detail: "Failed to execute SQL insert statement.",
			// Add URL endpoint as instance.
		}

		log.Printf("SYS : Failed to execute SQL insert statement. %v", err)
		fmt.Printf("SYS : Failed to execute SQL insert statement. %v\n", err)

		return nil, &errorResponse
	}

	log.Print("SYS : Creating file in storage system.")
	fmt.Print("SYS : Creating file in storage system.\n")

	// Create file that will soon contain uploaded file contents.
	uploadPath = path.Join(*uploadDirectory, movieStruct.FileName)
	out, err := os.Create(uploadPath)
	if err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload movie",
			Status: 500,
			Detail: "Failed to create file in storage system.",
			// Add URL endpoint as instance.
		}

		log.Printf("SYS : Failed to create file in storage system. %v", err)
		fmt.Printf("SYS : Failed to create file in storage system. %v\n", err)

		return nil, &errorResponse
	}

	defer out.Close()

	log.Print("SYS : Opening uploaded file.")
	fmt.Print("SYS : Opening uploaded file.\n")

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

		log.Printf("SYS : Failed to open uploaded file. %v", err)
		fmt.Printf("SYS : Failed to open uploaded file. %v\n", err)

		return nil, &errorResponse
	}

	log.Print("SYS : Copying file contents to newly created file in storage system.")
	fmt.Print("SYS : Copying file contents to newly created file in storage system.\n")

	// Copy the file to the destination and error handle.
	if _, err := io.Copy(out, file); err != nil {
		var errorResponse responses.Error = responses.Error{
			Type:   "null",
			Title:  "Failure to upload movie",
			Status: 500,
			Detail: "Failed to copy file contents to newly created file.",
			// Add URL endpoint as instance.
		}

		log.Printf("SYS : Failed to copy file contents to newly created file. %v", err)
		fmt.Printf("SYS : Failed to copy file contents to newly created file. %v\n", err)

		return nil, &errorResponse
	}

	return &movieStruct.Id, nil
}
