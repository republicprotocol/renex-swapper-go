package logger

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/republicprotocol/renex-swapper-go/service/logger"
)

const white = "\033[m"

type stdOut struct {
	UID [32]byte
}

func NewStdOut() logger.LoggerBuilder {
	return &stdOut{}
}

type loggerInstance struct {
	UID [32]byte
}

func (logger *stdOut) New(UID [32]byte) logger.Logger {
	return &loggerInstance{
		UID: UID,
	}
}

func (logger *loggerInstance) LogInfo(msg string) {
	clr := pickColor(logger.UID)
	log.Println(fmt.Sprintf("[INF] (%s%s%s) %s", clr, base64.StdEncoding.EncodeToString(logger.UID[:]), white, msg))
}

func (logger *loggerInstance) LogDebug(msg string) {
	clr := pickColor(logger.UID)
	log.Println(fmt.Sprintf("[DEB] (%s%s%s) %s", clr, base64.StdEncoding.EncodeToString(logger.UID[:]), white, msg))
}

func (logger *loggerInstance) LogError(msg string) {
	clr := pickColor(logger.UID)
	log.Println(fmt.Sprintf("[ERR] (%s%s%s) %s", clr, base64.StdEncoding.EncodeToString(logger.UID[:]), white, msg))
}

func pickColor(UID [32]byte) string {
	return fmt.Sprintf("\033[3%dm", int64(UID[0])%7)
}
