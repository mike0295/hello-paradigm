package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"hello-paradigm/uniswap_v2_router"
	"log"
	"math/big"
	"os/exec"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/ethclient"
)

const uniswapV2RouterAddress = "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"
const forkUrl = "https://mainnet.infura.io/v3/d5eecbd3816e44aeb07866fa9fec8a91" // HARDCODED!!!!
const localUrl = "http://127.0.0.1:8545"
const dummyPrivateKey = "1cf11a582d9c69b1c180e12535a7b3a2380fd9edf8afdda73a0762a300591894" // doesn't matter which private key, we will be using the first account from Anvil

func main() {
	// Step 1: Fork Ethereum Mainnet latest block using Foundry (Anvil)
	cmd := exec.Command("anvil", "--fork-url", forkUrl, "--port", "8545", "--steps-tracing")
	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start Anvil: %v", err)
	}
	defer cmd.Process.Kill()

	fmt.Println("Forking Ethereum mainnet...")
	// Give Anvil a few seconds to initialize
	select {
	case <-time.After(5 * time.Second):
	}

	// Step 2: Verify Fork is Complete
	client, err := ethclient.Dial(localUrl)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	fmt.Println("Fork complete!")

	// Step 3: Provide User Options
	fmt.Println("Executing a simple ETH-USDC swap transaction through Uniswap V2...")

	// Perform ETH-USDC swap using Uniswap V2
	err = executeSwapTransaction(client)
	if err != nil {
		log.Fatalf("Transaction execution failed: %v", err)
	}
}

func executeSwapTransaction(client *ethclient.Client) error {
	// Get the list of accounts from Anvil
	var accounts []common.Address
	err := client.Client().Call(&accounts, "eth_accounts")
	if err != nil {
		log.Fatalf("Failed to get accounts: %v", err)
	}

	if len(accounts) == 0 {
		log.Fatalf("No accounts found")
	}

	// Instantiate the Uniswap V2 Router contract
	router, err := uniswap_v2_router.NewUniswapV2Router(common.HexToAddress(uniswapV2RouterAddress), client)
	if err != nil {
		log.Fatalf("Failed to instantiate Uniswap V2 Router contract: %v", err)
	}

	// use dummy private key
	pKey, err := crypto.HexToECDSA(dummyPrivateKey)
	if err != nil {
		log.Fatalf("Failed to convert private key: %v", err)
	}

	// Extract the public key from the private key
	pubKey := pKey.Public()

	// Type assert the public key to *ecdsa.PublicKey
	pubKeyECDSA, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatalf("Failed to assert type: public key is not of type *ecdsa.PublicKey")
	}

	// Get the Ethereum address from the public key
	from := crypto.PubkeyToAddress(*pubKeyECDSA)

	// Give fake balance to the sender
	err = client.Client().Call(nil, "anvil_setBalance", from, "0x56bc75e2d63100000") // 1 ETH
	if err != nil {
		log.Fatalf("Failed to impersonate account: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(pKey, big.NewInt(1))
	if err != nil {
		log.Fatalf("Failed to create auth: %v", err)
	}
	auth.From = from
	auth.Value = big.NewInt(1e18) // 1 ETH

	// Build the transaction
	amountOutMin := big.NewInt(1000) // Minimal amount of USDC to receive
	path := []common.Address{
		common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), // WETH
		common.HexToAddress("0xA0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"), // USDC
	}
	to := from
	deadline := big.NewInt(time.Now().Add(1 * time.Minute).Unix())

	tx, err := router.SwapExactETHForTokens(auth, amountOutMin, path, to, deadline)
	if err != nil {
		log.Fatalf("Failed to send swap transaction: %v", err)
	}

	fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())

	return traceCall(client, auth, tx)
}

func traceCall(client *ethclient.Client, auth *bind.TransactOpts, tx *types.Transaction) error {
	msg := ethereum.CallMsg{
		From:     auth.From,
		To:       tx.To(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice(),
		Value:    tx.Value(),
		Data:     tx.Data(),
	}

	tracerString := "callTracer"
	traceConfig := &tracers.TraceConfig{
		Tracer: &tracerString,
	}

	var traceResult interface{}
	err := client.Client().CallContext(context.Background(), &traceResult, "debug_traceCall", msg, "latest", traceConfig)
	if err != nil {
		log.Fatalf("Failed to trace call: %v", err)
	}

	fmt.Printf("Trace result: %v\n", traceResult)
	return nil
}
