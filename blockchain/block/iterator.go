package block

import (
	"github.com/boltdb/bolt"
	"log"
)

// ChainIterator 区块链迭代器
type ChainIterator struct {
	DB          *bolt.DB // 迭代目标
	CurrentPath []byte   // 当前迭代目标的哈希
}

// NewIterator 创建迭代器 用于遍历
func (c *Chain) NewIterator() *ChainIterator {
	return &ChainIterator{
		c.DB, c.Tip,
	}
}

// Next 指向下一个区块
func (it *ChainIterator) Next() *Block {
	var block *Block
	err := it.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			currentBlockBytes := b.Get(it.CurrentPath)
			// 更新迭代器中区块的哈希值
			block = DeserializeBlock(currentBlockBytes)
			it.CurrentPath = block.PrevBlockHash // 获取到前区块哈希
		}
		return nil
	})
	if err != nil {
		log.Panicf("iterator the db failed: %v\n", err)
	}

	return block
}
