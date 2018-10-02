package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/republicprotocol/renex-swapper-go/driver/config"
	keystoreDriver "github.com/republicprotocol/renex-swapper-go/driver/keystore"
	"github.com/republicprotocol/renex-swapper-go/driver/swapper"
	"github.com/urfave/cli"
)

// Define flags for commands
var (
	networkFlag = cli.StringFlag{
		Name:  "network",
		Value: "mainnet",
		Usage: "name of the test network",
	}
	keyPhraseFlag = cli.StringFlag{
		Name:  "keyphrase",
		Value: "",
		Usage: "keyphrase to unlock the keystore file",
	}
	toFlag = cli.StringFlag{
		Name:  "to",
		Value: "",
		Usage: "receiver address for withdraw",
	}
	valueFlag = cli.Float64Flag{
		Name:  "value",
		Value: 0,
		Usage: "amount of token you want to withdraw,", // todo: specify the unit here
	}
	feeFlag = cli.Float64Flag{
		Name:  "fee",
		Value: 0,
		Usage: "amount of fee you want to pay for the withdraw transaction,", // todo: specify the unit here
	}
)

func main() {
	app := cli.NewApp()

	// Define sub-commands
	app.Commands = []cli.Command{
		{
			Name:  "http",
			Usage: "start running the swapper ",
			Flags: []cli.Flag{networkFlag, keyPhraseFlag},
			Action: func(c *cli.Context) error {
				// swapper, err  := initializeSwapper(c)
				// if err != nil {
				// 	return err
				// }
				panic("Implement the http logic here")
			},
		},
		{
			Name:  "withdraw",
			Usage: "withdraw the funds in the swapper accounts",
			Action: func(c *cli.Context) error {
				swapper, err := initializeSwapper(c)
				if err != nil {
					return err
				}

				return withdraw(c, swapper)
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func initializeSwapper(ctx *cli.Context) (swapper.Swapper, error) {
	network := ctx.String("network")
	keyPhrase := ctx.String("keyphrase")

	cfg, err := config.New(path.Join(os.Getenv("HOME"), fmt.Sprintf(".swapper/%v-config.json", network)), network)
	if err != nil {
		return nil, err
	}

	ks, err := keystoreDriver.LoadFromFile(cfg, keyPhrase)
	if err != nil {
		return nil, err
	}

	return swapper.NewSwapper(cfg, ks), nil
}

func withdraw(ctx *cli.Context, swapper swapper.Swapper) error {
	receiver := ctx.String("to")
	if receiver == "" {
		return errors.New("receiver address cannot be empty")
	}
	value := ctx.Float64("value")
	if value == 0 {
		return errors.New("please enter a valid withdraw amount ")
	}
	token := ctx.String("token")
	if token == "" {
		return errors.New("please enter a valid withdraw token")
	}
	fee := ctx.Float64("fee")
	if fee == 0 {
		return errors.New("please enter a valid withdraw fee")
	}

	swapper.Withdraw(token, receiver, value, fee)

	return nil
}
