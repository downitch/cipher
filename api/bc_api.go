// This package is developed in order to create a 
// useful API for Tor communication. It helps ship
// Tor inside the project and use it inside.
// It also allows to use Ethereum Blockchain as 
// communication protocol to exchange information
package api

import(
	"fmt"
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/core/types"
)

type Client ethclient.Client

type Account struct {
	key string
}

// This function returns client instance that allows using infura 
// gateway to communicate with the blockchain of Ethereum as if
// geth was running locally with light sync
func RunGeth() (ethclient.Client, ethclient.Client, error) {
	// dialing to main network of infura.io
	httpsClient, err := ethclient.Dial("https://mainnet.infura.io")
	if err != nil {
		// returning empty instance and error
  	return ethclient.Client{}, ethclient.Client{}, err
	}
	wsClient, err := ethclient.Dial("wss://mainnet.infura.io/ws")
	if err != nil {
		// returning empty instance and error
  	return ethclient.Client{}, ethclient.Client{}, err
	}
	// returning client instance
	return *httpsClient, *wsClient, nil
}

func Authorize(pkey string) Account {
	return Account{pkey}
}

// This functions requires ethclient entity and an address to
// retireve current balance of the account from actual Blockchain
func GetBalance(client ethclient.Client, addr string) string {
	// Decoding hex string to address (EVM)
	account := common.HexToAddress(addr)
	// asking ethclient entity to return balance as result
	balance, err := client.BalanceAt(context.Background(), account, nil)
	// if any error happens returning empty string
	if err != nil {
		return ""
	}
	// returning balance as a string
	return fmt.Sprintf("%d", balance)
}


// This function returns current block header hash
func GetLatestBlock(client ethclient.Client) (string, error) {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
	  return "", errors.New("can't get latest block")
	}
	return header.Number.String(), nil
}

// This function receives block number, casts it to bigInt and returns hash
func GetBlockHash(client ethclient.Client, number int64) (string, error) {
	blockNumber := big.NewInt(number)
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		return "", errors.New("can't parse block data")
	}
	return block.Hash().Hex(), nil
}

// This function will be later only available locally as it
// formats, ciphers and sends the message from current wallet
// to the recepients wallet address over ETH Blockchain
func SendMessageByBlockchain(client ethclient.Client, key string, msg string, recepient string) (string, error) {
	// parsing private key's ECDSA from hex string
	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return "can't parse key", err
	}
	// obtaining public key from private
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "can't parse to ecdsa", errors.New("can't parse to ecdsa")
	}
	// making backwards process to set fromAddress
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "can't get current nonce", err
	}
	// calculating optimal gas price so the tx will be rejected
	value := big.NewInt(0)
	gasLimit := uint64(21000 + (68 * len(msg)))
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "can't suggest gas price", err
	}
	lowGasPrice := big.NewInt(int64(0))
	lowGasPrice.Div(gasPrice, big.NewInt(10000000))
	castedGasPrice := lowGasPrice.Int64()
	// setting estimated gas price by dividing it
	delimiter := 2.8
	if castedGasPrice < int64(100) {
		delimiter = 1.5
	} else if castedGasPrice < int64(200) {
		delimiter = 2.1
	}
	// for better precision gasprice is divided in float
	// and converted to int64 type to cast it into big.Int
	division := int64(float64(castedGasPrice) / float64(delimiter))
	gasPriceBackCasted := big.NewInt(division)
	gasPriceBackCasted.Mul(gasPriceBackCasted, big.NewInt(10000000))
	gasPriceNeeded := gasPriceBackCasted
	// setting toAddress and forming the transaction
	toAddress := common.HexToAddress(recepient)
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPriceNeeded, []byte(msg))
	// obtaining chainID (1 for main network)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
 		return "can't get chainID", err
 	}
 	// signing the transaction before sending it
 	// it is required due to not using MetaMask
 	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
 	if err != nil {
 		return "can't sign transaction", err
 	}
 	// sending transaction into blockchain
 	err = client.SendTransaction(context.Background(), signedTx)
 	if err != nil {
 		fmt.Println(err)
 		return "can't send transaction", err
 	}
 	// returning transaction receipt hex value to the user
 	// this value may be used to find transaction in the blockchain
 	return signedTx.Hash().Hex(), nil
}

// This function connects to infura websocket and starts watching
// blockchain blocks checking if there are any messages delivered
// to current account (public key == address)
func WatchBlockchain(client ethclient.Client, key string) error {
	// setting a channel that will receive blocks and pass it into loop
	headers := make(chan *types.Header)
	// subscribing to receive new headers
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		fmt.Println("error subscribing")
		return err
	}
	// starting a forever loop that will watch every block and check it
	for {
		select {
			// in case if an error happens, returning an error
			case err := <-sub.Err():
				fmt.Println("error getting block")
				return err
			// checking header, parsing block information
			case header := <-headers:
				block, _ := client.BlockByHash(context.Background(), header.Hash())
				// if block contains transactions loop them
				if block != nil {
					if len(block.Transactions()) != 0 {
						go func() {
							for _, tx := range block.Transactions() {
								// if recepient and current public address are same
								// passing message back in order to store it in db
								if tx.To() != nil {
									if tx.To().Hex() == key {
										fmt.Println(tx)
									}
								}
							}
						}()
					}
				}
		}
	}
}