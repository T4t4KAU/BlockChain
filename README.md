# 基于Go开发公链

区块链开发

## 初始化区块

定义区块结构和链结构

### 定义区块结构

创建一个block包，编写区块相关的逻辑

首先定义区块的结构:

```go
type Block struct {
	TimeStamp     int64  // 时间戳
	Data          []byte // 交易数据
	Hash          []byte // 区块哈希
	PrevBlockHash []byte // 前区块哈希
	Height        int64  // 区块高度
}
```

创建一个区块:

```go
func NewBlock(data []byte, prevBlockHash []byte, height int64) *Block {
	block := &Block{
		TimeStamp:     time.Now().Unix(),
		Data:          data,
		Height:        height,
		PrevBlockHash: prevBlockHash,
	}
	block.SetHash()
	return block
}
```

对区块结构进行赋值

编写一个方法，为其设置哈希值

```go
// SetHash 计算区块哈希
func (b *Block) SetHash() {
	timeStampBytes := IntToHex(b.TimeStamp)
	heightBytes := IntToHex(b.Height)
	// 调用SHA256实现哈希生成
	blockBytes := bytes.Join([][]byte{
		heightBytes, timeStampBytes,
		b.PrevBlockHash, b.Data,
	}, []byte{})
	hash := sha256.Sum256(blockBytes)
	b.Hash = hash[:]
}
```

哈希值为4个成员的sha256算法求得

IntToHex函数的作用是将64位整数转化为字节

```go
func IntToHex(data int64) []byte {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.BigEndian, data)
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}
```

这样就实现好了NewBlock方法，作用是创建一个区块

### 定义区块链结构

区块链要将区块串联，如同一条链表

定义链结构:

```go
type Chain struct {
	Blocks []*Block
}
```

定义一个方法用于增加区块:

```go
func (c *Chain) AddBlock(data []byte) {
	index := len(c.Blocks) - 1
	height := c.Blocks[index].Height + 1
	prevBlockHash := c.Blocks[index].PrevBlockHash
	newBlock := NewBlock(data, prevBlockHash, height)
	c.Blocks = append(c.Blocks, newBlock)
}
```

增加区块时，只须传入数据，其高度等于前区块的高度加1，前区块的哈希值直接从链中获取，之后创建一个区块并添加到切片

主函数运行:

```go
package main

import (
	"fmt"
	"testchain/block"
)

func main() {
	chain := block.CreateBlockChainGenesisBlock()
	chain.AddBlock([]byte("alice send 100 btc to bob"))
	chain.AddBlock([]byte("alice send 5 btc to marry"))
	for _, b := range chain.Blocks {
		fmt.Printf("prevBlockHash: %x currentHash: %x\n", b.PrevBlockHash, b.Hash)
	}
}
```

输出:

```go
prevBlockHash:  currentHash: 92986e20ca67ec28bcc57b1dbf338981406d858b4801f049cce7f5e0caedc979
prevBlockHash: 92986e20ca67ec28bcc57b1dbf338981406d858b4801f049cce7f5e0caedc979 currentHash: 37f3fbc651d541971391fdc508369d0dfbafae12eeca38ac299628207dbd26cc
prevBlockHash: 37f3fbc651d541971391fdc508369d0dfbafae12eeca38ac299628207dbd26cc currentHash: bebeb54a8b43aacdd55c5b37be538a4d55e5bcd8a27c2275f64001c5de24daa1
```

## POW工作量证明

为了实现POW，要设计一个问题，让计算机去求解，这个问题要求计算难，验证易

如下创建一个POW对象，传入指定的区块:

```go
// NewProofOfWork 创建POW实例
func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	// 数据总长度为8位 满足前两位为0 才算解决问题
	target = target.Lsh(target, 256-targetBit) // 左移
	return &ProofOfWork{Block: block, Target: target}
}
```

target是一个大整数

下面是数据准备，将要求的变量添加进切片

```go
func (p *ProofOfWork) prepareData(nonce int64) []byte {
	var data []byte
	// 拼接区块属性 进行哈希计算
	timeStampBytes := IntToHex(p.Block.TimeStamp)
	heightBytes := IntToHex(p.Block.Height)
	data = bytes.Join([][]byte{
		heightBytes, timeStampBytes,
		p.Block.PrevBlockHash, p.Block.Data,
		IntToHex(nonce), IntToHex(targetBit)}, []byte{})
	return data
}
```

下面进行哈希碰撞

