package logger

type LoggerBuilder interface {
	New([32]byte) Logger
}
type Logger interface {
	LogInfo(msg string)
	LogDebug(msg string)
	LogError(msg string)
}
