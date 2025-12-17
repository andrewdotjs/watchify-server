package logger

func (thisLogger *Logger) Fatal(location string, message string) {
  thisLogger.Custom("FATAL", location, message)
}
