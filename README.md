# Coin 0.11 数字货币交易系统

使用Go开发的简易数字货币交易系统，本质上是一个分布式账本，借鉴于Bit Coin，仅作学习与娱乐之用，还在持续完善中

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

第一条交易已经保存在第一个区块中，对于coinbase交易，没有任何交易输入，只有交易输出，本交易系统采用这种方式来发行货币

打包交易，从而产生一个区块，在这一过程中要进行POW共识算法，计算一系列数学问题，求解完成后，负责打包交易的账户Alice就会得到一定系统奖励，即所谓的"挖矿"，交易输出含有钱包地址信息，对应的钱包可以使用其私钥来解锁这笔财产进行花费，区块的哈希值是基于merkle tree和交易信息生成的

区块中的交易作为merkle tree的叶子节点，逐层向上生成哈希值，最终到达根节点时生成的哈系值，就是这个区块要保存的哈希值

这有别于中心化的货币发行机构，例如在USA，美联储负责货币的发行，同时也负责执行国家货币政策、监管全美银行业、维持金融系统的稳定性，并对存款机构提供相应金融服务，而与BTC这样的数字货币有关的所有事宜，包括发行、交易处理和验证都是通过分布式网络进行的，不需要专门机构来监控整个资金流动过程

由于数字货币的交易过程需要网络中每个节点的认可，且每一笔交易都被记录在区块链上，所以历史交易记录永远不用担心丢失或者被篡改

只要**数字货币基础的加密算法不被攻破**，并且保护好私钥，你的资产便是真正意义上、只属于你自己的资产，而传统货币的交易过程最终是落到银行的，所以银行系统的安全性决定了传统货币在使用过程中的安全阈值，这也表示你的资产是托管在银行的

## 交易

比特币中没有账户的概念。因此，每次发生交易，用户需要将交易记录写到比特币网络账本中，等网络确认后即可认为交易完成，比特币网络中一笔合法的交易，必须是引用某些已存在交易的 UTXO（必须是属于付款方才能合法引用）作为新交易的输入，并生成新的 UTXO（将属于收款方）

除了挖矿获得奖励的 coinbase 交易只有输出，正常情况下每个交易需要包括若干输入和输出，未经使用（引用）的交易的输出（Unspent Transaction Outputs，UTXO）可以被新的交易引用作为其合法的输入。被使用过的交易的输出（Spent Transaction Outputs，STXO），则无法被引用作为合法输入。

交易基于**UTXO交易模型**，UTXO即Unspent Transaction Output，指的是未花费交易输出

```go
// UTXO 结构
type UTXO struct {
	TxHash []byte    // UTXO对应哈希
	Index  int       // 所属交易输出列表中索引
	Output *TxOutput // 交易输出
}
```

UTXO不记录**最终状态**，而是**交易事件**，将一个地址上所有的UTXO全部找出来进行加和，即可得到该地址总的余额，而不像我们所熟知的银行账户，直接保存了一个用户的余额

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

这时打印区块链，查看新生成的区块:

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

在上述交易过程中，Alice持有自己的私钥，存放在她的钱包中，凭借私钥可以判断某个UTXO是否属于自己(解锁UTXO)，这意味着她可以花费该UTXO，那么该UTXO就被这笔交易引用了，作为该交易的输入，并且不再是UTXO，交易同时会产生输出，又诞生出一系列新的UTXO，可以再作为其他交易的输入

可以看到交易信息完全不包括仍和与个人有关的真实信息，发送方和接收方都为钱包地址，数学已经保证了人类几乎无法通过这个地址来推测出这个钱包的拥有者

私钥是随机生成的，因此保证随机数熵源足够可靠，得到私钥后，使用椭圆曲线乘法(ECDSA)来产生其对应的公钥，再由此产生钱包地址

钱包地址为公钥的双哈希值(sha256 & ripemd160)，添加校验和，最后使用base58编码得到钱包地址

上述提及的交易行为有别于传统交易，例如银行发起转账时，有一个非常重要的操作就是清结算，目前的金融活动都是由中心化的机构在做担保交易，所以每个账户的对账、清洁算也是由这些中心化的机构来处理的，而UTXO交易模型是去中心化的设计，所以没有一个中心化的机构来进行清结算

## 共识

共识机制基于POW(Proof of Work)，即工作量证明

为了防止垃圾消息泛滥，接收者并不直接接受来自任意发送者的消息，所以在一次有效的会话中，发送者需要计算一个按照规则约定难题的答案，发送给接受者的同时，需要附带验证这个答案，如果这个答案被验证有效，那么接受者才会接受这个消息，这样的难题必须具有计算不对程性

上文提到coinbase交易是由"挖矿"生成的，这过程的步骤如下:

1. 生成 Coinbase 交易，并与其他所有准备打包进区块的交易组成交易列表，并生成默克尔哈希
2. 把默克尔哈希及其他相关字段组装成区块头，将区块头（Block Header）作为工作量证明的输入，区块头中包含了前一区块的哈希，区块头一共 80 字节数据
3. 不停地变更区块头中的随机数即 nonce 的数值，也就是暴力搜索，并对每次变更后的的区块头做双重 SHA256 运算，即 SHA256(SHA256(Block_Header))），将结果值与当前网络的目标值做对比，如果小于目标值，则解题成功，工作量证明完成

上述过程在源代码中已经体现:

```go
// 共识算法
// POW: 工作量证明

const targetBit = 16 // 目标难度值

// ProofOfWork POW结构
type ProofOfWork struct {
	Block  *Block   // 要共识验证的区块
	Target *big.Int // 目标难度的哈希(大数存储)
}

// NewProofOfWork 创建一个POW对象
func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	// 数据总长度为8 满足前两位为0
	// 1 * 2 << (8-2) = 64
	target = target.Lsh(target, 256-targetBit)
	return &ProofOfWork{Block: block, Target: target}
}

// Run 计算哈希值
func (p *ProofOfWork) Run() ([]byte, int) {
	var nonce = 0 // 碰撞次数
	var hashInt big.Int
	var hash [32]byte
	// 反复计算哈希值 直到符合要求
	for {
		// 生成准备数据
		dataBytes := p.prepareData(int64(nonce))
		hash = sha256.Sum256(dataBytes) // 计算SHA256值
		hashInt.SetBytes(hash[:])
		// 检测生成的哈希值是否符合条件
		if p.Target.Cmp(&hashInt) == 1 {
			// 找到符合条件的哈希值
			break
		}
		nonce++
	}
	return hash[:], nonce
}

// 生成准备数据
func (p *ProofOfWork) prepareData(nonce int64) []byte {
	// 拼接区块属性
	timeStampBytes := utils.IntToHex(p.Block.TimeStamp)
	heightBytes := utils.IntToHex(p.Block.Height)
	// 拼接区块数据 用于后续做哈希计算
	// 将区块高度 时间戳 前区块哈希 序列化交易数据 目标难度值 碰撞次数拼接
	data := bytes.Join([][]byte{
		heightBytes, timeStampBytes,
		p.Block.PrevBlockHash,
		p.Block.HashTransaction(),
		utils.IntToHex(targetBit), utils.IntToHex(nonce),
	}, []byte{})

	return data
}
```

## P2P

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
