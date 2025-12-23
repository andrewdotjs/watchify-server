package logger

func (thisLogger *Logger) Fatal(transactionId string, message string) {
  thisLogger.Custom("FATAL", &transactionId, &message)
}
