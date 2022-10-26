package block

import (
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"math/big"
	"os"
	"strconv"
)

const (
	dbName         = "block.db" // 数据库名称
	blockTableName = "blocks"   // 表名称
	banner         = "-------------------------------------------------------" +
		"---------------------------------------------------------------------" +
		"----------------------------------------"
)

// Chain 区块链的基本结构
type Chain struct {
	DB  *bolt.DB // 数据库连接
	Tip []byte   // 最新区块的哈希
}

func IsDBExists() bool {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}
	return true
}

// GetBlockChainObject 获取chain对象
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

// CreateBlockChainWithGenesisBlock 创建区块链
func CreateBlockChainWithGenesisBlock(address string) *Chain {
	if IsDBExists() {
		fmt.Println("GenesisBlock Exists")
		os.Exit(1)
	}

	var blockHash []byte
	db, err := bolt.Open(dbName, 0600, nil) // 创建bolt数据库
	if err != nil {
		log.Panicf("create db [%s] failed %v\n", blockTableName, err)
	}
	defer db.Close()
	// 创建桶 将生成的创世区块存入数据库
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName)) // 获取指定桶
		if b == nil {
			// 如果不存在则创建桶
			bucket, e := tx.CreateBucket([]byte(blockTableName))
			if e != nil {
				log.Panic(e)
			}
			txCoinbase := NewCoinbaseTransaction(address)                  // 生成一个coinbase交易
			genesisBlock := CreateGenesisBlock([]*Transaction{txCoinbase}) // 生成创世区块
			// key-哈希  value-序列化区块数据
			e = bucket.Put(genesisBlock.Hash, genesisBlock.Serialize()) // 储存数据
			if e != nil {
				log.Panicf("create bucket [%s] failed %v\n", blockTableName, e)
			}
			blockHash = genesisBlock.Hash
			// 存储最新区块的哈希
			e = bucket.Put([]byte("l"), genesisBlock.Hash)
			if e != nil {
				log.Panicf("put latest block error: %v", e)
			}
		}
		return nil
	})
	if err != nil {
		log.Panicf("boltdb update error: %v", err)
	}
	return &Chain{DB: db, Tip: blockHash}
}

// AddBlock 添加区块到区块链
func (c *Chain) AddBlock(txs []*Transaction) {
	// 从数据库获取到链上最新区块 反序列化后获取其哈希值
	// 基于当前最新区块和交易数据生成新区块哈希 插入数据库
	err := c.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName)) // 获取数据库桶
		if b != nil {
			blockBytes := b.Get(c.Tip) // 获取最新区块
			latestBlock := DeserializeBlock(blockBytes)
			newBlock := NewBlock(latestBlock.Height+1, latestBlock.Hash, txs)
			e := b.Put(newBlock.Hash, newBlock.Serialize()) // 将新生成区块插入数据库
			if e != nil {
				log.Panicf("put new block error: %v", e)
			}
			e = b.Put([]byte("l"), newBlock.Hash) // 设置最新区块哈希
			if e != nil {
				log.Panicf("put latest block error: %v", e)
			}
			c.Tip = newBlock.Hash
		}
		return nil
	})
	if err != nil {
		log.Panicf("update database error:%v", err)
	}
}

