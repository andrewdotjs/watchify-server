package logger

func (thisLogger *Logger) Info(transactionId string, message string) {
  thisLogger.Custom("INFO", &transactionId, &message)
}
