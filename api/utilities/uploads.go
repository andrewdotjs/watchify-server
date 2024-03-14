package utilities

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/andrewdotjs/watchify-server/api/types"
	"github.com/google/uuid"
)

// This takes a video file and stores it in the file system and database.
//
// This function does not return anything.
func HandleVideoUpload(uploadedFile *multipart.FileHeader, video *types.Video, database *sql.DB, uploadDirectory *string) {
	var uploadPath string
	var err error
	var splitFileName []string

	if _, err = os.ReadDir(*uploadDirectory); os.IsNotExist(err) {
		if err = os.Mkdir(*uploadDirectory, os.FileMode(777)); err != nil {
			log.Fatalf("ERR : %v", err)
		}
	}

	splitFileName = strings.Split(uploadedFile.Filename, ".")
	video.EpisodeNumber, err = strconv.Atoi(splitFileName[0])
	if err != nil {
		video.EpisodeNumber = 0
	}

	currentTime := time.Now().Format("01-02-2006 15:04:05")

	video.Id = uuid.New().String()
	video.FileExtension = splitFileName[1]
	video.FileName = fmt.Sprintf("%v.%v", video.Id, video.FileExtension)
	video.UploadDate = currentTime
	video.LastModified = currentTime

	if _, err = database.Exec(`
	  INSERT INTO videos
		VALUES (?, ?, ?, ?, ?, ?, ?, ?);
		`,
		video.Id,
		video.SeriesId,
		video.EpisodeNumber,
		video.Title,
		video.FileName,
		video.FileExtension,
		video.UploadDate,
		video.LastModified,
	); err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	uploadPath = path.Join(*uploadDirectory, video.FileName)
	out, err := os.Create(uploadPath)
	if err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	defer out.Close()

	file, err := uploadedFile.Open()
	if err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	// Copy the file to the destination and error handle.
	if _, err := io.Copy(out, file); err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}
}

// This takes a cover file and stores it in the file system and database.
//
// This function does not return anything.
func HandleCoverUpload(uploadedFile *multipart.FileHeader, cover *types.Cover, database *sql.DB, uploadDirectory *string) {
	if _, err := os.Stat(*uploadDirectory); err != nil { // Create the upload directory if it doesn't exist
		if !os.IsNotExist(err) {
			defer database.Close()
			log.Fatalf("ERR : %v", err)
		}

		log.Printf("SYS : Could not find upload directory at '%s'. Creating one.", *uploadDirectory)
		os.Mkdir(*uploadDirectory, os.ModePerm)
	}

	id := uuid.New().String()

	cover.Id = id
	cover.FileName = fmt.Sprintf("%s.%s", id, strings.Split(uploadedFile.Filename, ".")[1])
	cover.UploadDate = time.Now().Format("2006-01-02 15:04:05")

	if _, err := database.Exec(`
	  INSERT INTO covers
		VALUES (?, ?, ?, ?)
		`,
		cover.Id,
		cover.SeriesId,
		cover.FileName,
		cover.UploadDate,
	); err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	uploadPath := filepath.Join(*uploadDirectory, cover.FileName)
	out, err := os.Create(uploadPath)
	if err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	defer out.Close()

	file, err := uploadedFile.Open()
	if err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}

	// Copy the file to the destination and error handle.
	if _, err := io.Copy(out, file); err != nil {
		defer database.Close()
		log.Fatalf("ERR : %v", err)
	}
}
