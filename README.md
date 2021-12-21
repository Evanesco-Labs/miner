# miner

Miner implements the Pow Algorithm of GPow.

## Overview of Mining

Evanesco will reward a lucky miner once every fixed block period. We call this period the Mining Epoch.

A miners determine his Challenge Height based on VRF at the beginning of each Mining Epoch. Challenge Height is the "
time" when miner need to do some hard work. When the block height reaches the Challenge Height, the miner starts to work
on a ZKP problem. This will take about one minute.

After the problem is solved, miner submit the results to an Evanesco node. A problem result
can be scored. In the final block of this Mining Epoch, result with best score wins, and it's Coinbase Address gets rewards.

### Miner Address and Coinbase Address

There are two different addresses for mining, miner address and coinbase address.
**Miner Address** is used to determine the challenge block of each round and sign the submitted work.

**Coinbase Address** is used to receive mining rewards. If **Coinbase Address** is not set with command line parameters
or in config file, it will be the same as **Miner Address** by default.

**Mining rewards will be sent to Coinbase Address, NOT Miner address!**

## Quick Start EVA Miner
Follow this quick start guide to start miner on EVA network.
### 1. Download Miner
Download in [github release](https://github.com/Evanesco-Labs/miner/releases).

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

Before starting to mine, you also need to download a ZKP ProveKey file.

Click this IPFS [Download Link](https://ipfs.io/ipfs/QmQL4k1hKYiW3SDtMREjnrah1PBsak1VE3VgEqTyoDckz9) will download a file named `QmQL4k1hKYiW3SDtMREjnrah1PBsak1VE3VgEqTyoDckz9`.

Copy `QmQL4k1hKYiW3SDtMREjnrah1PBsak1VE3VgEqTyoDckz9` to the unzipped miner directory.

### 3. Generate Miner Account
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

### 4. Start Mining

The following command will startup miner with Miner Address derived from `keyfile.json` and set the Coinbase Address the same as Miner Address by default.

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

After miner starts, it waits for the new mining epoch, the following prints in log:
```shell
INFO [09-03|21:19:57.780] waiting for next mining epoch
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

3. Wait till the blockchain reach Challenge Height. 
   The waiting time is printed like the following. It will wait 420 seconds in this example.

```shell
INFO [09-03|21:21:13.180] waiting for challenge block              time duration (second)=420
```

5. When reaches the Challenge height, miner start to solve the ZKP problem.
   It takes about 1 to 2 minutes, depending on your device.
   
```shell
INFO [09-03|21:28:13.140] start working ZKP problem
INFO [09-03|21:29:21.203] ZKP problem finished
```

4. ZKP problem solved, miner submit result. Log prints:

```shell
INFO [08-24|15:27:25.153] submit work                              
miner address=<Your miner address>
coinbase address=<Your Coinbase address> 
score=114387952695245842542743197336030803861427527497199672381530456658266076847982
```

5. Wait for the next epoch. EVA mining epoch interval is 100 blocks.
   The waiting duration is printed like the following. 
   It will wait 294 seconds in this example.
```shell
INFO [09-03|21:36:24.472] waiting for next mining epoch            time duration (second)=294
```

6. If your mining work get the best score in this mining epoch, you will get mining reward and log prints:

```shell
INFO [09-18|12:18:28.865] Congratulations you got the best score!  height=215,500
```

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

### ZKP Prove Key
ZKP Prove Key: [Download Link](https://ipfs.io/ipfs/QmQL4k1hKYiW3SDtMREjnrah1PBsak1VE3VgEqTyoDckz9)

Before starting to mine, you also need to download a ZKP prove key. This is a unique ZKP prove key containing ZKP CRS parameters. Miner have to
load this ZKP prove key to start GPow work.

### Parameters and Configs

You can start mining with subcommand `mine`. There are some parameters for mining configuration：

- **--pk**: Path of the ZKP prove key (default: "./QmQL4k1hKYiW3SDtMREjnrah1PBsak1VE3VgEqTyoDckz9").
- **--key**: Path of the key file as miner address (default: "./keyfile.json").
- **--coinbase**: Coinbase Address where mining reward will be sent to. No need to set this flag if you've already set coinbase address in Fortress wallet.
- **--passwordfile**: The file that contains the password for the keyfile.  
- **--config**: Config file path (default: "./config.yml").
- **--url**: Set WebSocket raw url of Evanesco full node to connect to. (miner connects to official seed nodes by default)

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
zkp_prove_key: ./QmQL4k1hKYiW3SDtMREjnrah1PBsak1VE3VgEqTyoDckz9
miner_key: ./keyfile.json
coinbase_address: "Fill in an address to receive mining reward"
```


### Error Logs 
**If any log with `ERROR` prefix  prints, your mining work is not able to be successfully accepted. 
Please follow the instruction to fix it.**

- Coinbase Address Incorrect:

  This error may happen, if you've set coinbase address in Fortress and also use `--coinbase` flag when you start your coinbase. 
  
  If these two coinbase addresses are not the same, log prints an error:

  `ERROR[09-18|11:21:50.702] coinbase address conflict, check the coinbase address setting in Fortress `

  Fix: Check your coinbase settings and change your coinbase in Fortress or unset the `--coinbase`.

- Miner Address Not Valid:

  If your miner address is not staked or not in the valid time period due to early or overtime, log prints an error:

  `ERROR[09-18|11:21:50.702] miner address not staked or not in valid time period`

  Fix: Check your miner address in Fortress, then stake for your miner address or wait till valid time period.

## Local Test

Check git branch `localtest` to locally test your mining device, no need to connect to real Evanesco node.