// PrintChain 遍历数据库 输出所有区块信息
func (c *Chain) PrintChain() {
	if !IsDBExists() {
		fmt.Println("blockchain has not been created")
		os.Exit(1)
	}
	// 读取数据库
	fmt.Println("Blockchain complete information...")
	var curBlock *Block
	iter := c.NewIterator() // 创建迭代器对象
	// 从最新区块开始循环读取
	for {
		fmt.Println(banner)
		// 从数据库读取数据
		curBlock = iter.Next()
		curBlock.PrintBlock()
		fmt.Println(banner)

		// 遍历到创世区块则选择退出
		var hashInt big.Int
		hashInt.SetBytes(curBlock.PrevBlockHash)
		// 比较相等 则遍历到创世区块
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
}

// MineNewBlock 挖掘一个新区块
func (c *Chain) MineNewBlock(from, to, amount []string) {
	var txs []*Transaction
	var block *Block

	value, _ := strconv.Atoi(amount[0])
	tx := NewSimpleTransaction(from[0], to[0], value) // 生成普通交易
	txs = append(txs, tx)

	// 从数据库中获取最新区块
	err := c.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			// 获取最新区块哈希值
			hash := b.Get([]byte("l"))
			blockBytes := b.Get(hash)
			block = DeserializeBlock(blockBytes) // 反序列化
		}
		return nil
	})
	if err != nil {
		log.Panicf("view database error:%v", err)
	}
	block = NewBlock(block.Height+1, block.Hash, txs)
	// 持久化新生成的区块 插入数据库
	c.AddBlock(txs)
}

// SpentOutputs 获取该地址已花费输出
func (c *Chain) SpentOutputs(address string) map[string][]int {
	// 遍历区块链中区块 查找与地址匹配的交易输入
	// 将对应交易输出添加进map 最终得到地址下的已花费输出
	spentTXOutputs := make(map[string][]int)
	it := c.NewIterator()
	for {
		block := it.Next()
		for _, tx := range block.Txs {
			// 排除coinbase交易
			if tx.IsCoinbaseTransaction() {
				continue
			}
			for _, vin := range tx.Vins {
				if vin.CheckPubkeyWithAddress(address) {
					key := hex.EncodeToString(vin.TxHash)
					// 将索引号添加到哈希值对应的索引列表中
					spentTXOutputs[key] = append(spentTXOutputs[key], vin.Vout)
				}
			}
		}

		// 遍历到创世区块则终止循环
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return spentTXOutputs
}

// UnUTXOS 查找指定地址的UTXO
func (c *Chain) UnUTXOS(address string) []*UTXO {
	// 遍历区块链数据库中区块 查找所有与address匹配的交易输出
	// 待查找的交易满足条件: 1.属于传入的地址 2.未被花费
	spentTXOutputs := c.SpentOutputs(address) // 获取地址对应的已花费输出
	it := c.NewIterator()
	var unUTXOS []*UTXO // 未花费输出列表
	for {
		block := it.Next()
		// 遍历区块中的每笔交易
		for _, tx := range block.Txs {
		LOOP:
			for index, vout := range tx.Vouts {
				// index: 当前输出在交易中索引
				// vout: 当前交易输出
				if vout.CheckPubkeyWithAddress(address) {
					// 首先判断是否存在已花费输出
					// 若存在 则要忽略已花费的输出
					if len(spentTXOutputs) != 0 {
						var spent bool // 标志当前交易输出是否被引用
						for txHash, indexArray := range spentTXOutputs {
							// txHash: 当前输出所引用的交易哈希
							// indexArray: 哈希关联的vout索引列表
							for _, i := range indexArray {
								// 输出索引相等 则当前输出被引用
								if txHash == hex.EncodeToString(tx.TxHash) && index == i {
									spent = true
									continue LOOP // 直接跳转到下一交易输出
								}
							}
						}
						// 检查标志 未被引用则添加到结果
						if spent == false {
							utxo := &UTXO{tx.TxHash, index, vout}
							unUTXOS = append(unUTXOS, utxo)
						}
					} else {
						// 已花费输出为空 将当前地址所有输出添加到结果
						utxo := &UTXO{tx.TxHash, index, vout}
						unUTXOS = append(unUTXOS, utxo)
					}
				}
			}
		}
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return unUTXOS
}

// GetBalance 获取账户余额
func (c *Chain) GetBalance(address string) int {
	var amount int
	utxos := c.UnUTXOS(address)

	for _, utxo := range utxos {
		amount += utxo.Output.Value
	}
	return amount
}
