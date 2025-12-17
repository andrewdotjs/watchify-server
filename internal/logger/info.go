package logger

func (thisLogger *Logger) Info(location string, message string) {
  thisLogger.Custom("INFO", location, message)
}
