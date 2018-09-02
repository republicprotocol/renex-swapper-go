# RenEx Atomic Swapper

The RenEx Atomic Swapper is built and officially supported by, the Republic Protocol team. It can be used to execute atomic swaps between Ethereum and Bitcoin, and while it can be used independently of RenEx, it is designed for use with https://testnet.ren.exchange. Using this software, traders will be able to open Ethereum to Bitcoin orders on https://testnet.ren.exchange.
    
## Installation

### Mac/Ubuntu

#### Prerequisites

1. Unzip
2. Curl

#### Steps

1. Run the following command
`curl https://releases.republicprotocol.com/swapper/install.sh -sSf | sh`

2. When prompted, enter the Ethereum address that you will use with https://testnet.ren.exchange. This Ethereum address must hold all trading fees but does not hold the funds used for swapping. The swapper uses this address to distinguish between trades opened by RenEx vs. other malicious websites.

### Windows

> Coming soon!

### Funds
IMPORTANT: The swapper's ethereum address needs some ether to pay for the transaction costs on Ethereum blockchain. So, for example, you will need some amount of Ether even if you are selling Bitcoin for Ether in your swapper's Ethereum address.

#### Fees
The fees are always paid in Ether and are deducted from your RenEx's balance, so make sure that you have enough funds in it.

#### Trading Funds
The funds you want to swap should be deposited into the swapper.

## Usage

The RenEx Atomic Swapper is designed for use with https://testnet.ren.exchange. 

IMPORTANT: The RenEx Atomic Swapper must be running at all times. If it is not running, it will not be able to execute atomic swaps. If you fail to execute an atomic swap for matching orders, your trading account being fined, resulting in the loss of funds.

To open an atomic swap on RenEx:

1. Select the Ethereum / Bitcoin trading pair.

2. Click "Connect to atomic swapper".

3. Authorize your atomic swapping software to execute your atomic swaps.

4. Open the "Balances" tab and ensure that you have the necessary funds in the Ethereum and Bitcoin addresses.

5. Open the "Exchange" tab and open an order.
