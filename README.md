# hello-paradigm
Demonstrating an EVM simulation including tracing against a forked Ethereum mainnet in Go

## Requirements:
- Foundry
- Go 1.21
- Geth package `go get -u github.com/ethereum/go-ethereum`

## Notes
- ABI binding generated using abigen, `abigen --abi=uniswap_v2_router.abi --out=uniswap_v2_router/uniswap_v2_router.go --pkg=uniswap_v2_router` 
- Infura API key is HARD CODED in the `main.go` file for demonstration purposes. It's on a free account, but will be deleting soon after this repo is done with
- run using `go run main.go`. If doesn't work, do `go get` just in case ;)