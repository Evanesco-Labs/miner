module github.com/Evanesco-Labs/Miner

go 1.14

require (
	github.com/consensys/gnark v0.4.0
	github.com/consensys/gnark-crypto v0.4.1-0.20210428083642-6bd055b79906
	github.com/ethereum/go-ethereum v1.10.6
	github.com/google/uuid v1.2.0
	github.com/stretchr/testify v1.7.0
	gopkg.in/urfave/cli.v1 v1.20.0
)

replace (
	github.com/consensys/gnark v0.4.0 => /Users/houmy/Documents/GoPro/src/github.com/ConsenSys/gnark
	github.com/consensys/gnark-crypto v0.4.1-0.20210428083642-6bd055b79906 => /Users/houmy/Documents/GoPro/src/github.com/ConsenSys/gnark-crypto
)
