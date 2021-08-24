export GO_BUILD=env GO111MODULE=on go build
export GO_TEST=env GO111MODULE=on go test
default: build
build:
	$(GO_BUILD) -ldflags "-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" -o ./miner github.com/Evanesco-Labs/miner