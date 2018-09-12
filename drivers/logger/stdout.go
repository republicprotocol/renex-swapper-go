package logger

import (
	"fmt"
	"log"

	"github.com/republicprotocol/renex-swapper-go/adapter/logger"
	"github.com/republicprotocol/renex-swapper-go/domains/order"
)

const white = "\033[m"

type stdout struct {
}

// NewStdout creates a new stdout logger, which prints logs to the standard output.
func NewStdout() logger.Logger {
	return &stdout{}
}

func (logger *stdout) LogInfo(orderID [32]byte, msg string) {
	clr := pickColor(orderID)
	log.Println(fmt.Sprintf("[INF] (%s%s%s) %s", clr, order.Fmt(orderID), white, msg))
}

func (logger *stdout) LogDebug(orderID [32]byte, msg string) {
	clr := pickColor(orderID)
	log.Println(fmt.Sprintf("[DEB] (%s%s%s) %s", clr, order.Fmt(orderID), white, msg))
}

func (logger *stdout) LogError(orderID [32]byte, msg string) {
	clr := pickColor(orderID)
	log.Println(fmt.Sprintf("[ERR] (%s%s%s) %s", clr, order.Fmt(orderID), white, msg))
}

func pickColor(orderID [32]byte) string {
	return fmt.Sprintf("\033[3%dm", int64(orderID[0])%7)
}
