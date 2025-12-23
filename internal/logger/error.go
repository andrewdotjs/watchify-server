package logger

func (thisLogger *Logger) Error(transactionId string, message string) {
  thisLogger.Custom("ERROR", &transactionId, &message)
}
