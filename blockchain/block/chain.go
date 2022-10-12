package block

import (
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
	// 创建或打开数据库
	db, err := bolt.Open(dbName, 0600, nil)
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

// AddBlock 添加区块到区块链
func (c *Chain) AddBlock(txs []*Transaction) {
	// 更新区块数据
	err := c.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName)) // 获取数据库桶
		if b != nil {
			blockBytes := b.Get(c.Tip) // 获取最新区块
			latestBlock := DeserializeBlock(blockBytes)
			newBlock := NewBlock(latestBlock.Height+1, latestBlock.Hash, txs) // 创建区块
			e := b.Put(newBlock.Hash, newBlock.Serialize())                   // 存储最新区块
			if e != nil {
				log.Panicf("put new block error:%v", e)
			}
			e = b.Put([]byte("l"), newBlock.Hash) // 设置最新区块哈希
			if e != nil {
				log.Panicf("put latest block error:%v", e)
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
	// 读取数据库
	fmt.Println("Blockchain complete information...")
	var curBlock *Block
	iter := c.Iterator() // 获取迭代器对象
	// 从最新区块开始循环读取
	for {
		fmt.Println("------------------------------------------------------------------")
		// 从数据库读取数据
		curBlock = iter.Next()
		curBlock.PrintBlock()
		fmt.Println("------------------------------------------------------------------")

		// 遍历到创世区块则选择退出
		var hashInt big.Int
		hashInt.SetBytes(curBlock.PrevBlockHash)
		// 比较相等 则遍历到创世区块
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
}

// MineNewBlock 挖矿功能
func (c *Chain) MineNewBlock(from, to, amount []string) {
	var txs []*Transaction
	var block *Block

	value, _ := strconv.Atoi(amount[0])
	tx := NewSimpleTransaction(from[0], to[0], value)
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
	// 持久化新生成的区块
	err = c.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			e := b.Put(block.Hash, block.Serialize())
			if e != nil {
				log.Panicf("update the new block to db failed %v\n", e)
			}
			e = b.Put([]byte("l"), block.Hash)
			if e != nil {
				log.Panicf("update the new block hash to db failed %v\n", e)
			}
			c.Tip = block.Hash
		}
		return nil
	})
	if err != nil {
		log.Panicf("update database error:%v", err)
	}
}

// UnUTXOS 遍历查找区块链中每一个区块的每一个交易
func (c *Chain) UnUTXOS(address string) []*TxOutput {
	// 遍历数据库 查找所有与address相关的交易
	it := c.Iterator() // 迭代器对象
	for {
		block := it.Next()
		// 遍历区块中的每笔交易
		for _, tx := range block.Txs {
			for _, vout := range tx.Vouts {
				if vout.CheckPubkeyWithAddress(address) {

				}
			}
		}
	}
}
