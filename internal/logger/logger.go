package logger

import (
	"os"
)

type Logger struct {
  path string
  fileObject *os.File
}
