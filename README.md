# miner

Miner implements the Pow Algorithm of GPow.

## Build

Building miner requires both a Go (version 1.14 or later) and a C compiler.

Build with the following command will generate the executable file `miner`.

```shell
make
```

Once the project has been built, the following command can be used to explore all parameters and subcommands:

```shell
./miner help
```

## Manage Account

Starting mining need to load a miner account. You can generate this account with `miner` or use your existing key file.

### Generate

Generate a keyfile with the following command, this will new a key file `keyfile.json`. You can set and remember a
password for this key.

```shell
./miner generate
```

### Inspect

Print various information about the keyfile. Private key information can be printed by using the --private flag, make
sure to use this feature with great caution!

Check Address and Public key of key file `keyfile.json` with this command:

```shell
./miner inspect keyfile.json
```

## Startup Mining

To start mining, you need to have the WebSocket rpc URL of an Evanesco node to establish a connection. In addition, you
also need a key file as the miner account.

Please note that there are two different addresses for mining, miner address and coinbase address.
**Miner Address** is used to determine the challenge block of each round and sign the submitted work.
**Coinbase Address** is used to receive mining rewards. If **Coinbase Address** is not set with command line parameters
or in config file, it will be the same as **Miner Address** by default.

**Mining rewards will be sent to Coinbase Address, NOT Miner address!**

### Download ZKP Prove Key

Before starting to mine, you also need to download a ZKP prove key. This is a unique ZKP prove key, and miner have to
load this ZKP prove key to start GPow working.

Please download from IPFS, IPFS CID: 

### Startup

You can start mining with subcommand `mine`. There are We provide some parameters for mining configuration：

- **--url**:
  WebSocket URL of a running Evanesco node, for example `ws://127.0.0.1:8549`.
- **--pk**: Path of the ZKP prove key (default: "./provekey.txt").
- **--key**: Path of the key file as miner address (default: "./keyfile.json").
- **--coinbase**: Coinbase Address where mining reward will be sent to. It's the same as miner address by default.
- **--config**: Config file path (default: "./config.yml").

The following command can show you all the parameters and their uses:

```shell
./miner mine -h
```

Start mining by default configs with this command:

```shell
./miner mine --url ws://127.0.0.1:8549
```

#### Config File

You can write miner configuration into a yaml config file. Set the parameter `--config` to the path of this config file
at miner startup.

Config file for example:

```yaml
url: ws://127.0.0.1:8549
zkp_prove_key: ./provekey.txt
miner_key: ./keyfile.json
coinbase_address: ExaA460D005DE1d5e572a78f90E68d774cEB51c7fc
```

## Local Test

Check git branch `localtest` to locally test your mining device, no need to connect to real Evacesco node.
