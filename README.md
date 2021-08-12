# Miner Local Test

This branch is used for local test mining. In this branch the mining algorithm GPow and it's implementation is the same as the main branch with the same difficulty.
And miners can do local testing without connect to Evanesco network.

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