```go
// Run 哈希碰撞
func (p *ProofOfWork) Run() ([]byte, int) {
	var hash [32]byte
	var nonce = 0 // 碰撞次数
	var hashInt big.Int
	for {
		dataBytes := p.prepareData(int64(nonce))
		hash = sha256.Sum256(dataBytes)
		hashInt.SetBytes(hash[:])
		// 检测生成的哈希值是否符合条件
		if p.Target.Cmp(&hashInt) == 1 {
			break
		}
		nonce++
	}
	fmt.Printf("Number of hash collision:%d\n", nonce)  // 输出碰撞次数
	return hash[:], nonce
}
```

对上述得到的序列取SHA256值，将其转化为大整数和target进行比较，如果小于则算问题得解

运行:

```go
Number of hash collision:72023
Number of hash collision:167551
Number of hash collision:17189
prevBlockHash:  currentHash: 0000d1f8afc91a961847acffcd4394b01cd01ca9723fd013af66fa3cbbb44537
prevBlockHash: 0000d1f8afc91a961847acffcd4394b01cd01ca9723fd013af66fa3cbbb44537 currentHash: 00009b7bd3e8f3c2bc40fe4099747b60560ff2ddf5f07ee9d5c40119c821e168
prevBlockHash: 00009b7bd3e8f3c2bc40fe4099747b60560ff2ddf5f07ee9d5c40119c821e168 currentHash: 000065e9943c01a6cc26b3ee18d9e644dfccfd1f98d0a6a6cde54d041b303665
```

## 实现持久化

BoltDB是一个key/value数据库，基于此进行持久化

下载方式: `$ go get github.com/boltdb/bolt`

扩展Chain结构，增加成员，将区块和数据库关联

```go
// Chain 区块链的基本结构
type Chain struct {
	Blocks []*Block // 区块的切片
	DB     *bolt.DB // 数据库连接
	Tip    []byte   // 最新区块的哈希
}
```

定义数据库相关信息:

```go
const (
	dbName         = "block.db" // 数据库名称
	blockTableName = "blocks"   // 表名称 存储区块
)
```

修改区块链的创建函数:

```go
// CreateBlockChainWithGenesisBlock 创建区块链
func CreateBlockChainWithGenesisBlock() *Chain {
	var blockHash []byte
	// 创建或打开数据库
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	// 创建桶 将生成的创世区块存入数据库
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName)) // 获取指定桶
		if b == nil {
			// 如果不存在则创建桶
			bucket, e := tx.CreateBucket([]byte(blockTableName))
			if e != nil {
				log.Panic(e)
			}
			// 创建创世区块
			genesisBlock := CreateGenesisBlock([]byte("init blockchain"))
			// key-哈希  value-序列化区块数据
			e = bucket.Put(genesisBlock.Hash, genesisBlock.Serialize()) // 储存数据
			if e != nil {
				log.Panic(e)
			}
			blockHash = genesisBlock.Hash
			// 存储最新区块的哈希
			e = bucket.Put([]byte("l"), genesisBlock.Hash)
			if e != nil {
				log.Panic(e)
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return &Chain{DB: db, Tip: blockHash}
}
```

上述函数打开数据库，获取存储桶，将区块的哈希及其序列化的数据存入桶，同时设置一个最新区块哈希，以l为键

序列化和反序列化的实现:

```go
// Serialize 区块结构序列化
func (b *Block) Serialize() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer) // 创建编码对象
	if err := encoder.Encode(b); err != nil {
		log.Panic(err)
	}
	return buffer.Bytes()
}

// DeserializeBlock 区块数据反序列化
func DeserializeBlock(blockBytes []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes)) // 创建解码对象
	if err := decoder.Decode(&block); err != nil {
		log.Panic(err)
	}
	return &block
}
```

于是添加区块的函数也要修改:

```go
// AddBlock 添加区块到区块链
func (c *Chain) AddBlock(data []byte) {
	// 更新区块数据
	err := c.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			blockBytes := b.Get(c.Tip) // 获取最新区块
			latestBlock := DeserializeBlock(blockBytes)
			newBlock := NewBlock(latestBlock.Height+1, latestBlock.Hash, data) // 创建区块
			e := b.Put(newBlock.Hash, newBlock.Serialize())
			if e != nil {
				log.Panic(e)
			}
			e = b.Put([]byte("l"), newBlock.Hash)
			if e != nil {
				log.Panic(e)
			}
			c.Tip = newBlock.Hash
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
```

添加区块时，原来的最新区块作为前区块，将新区块的哈希值和序列化后的数据添加进数据库，同时设置最新区块

实现一个遍历函数，用于遍历整条链

