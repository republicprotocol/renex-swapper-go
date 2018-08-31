package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/republicprotocol/renex-swapper-go/adapters/configs/general"
)

func main() {
	cfg := config.NewConfig()
	addresses := []string{}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Authorize your ethereum address(es): ")
	for {
		text, _ := reader.ReadString('\n')
		if text == "\n" {
			break
		}
		addresses = append(addresses, strings.Trim(text, "\n"))
	}
	cfg.AuthorizedAddresses = addresses
	if err := cfg.Update(); err != nil {
		panic(err)
	}

}
