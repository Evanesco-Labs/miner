# Miner Local Test

This branch is used for local test mining. In this branch the mining algorithm GPow and it's implementation is the same as the main branch with the same difficulty.
And miners can do local testing without connect to Evanesco network.

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
Before starting to mine, you also need to download a ZKP prove key. This is a unique ZKP prove key, and miner have to load this ZKP prove key to start GPow working.

Please download from IPFS, IPFS CID:

Copy the ZKP prove key file `provekey.txt` to the same directory of executable file `miner`.
### Start Mining Test
Start mining with this command:
```shell
./miner mine
```

Enter keyfile password when the following prints:
```shell
Password: 
```

Then miner starts to initialize ZKP prover, includes compiling ZKP circuit and loading ZKP prove key.
You can see the followings in log:
```shell
2021/08/13 15:30:37.275726 [INFO ] GID 1, [problem] Compiling ZKP circuit
2021/08/13 15:30:41.724334 [INFO ] GID 1, [problem] Loading ZKP prove key, this takes a few minutes.
```
This may take 5 to 10 minutes, please wait. When the initialization success, the following prints in log:
```shell
2021/08/13 15:47:05.869020 [INFO ] GID 1, [miner] Init ZKP Problem Worker success!
2021/08/13 15:47:05.869168 [DEBUG] GID 1, [miner] github.com/Evanesco-Labs/miner.(*Miner).NewWorker miner.go:234 worker start
2021/08/13 15:47:05.869192 [INFO ] GID 1, [miner] miner start
```
After that, the log will continue to print the block height, block index, and the best mining score in the mining epoch.

In this test, the mining epoch is 100 blocks. The block rate is 6s/block. 





