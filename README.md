# Miner Local Test

This branch is used for local test mining. In this branch the mining algorithm GPow and it's implementation is the same
as the main branch with the same difficulty. And miners can do local testing without connect to Evanesco network.

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

Build with the following command will generate the executable file miner.

```shell
main
```

## Start Test

### Generate Account

Use the following command to generate the keyfile for miner:

```shell
./miner generate
```

This generates a new key file `keyfile.json` in the current directory.

### Download ZKP ProveKey

Before starting to mine, you also need to download a ZKP prove key. This is a unique ZKP prove key, and miner have to
load this ZKP prove key to start GPow working.

[IPFS Download Link](https://ipfs.io/ipfs/QmNpJg4jDFE4LMNvZUzysZ2Ghvo4UJFcsjguYcx4dTfwKx)

Copy the ZKP prove key file `QmNpJg4jDFE4LMNvZUzysZ2Ghvo4UJFcsjguYcx4dTfwKx` to the same directory of executable file `miner`.

### Start Mining Test

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

5. Wait for the next Mining Epoch.

In this test, a Mining Epoch is 100 blocks, and the block rate is 6 seconds/block. 


