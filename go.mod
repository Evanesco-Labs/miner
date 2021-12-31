module github.com/Evanesco-Labs/miner

go 1.14

require (
	github.com/Evanesco-Labs/go-evanesco v1.0.1-0.20211227031128-25f06ef16c40
	github.com/google/uuid v1.2.0
	github.com/stretchr/testify v1.7.0
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	gopkg.in/urfave/cli.v1 v1.20.0
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/consensys/gnark v0.4.0 => github.com/Evanesco-Labs/gnark v0.4.1-0.20211224081327-4f8af4c8073f
	github.com/consensys/gnark-crypto v0.4.1-0.20210428083642-6bd055b79906 => github.com/Evanesco-Labs/gnark-crypto v0.4.1-0.20211220040057-c079b829266f
)
