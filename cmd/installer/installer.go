package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	config "github.com/republicprotocol/renex-swapper-go/adapters/configs/general"
	"github.com/republicprotocol/renex-swapper-go/adapters/configs/keystore"
	"github.com/republicprotocol/renex-swapper-go/adapters/configs/network"
	"github.com/republicprotocol/renex-swapper-go/utils"
)

func main() {
	home := utils.GetHome()
	repNet := flag.String("republic", "nightly", "which republic protocol network to use")
	flag.Parse()

	keystore.NewKeystore([]uint32{0, 1}, []string{"testnet", "kovan"}, home+"/.swapper/keystore.json")

	var republicNet network.RepublicNetwork
	switch *repNet {
	case "nightly":
		republicNet = network.RepublicNetworkNightly
	case "falcon":
		republicNet = network.RepublicNetworkFalcon
	case "testnet":
		republicNet = network.RepublicNetworkTestnet
	default:
		panic("Unknown Republic Network")
	}

	cfg := config.NewConfig()
	net := network.NewNetwork(republicNet)
	addresses := []string{}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nEnter your RenEx Ethereum address: ")
	text, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	addresses = append(addresses, strings.Trim(text, "\r\n"))
	cfg.AuthorizedAddresses = addresses

	if err := cfg.Update(); err != nil {
		panic(err)
	}

	if err := net.Update(); err != nil {
		panic(err)
	}
}
