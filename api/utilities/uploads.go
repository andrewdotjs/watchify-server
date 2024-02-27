package utilities

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

	"github.com/andrewdotjs/watchify-server/api/types"
	"github.com/google/uuid"
)

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
	video.Episode, err = strconv.Atoi(splitFileName[0])
	if err != nil {
		video.Episode = 0
	}

	video.Id = uuid.New().String()
	video.FileExtension = splitFileName[1]
	video.FileName = fmt.Sprintf("%v.%v", video.Id, video.FileExtension)

	_, err = database.Exec(`
	  INSERT INTO videos
		VALUES (
			?,
			?,
			?,
			?,
			?,
			datetime("now", "localtime"),
			datetime("now", "localtime"),
		);
	`, video.Id, video.SeriesId, video.Episode, video.Title, video.FileName, video.FileExtension) // TODO: Find a way to make this shorter w/o lame formatting

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
