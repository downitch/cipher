// This package is developed in order to create a 
// useful API for Tor communication. It helps ship
// Tor inside the project and use it inside.
// It also allows to use Ethereum Blockchain as 
// communication protocol to exchange information
package api

import(
	"fmt"
	"context"
	"errors"
	"math/big"
	"math/rand"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type RandomBlock struct {
	hash   string
	number int
}

// This function returns client instance that allows using infura 
// gateway to communicate with the blockchain of Ethereum as if
// geth was running locally with light sync
func runGeth() (ethclient.Client, error) {
	// dialing to main network of infura.io
	httpsClient, err := ethclient.Dial("https://mainnet.infura.io")
	if err != nil {
		// returning empty instance and error
  	return ethclient.Client{}, err
	}
	// returning client instance
	return *httpsClient, nil
}

// This functions requires ethclient entity and an address to
// retireve current balance of the account from actual Blockchain
func GetBalance(addr string) string {
	client, err := runGeth()
	if err != nil {
		return "0"
	}
	// Decoding hex string to address (EVM)
	account := common.HexToAddress(addr)
	// asking ethclient entity to return balance as result
	balance, err := client.BalanceAt(context.Background(), account, nil)
	// if any error happens returning empty string
	if err != nil {
		return "0"
	}
	// returning balance as a string
	b := fmt.Sprintf("%d", balance)
	return b
}

// This function returns current block header hash
func GetLatestBlock() (string, error) {
	client, _ := runGeth()
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
	  return "", errors.New("can't get latest block")
	}
	return header.Number.String(), nil
}

// This function receives block number, casts it to bigInt and returns hash
func GetBlockHash(number int64) (string, error) {
	client, _ := runGeth()
	blockNumber := big.NewInt(number)
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		return "", errors.New("can't parse block data")
	}
	return strings.Split(block.Hash().Hex(), "x")[1], nil
}

// This function let's caller get random block from an updated blockchain
func GetRandomBlock() (RandomBlock, error) {
	latest, _ := GetLatestBlock()
	latestInt, _ := strconv.Atoi(latest)
	latestInt = latestInt - 1
	randomInt := rand.Intn(latestInt)
	randomInt = randomInt + 1
	data, _ := GetBlockHash(int64(randomInt))
	return RandomBlock{ data, randomInt }, nil
}

// This function writes down 50 random blocks one-by-one
// in compare to all other functions this one is built wrong
// since it directly communicates with database, which is wrong
func (c *Commander) GetManyRandomBlocks() {
	fmt.Println("generating new blocks...")
	limit := 50
	step := 1
	latest, _ := GetLatestBlock()
	latestInt, _ := strconv.Atoi(latest)
	latestInt = latestInt - 1
	for {
		if step > limit {
			break
		}
		randomInt := rand.Intn(latestInt)
		randomInt = randomInt + 1
		data, _ := GetBlockHash(int64(randomInt))
		err := c.SaveBlock(data, randomInt)
		if err != nil {
			break
		}
		step = step + 1
	}
	return
}

// This function will be later only available locally as it
// formats, ciphers and sends the message from current wallet
// to the recepients wallet address over ETH Blockchain
func FormRawTxWithBlockchain(msg []byte, recepient string) (string, error) {
	// parsing private key's ECDSA from hex string
	key := GenRandomString(32)
	hexedKey := Hexify(key)
	privateKey, err := crypto.HexToECDSA(hexedKey)
	if err != nil {
		return "can't parse key", err
	}
	value := big.NewInt(int64(0))
	nonce := uint64(0)
	gasLimit := uint64(0)
	gasPrice := big.NewInt(int64(0))
	to := common.HexToAddress(recepient)
	tx := types.NewTransaction(nonce, to, value, gasLimit, gasPrice, msg)
 	// signing the transaction before sending it
 	// it is required due to not using MetaMask
 	CID := big.NewInt(int64(1))
 	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(CID), privateKey)
 	if err != nil {
 		return "can't sign transaction", err
 	}
 	ts := types.Transactions{signedTx}
 	rawTxBytes := ts.GetRlp(0)
 	rawTxHex := Hexify(rawTxBytes)
 	// raw transaction is return
 	result := fmt.Sprintf("0x%s", rawTxHex)
 	return result, nil
}

func FormRawAccoutBlockchain() ([]byte, error) {
	// parsing private key's ECDSA from hex string
	key := GenRandomString(32)
	hexedKey := Hexify(key)
	privateKey, err := crypto.HexToECDSA(hexedKey)
	if err != nil {
		return []byte{}, err
	}
	b := crypto.FromECDSA(privateKey)
	return b, nil
}

// This function parses transaction and returns data field as
// a result. It is neccessary, because messages are sent via tx
func DecodeRawTx(rawTx string) ([]byte, error) {
	var tx *types.Transaction
	raw, err := Dehexify(rawTx)
	if err != nil {
		return []byte("can't parse raw tx"), err
	}
	rlp.DecodeBytes(raw, &tx)
	return tx.Data(), nil
}