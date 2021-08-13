# miner

Miner implements the Pow Algorithm of GPow.

## Overview of Mining

Evanesco will reward a lucky miner once every fixed block period. We call this period the Mining Epoch.

A miners determine his Challenge Height based on VRF at the beginning of each Mining Epoch. Challenge Height is the "
time" when miner need to do some hard work. When the block height reaches the Challenge Height, the miner starts to work
on a ZKP problem. This will take about one minute.

After the problem is solved, miner submit the results to an Evanesco node. We call this result a Lottery. Each Lottery
can be scored. In the final block of this Mining Epoch, the Coinbase Address of the Lottery with best score will get the
reward.

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

Start mining with this command:

```shell
./miner mine
```

Enter keyfile password when the following prints:

```shell
Password: 
```

Then miner starts to initialize ZKP prover, includes compiling ZKP circuit and loading ZKP prove key. You can see the
followings in log:

```shell
2021/08/13 15:30:37.275726 [INFO ] GID 1, [problem] Compiling ZKP circuit
2021/08/13 15:30:41.724334 [INFO ] GID 1, [problem] Loading ZKP prove key, this takes a few minutes.
```

This may take 5 to 15 minutes, please wait. When the initialization success, the following prints in log:

```shell
2021/08/13 15:47:05.869020 [INFO ] GID 1, [miner] Init ZKP Problem Worker success!
2021/08/13 15:47:05.869168 [DEBUG] GID 1, [miner] github.com/Evanesco-Labs/miner.(*Miner).NewWorker miner.go:234 worker start
2021/08/13 15:47:05.869192 [INFO ] GID 1, [miner] miner start
```

After that, the log will continue to print the block height, block index.

Miners follow new blocks and work according to the following process：

1. Wait for new mining epoch. Log prints:

```shell
2021/08/13 16:20:20.987955 [INFO ] GID 106, [miner] start new mining epoch
```

2. Doing VRF and get the Challenged Height of this mining epoch. Log prints:

```shell
2021/08/13 16:20:20.988758 [INFO ] GID 111, [miner] vrf finished challenge height: 56 index: 56
```

3. Wait till the blockchain reach Challenge Height, and start solving ZKP problem. Log prints:

```shell
2021/08/13 16:25:50.428437 [INFO ] GID 167, [miner] start working ZKP problem
```

4. ZKP problem solved, miner submit Lottery. Log prints:

```shell
2021/08/13 16:26:52.522183 [INFO ] GID 167, [miner] ZKP problem finished
2021/08/13 16:26:52.522315 [INFO ] GID 106, [miner] submit new work 
miner: Exb5Cf748be8414293B83fC9eDCfc47cF8Ece05Eec 
score: 99012458875274928854094129045709882384886482989475090027812147624624819215500 
coinbase: Exb5Cf748be8414293B83fC9eDCfc47cF8Ece05Eec
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