```go
// PrintChain 遍历数据库 输出所有区块信息
func (c *Chain) PrintChain() {
	// 读取数据库
	fmt.Println("Blockchain complete information...")
	var curBlock *Block
	iter := c.Iterator() // 获取迭代器对象
	// 从最新区块开始循环读取
	for {
		fmt.Println("---------------------------------------------------------------")
		// 从数据库读取数据
		curBlock = iter.Next()
		curBlock.PrintBlock()
		fmt.Println("---------------------------------------------------------------")

		// 遍历到创世区块则选择退出
		var hashInt big.Int
		hashInt.SetBytes(curBlock.PrevBlockHash)
		// 比较相等 则遍历到创世区块
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
}
```

其中额外定义了一个迭代器:

```go
type ChainIterator struct {
	DB          *bolt.DB // 迭代目标
	CurrentPath []byte   // 当前迭代目标的哈希
}

func (c *Chain) Iterator() *ChainIterator {
	return &ChainIterator{
		c.DB, c.Tip,
	}
}

func (it *ChainIterator) Next() *Block {
	var block *Block
	err := it.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			currentBlockBytes := b.Get(it.CurrentPath)
			// 更新迭代器中区块的哈希值
			block = DeserializeBlock(currentBlockBytes)
			it.CurrentPath = block.PrevBlockHash
		}
		return nil
	})
	if err != nil {
		log.Panicf("iterator the db failed: %v\n", err)
	}

	return block
}
```

Next方法用于获得到下一个区块

定义一个PrintChain方法:

```go
// PrintChain 遍历数据库 输出所有区块信息
func (c *Chain) PrintChain() {
	// 读取数据库
	fmt.Println("Blockchain complete information...")
	var curBlock *Block
	iter := c.Iterator() // 获取迭代器对象
	// 从最新区块开始循环读取
	for {
		fmt.Println("---------------------------------------------------------------")
		// 从数据库读取数据
		curBlock = iter.Next()
		curBlock.PrintBlock()
		fmt.Println("---------------------------------------------------------------")

		// 遍历到创世区块则选择退出
		var hashInt big.Int
		hashInt.SetBytes(curBlock.PrevBlockHash)
		// 比较相等 则遍历到创世区块
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
}
```

## 定义命令行接口

为程序定义命令行接口，通过命令行来操作程序，如下实现了一个cmd包

```go
package cmd

import (
	"blockchain/block"
	"flag"
	"fmt"
	"log"
	"os"
)

// Client 客户端对象
type Client struct {
	BlockChain *block.Chain
}

func PrintUsage() {
	fmt.Println("Usage: ")
	// 创建区块链
	fmt.Println("\tcreateblockchain -- create blockchain")
	// 添加区块
	fmt.Println("\taddblock -- add block")
	// 打印完整信息
	fmt.Println("\tprintchain -- print blockchain")
}

// 初始化区块链
func (cli *Client) createBlockChain() {
	cli.BlockChain = block.CreateBlockChainWithGenesisBlock()
}

// 添加区块
func (cli *Client) addBlock(data string) {
	cli.BlockChain = block.GetBlockChainObject()
	cli.BlockChain.AddBlock([]byte(data))
}

// 打印完整区块链信息
func (cli *Client) printChain() {
	cli.BlockChain = block.GetBlockChainObject()
	cli.BlockChain.PrintChain()
}

func IsValidArgs() {
	if len(os.Args) < 2 {
		PrintUsage()
		// 直接退出
		os.Exit(1)
	}
}

// Run 命令行
func (cli *Client) Run() {
	// 检测命令行参数个数
	IsValidArgs()
	// 新建相关命令 添加区块
	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	// 输出区块链完整信息
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	// 创建区块链
	createChainWithGenesisBlockCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	// 数据参数
	flagAddBlockArg := addBlockCmd.String("data", "sent 100 btc to player", "add block")
	switch os.Args[1] {
	case "addblock": // 添加区块
		if err := addBlockCmd.Parse(os.Args[2:]); err != nil {
			log.Panicf("parse addBlockCmd failed %v\n", err)
		}
	case "printchain": // 输出公链
		if err := printChainCmd.Parse(os.Args[2:]); err != nil {
			log.Panicf("parse printChainCmd failed %v\n", err)
		}
	case "createblockchain": // 创建区块链
		if err := createChainWithGenesisBlockCmd.Parse(os.Args[2:]); err != nil {
			log.Panicf("parse createBLCWithGenesisBlockCmd failed %v\n", err)
		}
	default:
		// 没有传递任何命令或者传递的命令不在命令列表中
		PrintUsage()
		os.Exit(1)
	}
	if addBlockCmd.Parsed() {
		if *flagAddBlockArg == "" {
			PrintUsage()
			os.Exit(1)
		}
		cli.addBlock(*flagAddBlockArg)
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}
	if createChainWithGenesisBlockCmd.Parsed() {
		cli.createBlockChain()
	}
}
```

