package logger

func (thisLogger *Logger) Error(location string, message string) {
  thisLogger.Custom("ERROR", location, message)
}
