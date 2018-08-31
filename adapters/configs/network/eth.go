package network

// EthereumNetwork are the parameters required to create an ethereum client
type EthereumNetwork struct {
	Network            string `json:"network"`
	URL                string `json:"url"`
	RenExAtomicSwapper string `json:"renExAtomicSwapper"`
	RenExAtomicInfo    string `json:"renExAtomicInfo"`
	RenExSettlement    string `json:"renExSettlement"`
	Orderbook          string `json:"orderbook"`
}

var EthereumKovan = EthereumNetwork{
	Network: "kovan",
	URL:     "https://kovan.infura.io",
}

var EthereumRopsten = EthereumNetwork{
	Network: "ropsten",
	URL:     "https://ropsten.infura.io",
}

var EthereumMainnet = EthereumNetwork{
	Network: "mainnet",
	URL:     "https://mainnet.infura.io",
}

func NewEthNetwork(net RepublicNetwork) EthereumNetwork {
	switch net {
	case RepublicNetworkNightly:
		nightly := EthereumKovan
		nightly.RenExAtomicInfo = "0xe1A660657A32053fe83B19B1177F6B56C6F37b1f"
		nightly.RenExAtomicSwapper = "0x888D1e20E2e94D4d66aA6E80580012C65Fc69a78"
		nightly.RenExSettlement = "0x5f25233ca99104D31612D4fB937B090d5A2EbB75"
		nightly.Orderbook = "0xa95dE870dDFB6188519D5CC63CEd5E0FBac1aa8E"
		return nightly
	case RepublicNetworkFalcon:
		falcon := EthereumKovan
		falcon.RenExAtomicInfo = ""
		falcon.RenExAtomicSwapper = ""
		falcon.RenExSettlement = ""
		falcon.Orderbook = ""
		return falcon
	default:
		testnet := EthereumKovan
		testnet.RenExAtomicInfo = ""
		testnet.RenExAtomicSwapper = ""
		testnet.RenExSettlement = ""
		testnet.Orderbook = ""
		return testnet
	}
}

func (network *Config) GetEthereumNetwork() EthereumNetwork {
	network.mu.RLock()
	defer network.mu.RUnlock()
	return network.Ethereum
}

func (config *Config) SetEthereumNetwork(ethereumConfig EthereumNetwork) {
	config.mu.Lock()
	defer config.mu.Unlock()
	config.Ethereum = ethereumConfig
	config.Update()
}
