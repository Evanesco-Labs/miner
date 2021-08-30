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

## Quick Start Avis TestNet Miner
Follow this quick start guide to start miner on AVIS TestNet. Please note that when the AVIS testnet is officially launched on September 7, it will be lauched from the Genesis Block, so all previous mining rewards and transactions will be cleared.
### 1. Download Miner
Download in github release.

MacOS: [miner-darwin.zip](https://github.com/Evanesco-Labs/miner/releases/download/v0.0.1/miner-darwin.zip)

Windows: [miner-windows.zip](https://github.com/Evanesco-Labs/miner/releases/download/v0.0.1/miner-windows.zip)

Linux: [miner-linux.zip](https://github.com/Evanesco-Labs/miner/releases/download/v0.0.1/miner-linux.zip)

#### MacOS
Download the file `miner-darwin.zip`  and unzip the file.
Start terminal and go to the unzipped directory `miner-darwin` set executable permission with this command:

```
chmod 777 ./miner
```
#### Windows
Download the file `miner-windows.zip`  and unzip the file.
#### Linux
Download the file `miner-linux.zip`  and unzip the file.
Go to the unzipped directory `miner-linux` set executable permission with this command:
```
chmod 777 ./miner
```

### 2. Download ZKP ProveKey

Before starting to mine, you also need to download a ZKP ProveKey file `provekey.txt`.

Download from IPFS [Download Link](https://ipfs.io/ipfs/QmNpJg4jDFE4LMNvZUzysZ2Ghvo4UJFcsjguYcx4dTfwKx) and copy `provekey.txt` to the unzipped miner directory.

### 3. Generate Account
The following  command will generate a new key file named `keyfile.json`to derive you Miner Account. You can set a password for your keyfile and remember this password.

#### MacOS

Go to the unzipped directory `miner-darwin` in terminal and use this command:
```
./miner generate
```


#### Linux
Go to the unzipped directory `miner-linux` in command line  and use this command:
```
./miner generate
```

#### Windows
Go to the unzipped directory `miner-windows` in command line  and use this command:
```
./miner.exe generate
```

### 4. Start Miner

The following command will startup miner with Miner Address derived from `keyfile.json` and set the Coinbase Address the same as Miner Addresss by default.

#### MacOS
Go to the unzipped directory `miner-darwin` in terminal and use this command:
```
./miner mine
```

#### Linux
Go to the unzipped directory `miner-linux` in command line  and use this command:
```
./miner mine
```

#### Windows
Go to the unzipped directory `miner-windows` in command line  and use this command:
```
./miner.exe mine
```


Enter keyfile password when the following prints:

```shell
Password: 
```


Then miner starts to initialize ZKP prover, includes compiling ZKP circuit and loading ZKP prove key. You can see the
followings in log:

```shell
INFO [08-24|15:56:41.152] Compiling ZKP circuit 
INFO [08-24|15:56:46.234] Loading ZKP prove key. This takes a few minutes
```

This may take 5 to 15 minutes, please wait. When the initialization success, the following prints in log:

```shell
INFO [08-24|16:02:46.363] Init ZKP Problem worker success! 
INFO [08-24|16:02:46.380] miner start
```

After that, the log will continue to print the block height, block index:
```shell
INFO [08-25|20:11:11.187] chain status                             height=359 index=59
INFO [08-25|20:11:17.159] chain status                             height=360 index=60
INFO [08-25|20:11:23.161] chain status                             height=361 index=61
...
```

Miners follow new blocks and keep working in the following process：

1. Wait for new mining epoch. When it reaches a new epoch, log prints:

```shell
INFO [08-24|15:24:51.004] start new mining epoch
```

2. Doing VRF and get the Challenged Height of this mining epoch. Log prints:

```shell
INFO [08-24|15:24:51.008] vrf finished                             challenge height=3015 index=15
```

3. Wait till the blockchain reach Challenge Height, and start solving ZKP problem. Log prints:

```shell
INFO [08-24|15:26:21.102] start working ZKP problem
```

4. ZKP problem solved, miner submit Lottery. Log prints:

```shell
INFO [08-24|15:27:25.153] submit work                              
miner address=Ex3d9C3C64b1083b8DA73E8E247B8484309EE429ff 
coinbase address=Ex3d9C3C64b1083b8DA73E8E247B8484309EE429ff 
score=114387952695245842542743197336030803861427527497199672381530456658266076847982
```

5. Wait for the next epoch. Avis mining epoch interval is 100 blocks.


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

## Miner Specifics

To start mining, you need a key file as the miner account.

Please note that there are two different addresses for mining, miner address and coinbase address.
**Miner Address** is used to determine the challenge block of each round and sign the submitted work.
**Coinbase Address** is used to receive mining rewards. If **Coinbase Address** is not set with command line parameters
or in config file, it will be the same as **Miner Address** by default.

**Mining rewards will be sent to Coinbase Address, NOT Miner address!**

### Download ZKP Prove Key
ZKP Prove Key: [Download Link](https://ipfs.io/ipfs/QmNpJg4jDFE4LMNvZUzysZ2Ghvo4UJFcsjguYcx4dTfwKx)

Before starting to mine, you also need to download a ZKP prove key. This is a unique ZKP prove key, and miner have to
load this ZKP prove key to start GPow working.

### Parameters and Configs

You can start mining with subcommand `mine`. There are some parameters for mining configuration：

- **--pk**: Path of the ZKP prove key (default: "./provekey.txt").
- **--key**: Path of the key file as miner address (default: "./keyfile.json").
- **--coinbase**: Coinbase Address where mining reward will be sent to. It's the same as miner address by default.
- **--passwordfile**: The file that contains the password for the keyfile.  
- **--config**: Config file path (default: "./config.yml").

The following command can show you all the parameters and their uses:

```shell
./miner mine -h
```

Start mining by default configs with this command:

```shell
./miner mine 
```

Enter keyfile password when the following prints:

```shell
Password: 
```

If you don't want to enter password everytime, it is possible
to pass the password by using the --passwordfile flag pointing to a file that
contains the password.

#### Config File

You can write miner configuration into a yaml config file. Set the parameter `--config` to the path of this config file
at miner startup.

Config file for example:

```yaml
zkp_prove_key: ./provekey.txt
miner_key: ./keyfile.json
coinbase_address: ExaA460D005DE1d5e572a78f90E68d774cEB51c7fc
```

## Local Test

Check git branch `localtest` to locally test your mining device, no need to connect to real Evacesco node.
