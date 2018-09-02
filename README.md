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

### How to buy Bitcoin

1. Turn on the swapper by running the following command.

`swapper`

2. Make sure that you donot close this terminal before the swap is completed. If you have to close the terminal for some reason please run the command again once you open it. *You have to keep the swapper running to do the atomic swap*

3. Open https://testnet.ren.exchange.

4. Select the Ethereum / Bitcoin trading pair.

5. Click "Connect to atomic swapper".

6. Authorize your atomic swapping software to execute your atomic swaps.

7. Open the "Balances" tab and ensure that you have the necessary funds in your swapper's Ethereum address under "Balances held by Atom", and at least 0.6% of the intended trading amount in your RenEx account.

8. Open the "Exchange" tab and open a Buy BTC-ETH order for the intended amount.


### How to buy Ether

1. Turn on the swapper by running the following command

```swapper```

2. Make sure that you donot close this terminal before the swap is completed. If you have to close the terminal for some reason please run the command again once you open it. *You have to keep the swapper running to do the atomic swap*

3. Open https://testnet.ren.exchange.

4. Select the Ethereum / Bitcoin trading pair.

5. Click "Connect to atomic swapper".

6. Authorize your atomic swapping software to execute your atomic swaps.

7. Open the "Balances" tab and ensure that you have the necessary funds in your swapper's Bitcoin address under "Balances held by Atom", some amount in your ethereum address (to pay for the transaction fees), and at least 0.6% of the intended trading amount in your RenEx account.

8. Open the "Exchange" tab and open a Sell BTC-ETH order for the intended amount.

#### Note:
The Bitcoin fees are set to high (500000 Satoshi) for faster trades in the testnet, it will be reverted to standard (10000 Satoshi) in the mainnet. 
