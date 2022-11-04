package block

import (
	"blockchain/crypto"
	"blockchain/wallet"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"math/big"
	"os"
	"strconv"
)

const (
	dbName         = "block-%s.db" // 数据库名称
	blockTableName = "blocks"      // 表名称
	banner         = "-------------------------------------------------------" +
		"---------------------------------------------------------------------" +
		"----------------------------------------"
)

// Chain 区块链的基本结构
type Chain struct {
	DB  *bolt.DB // 数据库连接
	Tip []byte   // 最新区块的哈希
}

func IsDBExists(nodeId string) bool {
	if _, err := os.Stat(fmt.Sprintf(dbName, nodeId)); os.IsNotExist(err) {
		return false
	}
	return true
}

// GetBlockChainObject 获取chain对象
func GetBlockChainObject(nodeId string) *Chain {
	if !IsDBExists(nodeId) {
		os.Exit(1)
	}

	name := fmt.Sprintf(dbName, nodeId)
	db, err := bolt.Open(name, 0600, nil)
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
func CreateBlockChainWithGenesisBlock(address string, nodeId string) *Chain {
	if IsDBExists(nodeId) {
		fmt.Println("GenesisBlock Exists")
		os.Exit(1)
	}

	var blockHash []byte
	name := fmt.Sprintf(dbName, nodeId)
	db, err := bolt.Open(name, 0600, nil) // 创建bolt数据库
	if err != nil {
		log.Panicf("create db [%s] failed %v\n", blockTableName, err)
	}
	// 创建桶 将生成的创世区块存入数据库
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName)) // 获取指定桶
		if b == nil {
			// 如果不存在则创建桶
			bucket, e := tx.CreateBucket([]byte(blockTableName))
			if e != nil {
				log.Panicf("create bucket[%s] failed: %v", blockTableName, err)
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
func (c *Chain) AddBlock(newBlock *Block) {
	// 从数据库获取到链上最新区块 反序列化后获取其哈希值
	// 基于当前最新区块和交易数据生成新区块哈希 插入数据库
	err := c.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName)) // 获取数据库桶
		if b != nil {
			if b.Get(newBlock.Hash) != nil {
				return nil
			}
			e := b.Put(newBlock.Hash, newBlock.Serialize()) // 将新生成区块插入数据库
			if e != nil {
				log.Panicf("put new block error: %v", e)
			}
			blockHash := b.Get([]byte("l"))
			latestBlock := b.Get(blockHash)
			rawBlock := DeserializeBlock(latestBlock)
			if rawBlock.Height < newBlock.Height {
				err := b.Put([]byte("l"), newBlock.Hash)
				if err != nil {
					log.Panicf("put latest block failed: %v", err)
				}
				c.Tip = newBlock.Hash
			}
		}
		return nil
	})
	if err != nil {
		log.Panicf("update database error:%v", err)
	}
	fmt.Println("the block is added")
}

// PrintChain 遍历数据库 输出所有区块信息
func (c *Chain) PrintChain(nodeId string) {
	if !IsDBExists(nodeId) {
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
		fmt.Printf(banner + "\n\n")

		// 遍历到创世区块则选择退出
		var hashInt big.Int
		hashInt.SetBytes(curBlock.PrevBlockHash)
		// 比较相等 则遍历到创世区块
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
}

func (c *Chain) GetLatestBlock() *Block {
	var block *Block
	// 从数据库中获取一个最新区块
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
	return block
}

// MineNewBlock 挖掘一个新区块
func (c *Chain) MineNewBlock(from, to, amount []string, nodeId string) {
	var txs []*Transaction
	var block *Block

	// 遍历交易参与者
	for index, address := range from {
		value, _ := strconv.Atoi(amount[index])
		tx := NewSimpleTransaction(address, to[index], value, c, txs, nodeId)
		txs = append(txs, tx) // 追加到交易列表

		// 给予交易发起者(矿工)奖励
		tx = NewCoinbaseTransaction(address)
		txs = append(txs, tx)
	}

	// 对txs中每笔交易进行验证
	for _, tx := range txs {
		// 验证签名 只要有一笔签名的验证失败
		c.VerifyTransaction(tx)
	}
	block = NewBlock(block.Height+1, block.Hash, txs)
	// 持久化新生成的区块 插入数据库
	c.AddBlock(block)
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
				if vin.UnLockRipemd160Hash(wallet.StringToHash160(address)) {
					key := hex.EncodeToString(vin.TxHash)
					// 将索引号添加到哈希值对应的索引列表中
					spentTXOutputs[key] = append(spentTXOutputs[key], vin.Vout)
				}
			}
		}

		// 遍历到创世区块则终止循环
		if isBreakLoop(block.PrevBlockHash) {
			break
		}
	}
	return spentTXOutputs
}