要补充实现几个函数

获取Chain对象，考虑到要获取已经保存下来的区块

```go
// GetBlockChainObject GetChainObject ChainObject 获取chain对象
func GetBlockChainObject() *Chain {
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Panicf("open the db [%s] failed %v\n", dbName, err)
	}
	var tip []byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	if err != nil {
		log.Panicf("get the blockchain object failed %v\n", err)
	}
	return &Chain{db, tip}
}
```

判断数据库是否存在:

```go
func isDBExists() bool {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}
	return true
}
```

并且进行判断:

```go
// CreateBlockChainWithGenesisBlock 创建区块链
func CreateBlockChainWithGenesisBlock() *Chain {
	if isDBExists() {
		fmt.Println("GenesisBlock Exists")
		os.Exit(1)
	}
    .....
}
```

运行

不带参数时输出信息:

```powershell
$ ./main
Usage: 
        createblockchain -- create blockchain
        addblock -- add block
        printchain -- print blockchain
```

输出:

```powershell
$ ./main printchain
Blockchain complete information...
--------------------------------------------
Hash:00004583df454a99f07799773a4f5ddef4e51c7001e9eb23c39b1cc9c6a138be
PrevBlockHash:
TimeStamp:1663981560
Data:[105 110 105 116 32 98 108 111 99 107 99 104 97 105 110]
Heigth:1
Nonce: 19299
--------------------------------------------
$ ./main addblock Alice sent 100 btc to Bob
```

## 交易管理

比特币交易原理，从以下几个方面考虑

1. 传统的web交易
2. 基本概念
3. 交易组成
4. UTXO交易模型
5. 交易过程

比特币系统中的交易没有余额这个概念，使用的是UTXO交易模型，在传统交易过程中所说的交易余额实际上指的是一个比特币钱包地址的UTXO集合

交易主要由输入，输入，ID(Hash值)，交易时间

UTXO交易模型是比特币专有的交易模型，在比特币的交易过程就是不断查找指定钱包地址的UTXO集合，然后进行修改的过程

UTXO是比特交易中最基本的单元，是不可拆分的，可以将其理解为一个币，该币拥有一个金额

交易可分为coinbase和普通转账，coinbase是挖矿奖励的比特币，没有发送者，由系统提供。普通转账就是正常的转账交易，有发送者参与，所以包含input

### 交易输入输出

添加一个Transaction结构

```go
// Transaction 交易管理文件
type Transaction struct {
	// 交易哈希
	TxHash []byte
}
```

修改区块结构，将Data修改为Txs

```go
// Block 区块基本结构和功能管理
type Block struct {
	TimeStamp     int64          // 区块时间戳
	Hash          []byte         // 哈希值
	PrevBlockHash []byte         // 前区块哈希
	Height        int64          // 区块高度
	Txs           []*Transaction // 交易数据
	Nonce         int64          // POW哈希变化值
}
```

添加一个将交易序列化的函数:

```go
// HashTransaction 将指定区块交易结构序列化
func (b *Block) HashTransaction() []byte {
	var txHashes [][]byte
	// 将指定区块中所有交易哈希进行拼接
	for _, tx := range b.Txs {
		txHashes = append(txHashes, tx.TxHash)
	}
	txHash := sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}
```

对应的，涉及到的函数都要做改动

### 实现Coinbase

一笔交易有输入和输出，如下定义相关的结构

```go
// TxInput 交易输入结构
type TxInput struct {
	TxHash    []byte // 交易哈希
	Vout      int    // 索引 引用上一笔交易的输出索引号
	ScriptSig string // 用户名
}

// TxOutput 交易输出
type TxOutput struct {
	Value        int    // 金额
	ScriptPubkey string // 用户名
}
```

Coinbase就是挖矿生成的转账，为系统奖励

```go
// NewCoinbaseTransaction 创建Coinbase交易
func NewCoinbaseTransaction(address string) *Transaction {
	var txCoinbase *Transaction
	txInput := &TxInput{
		[]byte{}, -1, "system reward",
	}
	txOutput := &TxOutput{10, address} // 挖矿奖励
	// 组装奖励
	txCoinbase = &Transaction{
		nil, []*TxInput{txInput}, []*TxOutput{txOutput},
	}
	txCoinbase.HashTransaction()
	return txCoinbase
}
```

生成一笔交易的哈希值

```go
// HashTransaction 生成交易哈希
func (tx *Transaction) HashTransaction() {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(tx); err != nil {
		log.Panicf("tx Hash encoded failed %v\n", err)
	}
	// 生成哈希值
	hash := sha256.Sum256(result.Bytes())
	tx.TxHash = hash[:]
}
```

