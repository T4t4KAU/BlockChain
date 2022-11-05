# Coin 0.11 数字货币交易系统

使用Go开发的简易数字货币交易系统，本质上是一个分布式账本，借鉴于BitCoin，仅作学习与娱乐之用，还在持续完善中

## 存储

区块链使用区块(block)保存交易数据，以数据库为数据载体，区块链只对添加有效，其他操作均无效

```go
// Block 区块基本结构和功能管理
type Block struct {
	TimeStamp     int64          // 区块时间戳
	Hash          []byte         // 哈希值
	PrevBlockHash []byte         // 前区块哈希
	Height        int64          // 区块高度
	Txs           []*Transaction // 交易数据
	Nonce         int64          // POW计算次数
}
```

区块按时间序列化区块数据，整个网络有一个最终确定状态，每一个区块在链上都是可追溯的，可以遍历整个区块链最后达到创世区块

本程序，暂时使用多个端口来模拟分布式节点，以后会逐步改进

如下演示创建区块链

```
$ export NODE_ID=3000
$ ./coin createwallet
 ____  _            _     ____ _           _       
| __ )| | ___   ___| | __/ ___| |__   __ _(_)_ __  
|  _ \| |/ _ \ / __| |/ / |   | '_ \ / _` | | '_ \ 
| |_) | | (_) | (__|   <| |___| | | | (_| | | | | |
|____/|_|\___/ \___|_|\_\\____|_| |_|\__,_|_|_| |_|
                                                   
NODE ID: 3000
[1L44XkA9ezF5Dp4iUjhHJ9kPK6Y7rRxYNh]
$ ./coin createwallet
 ____  _            _     ____ _           _       
| __ )| | ___   ___| | __/ ___| |__   __ _(_)_ __  
|  _ \| |/ _ \ / __| |/ / |   | '_ \ / _` | | '_ \ 
| |_) | | (_) | (__|   <| |___| | | | (_| | | | | |
|____/|_|\___/ \___|_|\_\\____|_| |_|\__,_|_|_| |_|
                                                   
NODE ID: 3000
[1FAMvau5FGq1H2CcCr7hkHdkUpiwR8ze9i]
```

首先执行`export NODE_ID=3000`，设置节点ID，即端口号

之后执行两条命令，创建两个钱包地址，钱包是一个管理密钥对的结构，其地址是base58编码

现在假设第一个地址属于Alice，第二个地址属于Bob

```
$ ./coin accounts
 ____  _            _     ____ _           _       
| __ )| | ___   ___| | __/ ___| |__   __ _(_)_ __  
|  _ \| |/ _ \ / __| |/ / |   | '_ \ / _` | | '_ \ 
| |_) | | (_) | (__|   <| |___| | | | (_| | | | | |
|____/|_|\___/ \___|_|\_\\____|_| |_|\__,_|_|_| |_|
                                                   
NODE ID: 3000
account list:
 [1L44XkA9ezF5Dp4iUjhHJ9kPK6Y7rRxYNh]
 [1FAMvau5FGq1H2CcCr7hkHdkUpiwR8ze9i]
```

输入以上命令可以查看当前钱包列表

选择一个地址来创建创世区块，产生整个链上第一笔交易，并且是coinbase交易:

```
$ ./coin createchain -address 1L44XkA9ezF5Dp4iUjhHJ9kPK6Y7rRxYNh
 ____  _            _     ____ _           _       
| __ )| | ___   ___| | __/ ___| |__   __ _(_)_ __  
|  _ \| |/ _ \ / __| |/ / |   | '_ \ / _` | | '_ \ 
| |_) | | (_) | (__|   <| |___| | | | (_| | | | | |
|____/|_|\___/ \___|_|\_\\____|_| |_|\__,_|_|_| |_|
                                                   
NODE ID: 3000
txHash: 7aa79687af0d596cf9c55302990ff7e6c8a2e13b22debb7ddea52cc8eb022306
```

现在就可以打印整个区块链:

```
$ ./coin printchain
 ____  _            _     ____ _           _       
| __ )| | ___   ___| | __/ ___| |__   __ _(_)_ __  
|  _ \| |/ _ \ / __| |/ / |   | '_ \ / _` | | '_ \ 
| |_) | | (_) | (__|   <| |___| | | | (_| | | | | |
|____/|_|\___/ \___|_|\_\\____|_| |_|\__,_|_|_| |_|
                                                   
NODE ID: 3000
Blockchain complete information...
--------------------------------------------------------------------------------------------------------------------------------------------------------------------
Hash:0000ec49d001302725c92d3a825dc6255afb50133d39206f3d7f92439873f5bf
PrevBlockHash:
TimeStamp:1667615813
Height:1
Nonce: 31851
        tx-hash: 7aa79687af0d596cf9c55302990ff7e6c8a2e13b22debb7ddea52cc8eb022306
        input...
                vin-txHash: 
                vin-vout: -1
                vin-PublicKey: 
                vin-Signature: 
        output...
                vout-value: 10
                vout-Ripemd160Hash: d0fe95a3df506db0ec740007cd497c6efc9de34b
--------------------------------------------------------------------------------------------------------------------------------------------------------------------
```

第一条交易已经保存在第一个区块中，对于coinbase交易，没有任何交易输入，只有交易输出，交易系统采用这种方式来发行货币

打包交易，从而产生一个区块，在这一过程中要进行POW共识算法，计算一系列数学问题，求解完成后，负责打包交易的账户Alice就会得到一定系统奖励，即所谓的"挖矿"，交易输出含有钱包地址信息，对应的钱包可以使用其私钥来解锁这笔财产进行花费

区块的哈希值是基于merkle tree生成的

## 交易

交易基于UTXO账户模型，UTXO即Unspent Transaction Output，指的是未花费交易输出

```go
// UTXO 结构
type UTXO struct {
	TxHash []byte    // UTXO对应哈希
	Index  int       // 所属交易输出列表中索引
	Output *TxOutput // 交易输出
}
```

UTXO不记录最终状态，而是交易事件，将一个地址上所有的UTXO全部找出来加和，即可得到该地址总的余额

查询地址的余额:

```
$ ./coin getbalance -address 1L44XkA9ezF5Dp4iUjhHJ9kPK6Y7rRxYNh
 ____  _            _     ____ _           _       
| __ )| | ___   ___| | __/ ___| |__   __ _(_)_ __  
|  _ \| |/ _ \ / __| |/ / |   | '_ \ / _` | | '_ \ 
| |_) | | (_) | (__|   <| |___| | | | (_| | | | | |
|____/|_|\___/ \___|_|\_\\____|_| |_|\__,_|_|_| |_|
                                                   
NODE ID: 3000
utxo-txhash: 
utxo-index: 0
utxo-Ripemd160Hash: d0fe95a3df506db0ec740007cd497c6efc9de34b
utxo-value: a
balance of address[1L44XkA9ezF5Dp4iUjhHJ9kPK6Y7rRxYNh]: 10
```

发起转账:

```
$ ./coin send -from "[\"1L44XkA9ezF5Dp4iUjhHJ9kPK6Y7rRxYNh\"]" -to "[\"1FAMvau5FGq1H2CcCr7hkHdkUpiwR8ze9i\"]" -amount "[\"2\"]"
 ____  _            _     ____ _           _       
| __ )| | ___   ___| | __/ ___| |__   __ _(_)_ __  
|  _ \| |/ _ \ / __| |/ / |   | '_ \ / _` | | '_ \ 
| |_) | | (_) | (__|   <| |___| | | | (_| | | | | |
|____/|_|\___/ \___|_|\_\\____|_| |_|\__,_|_|_| |_|
                                                   
NODE ID: 3000
FROM: [1L44XkA9ezF5Dp4iUjhHJ9kPK6Y7rRxYNh]
TO: [1FAMvau5FGq1H2CcCr7hkHdkUpiwR8ze9i]
AMOUNT: [2]
money: 10
the block is added
```

打印区块链:

```
$ ./coin printchain
 ____  _            _     ____ _           _       
| __ )| | ___   ___| | __/ ___| |__   __ _(_)_ __  
|  _ \| |/ _ \ / __| |/ / |   | '_ \ / _` | | '_ \ 
| |_) | | (_) | (__|   <| |___| | | | (_| | | | | |
|____/|_|\___/ \___|_|\_\\____|_| |_|\__,_|_|_| |_|
                                                   
NODE ID: 3000
Blockchain complete information...
--------------------------------------------------------------------------------------------------------------------------------------------------------------------
Hash:0000a8148fb7c735274e12da0f970bb8f9b0bb0a3d0e256cdd37c9db2f030d26
PrevBlockHash:0000ec49d001302725c92d3a825dc6255afb50133d39206f3d7f92439873f5bf
TimeStamp:1667616710
Height:2
Nonce: 90423
        tx-hash: e88373c729b2245a101fa037c1a39046260297a707176ba08004666de38b1793
        input...
                vin-txHash: 7aa79687af0d596cf9c55302990ff7e6c8a2e13b22debb7ddea52cc8eb022306
                vin-vout: 0
                vin-PublicKey: 71450cf3dd3b8c5d7391301171d85634783a99911b7134fe00ce00f9e368d75630e9c6d6d44587f57ffcf8e89050436d1af0ebeb1df610eed126f8cdc52ed158
                vin-Signature: af916cc0e7f99d8f0ae2f3158d1737bf1586da308266e803f2a86ad494db72c11667d81c429df2177d98a72783eca82b4216f3058d5a768fccb2940e67432c7c
        output...
                vout-value: 2
                vout-Ripemd160Hash: 9b56f8932f210c06b6ff9d7155011d8106a3507f
                vout-value: 8
                vout-Ripemd160Hash: d0fe95a3df506db0ec740007cd497c6efc9de34b
        tx-hash: 9470fd18eb8e1e61ab71325a10bbdb38e2cb7494236084189d860f72b323b925
        input...
                vin-txHash: 
                vin-vout: -1
                vin-PublicKey: 
                vin-Signature: 
        output...
                vout-value: 10
                vout-Ripemd160Hash: d0fe95a3df506db0ec740007cd497c6efc9de34b
--------------------------------------------------------------------------------------------------------------------------------------------------------------------

--------------------------------------------------------------------------------------------------------------------------------------------------------------------
Hash:0000ec49d001302725c92d3a825dc6255afb50133d39206f3d7f92439873f5bf
PrevBlockHash:
TimeStamp:1667615813
Height:1
Nonce: 31851
        tx-hash: 7aa79687af0d596cf9c55302990ff7e6c8a2e13b22debb7ddea52cc8eb022306
        input...
                vin-txHash: 
                vin-vout: -1
                vin-PublicKey: 
                vin-Signature: 
        output...
                vout-value: 10
                vout-Ripemd160Hash: d0fe95a3df506db0ec740007cd497c6efc9de34b
--------------------------------------------------------------------------------------------------------------------------------------------------------------------
```

如上所示，最新区块保存了最新一笔转账交易信息，交易输入就来自上面的coinbase交易输出，因为出现找零，所以这笔交易的输出产生两部分，一笔指向发起者自己(地址)，一笔指向接收方的钱包地址，交易信息中还附带交易签名和公钥，用于校验数据完整性

在上述交易过程中，Alice持有自己的私钥，存放在她的钱包中，凭借私钥可以判断某个UTXO是否属于自己，这意味着她可以花费该UTXO，那么该UTXO就被这笔交易引用了，作为该交易的输入

Go来保证随机数熵源足够可靠，私钥是一个随机数，使用椭圆曲线乘法来产生其对应的公钥，再由此产生钱包地址，哈希采用SHA256

## 分布式

目前先使用端口来模拟分布式节点，以后会采用微服务来进行替换

选择一个节点来作为主节点，其余都是从节点，从节点会向主节点发送数据，当一个从节点上线时，会先和主节点校验版本，也就是区块链高度，数据同步到最新版本，保证数据的一致性

主节点设置为3000

```
$ ./coin start
 ____  _            _     ____ _           _       
| __ )| | ___   ___| | __/ ___| |__   __ _(_)_ __  
|  _ \| |/ _ \ / __| |/ / |   | '_ \ / _` | | '_ \ 
| |_) | | (_) | (__|   <| |___| | | | (_| | | | | |
|____/|_|\___/ \___|_|\_\\____|_| |_|\__,_|_|_| |_|
                                                   
NODE ID: 3000
node address: localhost:3000

```

从节点为9000

```
$ ./coin start
 ____  _            _     ____ _           _       
| __ )| | ___   ___| | __/ ___| |__   __ _(_)_ __  
|  _ \| |/ _ \ / __| |/ / |   | '_ \ / _` | | '_ \ 
| |_) | | (_) | (__|   <| |___| | | | (_| | | | | |
|____/|_|\___/ \___|_|\_\\____|_| |_|\__,_|_|_| |_|
                                                   
NODE ID: 9000
node address: localhost:9000
send request to server[localhost:3000]...
```

分布式相关功能还在持续开放中