// UnUTXOS 查找指定地址的UTXO
// 先在交易缓存中查找 再在数据库中查找
func (c *Chain) UnUTXOS(address string, txs []*Transaction) []*UTXO {
	// 遍历区块链数据库中区块 查找所有与address匹配的交易输出
	// 待查找的交易满足条件: 1.属于传入的地址 2.未被花费
	var unUTXOS []*UTXO                       // 未花费输出列表
	spentTXOutputs := c.SpentOutputs(address) // 获取地址对应的已花费输出

	// 查找缓存中已花费输出
	for _, tx := range txs {
		// 忽略coinbase交易
		if !tx.IsCoinbaseTransaction() {
			for _, in := range tx.Vins {
				if in.UnLockRipemd160Hash(wallet.StringToHash160(address)) {
					key := hex.EncodeToString(in.TxHash)
					spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout) // 添加到已花费输出
				}
			}
		}
	}

	// 优先遍历缓存中交易
	for _, tx := range txs {
	CacheTx:
		for index, vout := range tx.Vouts {
			// 遍历交易中输出列表
			if vout.UnLockScriptPubkeyWithAddress(address) {
				if len(spentTXOutputs) != 0 {
					var isUtxoTx bool
					// 对于交易的每一个输出 检查在已花费输出中是否存在
					// 如果存在则说明该输出已经在其他交易中被引用
					// 如果已花费输出列表包含此交易的哈希 则说明输出已经被引用
					for txHash, indexArray := range spentTXOutputs {
						txHashStr := hex.EncodeToString(tx.TxHash)
						// 若交易哈希值相等 表明当前交易存在输出被其他交易输入引用
						if txHash == txHashStr {
							isUtxoTx = true
							var isSpentUTXO bool // 状态变量 判断指定输出是否被引用
							// 若索引号在列表中 说明已经该输出已经被引用 直接跳过该vout
							for _, voutIndex := range indexArray {
								if index == voutIndex {
									isSpentUTXO = true
									continue CacheTx
								}
							}
							// 输出未被引用 放入UTXO集
							if isSpentUTXO == false {
								utxo := &UTXO{tx.TxHash, index, vout}
								unUTXOS = append(unUTXOS, utxo)
							}
						}
					}
					if isUtxoTx == false {
						// 此交易不存在输出被引用 直接将输出添加到UTXO集合
						utxo := &UTXO{tx.TxHash, index, vout}
						unUTXOS = append(unUTXOS, utxo)
					}
				} else {
					// 该情况下 所有交易输出都属于UTXO
					utxo := &UTXO{tx.TxHash, index, vout}
					unUTXOS = append(unUTXOS, utxo)
				}
			}
		}
	}

	it := c.NewIterator()
	for {
		block := it.Next()
		// 遍历区块中的每笔交易
		for _, tx := range block.Txs {
		LOOP:
			for index, vout := range tx.Vouts {
				// index: 当前输出在交易中索引
				// vout: 当前交易输出
				if vout.UnLockScriptPubkeyWithAddress(address) {
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
		if isBreakLoop(block.PrevBlockHash) {
			break
		}
	}
	return unUTXOS
}

// FindAllSpentOutputs 查找链上所有已花费输出
func (c *Chain) FindAllSpentOutputs() map[string][]*TxInput {
	it := c.NewIterator()
	spentTXOutputs := make(map[string][]*TxInput)
	for {
		// 遍历链上所有区块中交易
		block := it.Next()
		for _, tx := range block.Txs {
			if !tx.IsCoinbaseTransaction() {
				for _, txInput := range tx.Vins {
					txHash := hex.EncodeToString(txInput.TxHash)
					spentTXOutputs[txHash] = append(spentTXOutputs[txHash], txInput)
				}
			}
		}
		if isBreakLoop(block.PrevBlockHash) {
			break
		}
	}
	return spentTXOutputs
}

// 判断是否遍历完整个区块链
func isBreakLoop(prevBlockHash []byte) bool {
	var hashInt big.Int
	hashInt.SetBytes(prevBlockHash)
	if hashInt.Cmp(big.NewInt(0)) == 0 {
		return true
	}
	return false
}

// FindUTXOMap 查找链上所有UTXO
func (c *Chain) FindUTXOMap() map[string]*TxOutputs {
	it := c.NewIterator()
	utxoMaps := make(map[string]*TxOutputs)
	spentTXOutputs := c.FindAllSpentOutputs() // 查找已花费输出

	for {
		block := it.Next()
		// 遍历区块中每一条交易
		for _, tx := range block.Txs {
			txOutputs := &TxOutputs{[]*TxOutput{}}
			txHash := hex.EncodeToString(tx.TxHash)
		LOOP:
			for index, vout := range tx.Vouts {
				// 获取指定交易输入
				txInputs := spentTXOutputs[txHash]
				if len(txInputs) > 0 {
					spent := false
					for _, in := range txInputs {
						outPubkey := vout.Ripemd160Hash
						inPubkey := in.PublicKey
						// 将input的公钥与交易输出的公钥哈希进行比对
						if bytes.Compare(outPubkey, crypto.Ripemd160Hash(inPubkey)) == 0 {
							if index == in.Vout {
								spent = true
								continue LOOP
							}
						}
					}
					if spent == false {
						txOutputs.Set = append(txOutputs.Set, vout)
					}
				} else {
					// 没有input引用该交易输出 代表当前交易所有输出为UTXO
					txOutputs.Set = append(txOutputs.Set, vout)
				}
			}
			utxoMaps[txHash] = txOutputs
		}
		if isBreakLoop(block.PrevBlockHash) {
			break
		}
	}
	return utxoMaps
}

// GetBalance 获取账户余额
func (c *Chain) GetBalance(address string) int {
	var amount int
	utxos := c.UnUTXOS(address, []*Transaction{})

	for _, utxo := range utxos {
		amount += utxo.Output.Value
	}
	return amount
}

func (c *Chain) FindTransaction(ID []byte) Transaction {
	it := c.NewIterator()
	for {
		block := it.Next()
		for _, tx := range block.Txs {
			if bytes.Compare(ID, tx.TxHash) == 0 {
				return *tx
			}
		}
		if isBreakLoop(block.PrevBlockHash) {
			break
		}
	}
	fmt.Printf("not found the tx[%x]\n", ID)
	return Transaction{}
}

// FindSpendableUTXO 查找指定地址的可用UTXO
func (c *Chain) FindSpendableUTXO(from string,
	amount int, txs []*Transaction) (int, map[string][]int) {
	// txs: 缓存交易列表 用于多笔交易处理
	spendableUTXO := make(map[string][]int)
	var value int
	utxos := c.UnUTXOS(from, txs)
	// 遍历UTXO
	for _, utxo := range utxos {
		value += utxo.Output.Value
		hash := hex.EncodeToString(utxo.TxHash)
		spendableUTXO[hash] = append(spendableUTXO[hash], utxo.Index)
		if value >= amount {
			break
		}
	}
	// UTXO金额总和小与amount
	if value < amount {
		fmt.Printf("insufficient balance of address[%s] "+
			"current balance[%d] transfer amount[%d]\n", from, value, amount)
		os.Exit(1)
	}
	return value, spendableUTXO
}

// ECDSA数字签名 有3种用途:
// 1.证明私钥的所有者已经授权支出这笔资金
// 2.授权证明不可否认
// 3.签名交易后 不能被任何人修改

// SignTransaction 交易签名
func (c *Chain) SignTransaction(tx *Transaction, priKey ecdsa.PrivateKey) {
	// 无须对coinbase签名
	if tx.IsCoinbaseTransaction() {
		return
	}
	prevTxs := make(map[string]Transaction)
	// 处理交易输入 查找交易中输入所引用的vout所属交易(查找发送者)
	// 对花费的每一笔UTXO进行签名
	for _, vin := range tx.Vins {
		// 查找当前交易输入所引用的交易
		foundTx := c.FindTransaction(vin.TxHash)
		prevTxs[hex.EncodeToString(foundTx.TxHash)] = foundTx
	}
	tx.Sign(priKey, prevTxs)
}

// VerifyTransaction 验证交易
func (c *Chain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbaseTransaction() {
		return true
	}
	prevTxs := make(map[string]Transaction)
	// 查找输入引用的交易
	for _, vin := range tx.Vins {
		foundTx := c.FindTransaction(vin.TxHash)
		prevTxs[hex.EncodeToString(foundTx.TxHash)] = foundTx
	}
	return tx.Verify(prevTxs)
}

// GetHeight 获取当前区块高度
func (c *Chain) GetHeight() int64 {
	return c.NewIterator().Next().Height
}

func (c *Chain) GetBlock(hash []byte) []byte {
	var blockByte []byte
	err := c.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			blockByte = b.Get(hash)
		}
		return nil
	})
	if err != nil {
		log.Panicf("view database failed: %v", err)
	}

	return blockByte
}

// GetBlockHashes 获取链上所有区块哈希
func (c *Chain) GetBlockHashes() [][]byte {
	var blockHashes [][]byte
	it := c.NewIterator() // 创建迭代器
	for {
		block := it.Next()
		blockHashes = append(blockHashes, block.Hash) // 添加到列表中
		if isBreakLoop(block.PrevBlockHash) {
			break
		}
	}
	return blockHashes
}
