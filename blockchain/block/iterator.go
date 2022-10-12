package block

import (
	"github.com/boltdb/bolt"
	"log"
)

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
