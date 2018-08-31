package network

var BitcoinTestnet = BitcoinNetwork{
	Network: "testnet",
	URL:     "https://testnet.blockchain.info",
}

var BitcoinMainnet = BitcoinNetwork{
	Network: "mainnet",
	URL:     "https://mainnet.blockchain.info",
}

type BitcoinNetwork struct {
	Network string `json:"network"`
	URL     string `json:"url"`
}

func NewBtcNetwork(net RepublicNetwork) BitcoinNetwork {
	switch net {
	case RepublicNetworkNightly:
		return BitcoinTestnet
	case RepublicNetworkFalcon:
		return BitcoinTestnet
	case RepublicNetworkTestnet:
		return BitcoinTestnet
	default:
		return BitcoinTestnet
	}
}

func (network *Config) GetBitcoinNetwork() BitcoinNetwork {
	network.mu.RLock()
	defer network.mu.RUnlock()
	return network.Bitcoin
}

func (network *Config) SetBitcoinNetwork(bitcoinNetwork BitcoinNetwork) {
	network.mu.Lock()
	defer network.mu.Unlock()
	network.Bitcoin = bitcoinNetwork
	network.Update()
}
