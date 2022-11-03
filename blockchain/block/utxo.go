package block

import (
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

// UTXO： 未花费的交易输出
// 每一个交易代表UTXO集的变化 用户比特币余额指用户钱包中可用UTXO总和
// UTXO具有不可分割性 比特币交易必须从用户可用的UTXO中创建出来
// 一笔交易会消耗先前已经被记录的UTXO 同时创建新的UTXO以供未来被消耗

// UTXO查找三要素：
// 1. 输入引用的交易哈希
// 2. 输出索引
// 3. 交易输出

const utxoTableName = "utxoTable" // UTXO表

// UTXO 结构
type UTXO struct {
	TxHash []byte    // UTXO对应哈希
	Index  int       // 所属交易输出列表中索引
	Output *TxOutput // 交易输出
}

// UTXOSet UTXO集合结构
type UTXOSet struct {
	Chain *Chain
}

// UpdateUTXOSet 更新UTXO集
func (s *UTXOSet) UpdateUTXOSet() {
	// 获取最新区块
	latestBlock := s.Chain.NewIterator().Next()
	err := s.Chain.DB.Update(func(tx *bolt.Tx) error {
		// 将最新区块中的UTXO插入
		b := tx.Bucket([]byte(utxoTableName))
		if b != nil {
			for _, t := range latestBlock.Txs {
				if !t.IsCoinbaseTransaction() {
					for _, vin := range t.Vins {
						updatedOutputs := TxOutputs{}
						outputBytes := b.Get(vin.TxHash)
						outs := DeserializeTxOutputs(outputBytes)
						for outIdx, out := range outs.Set {
							if vin.Vout != outIdx {
								updatedOutputs.Set = append(updatedOutputs.Set, out)
							}
						}
						if len(updatedOutputs.Set) == 0 {
							err := b.Delete(vin.TxHash)
							if err != nil {
								log.Panicf("delete tx failed: %v\n", err)
							}
						} else {
							// 将更新后的UTXO数据存入数据库
							err := b.Put(vin.TxHash, updatedOutputs.Serialize())
							if err != nil {
								log.Panicf("put tx failed: %v\n", err)
							}
						}
					}
				}
				newOutputs := TxOutputs{}
				newOutputs.Set = append(newOutputs.Set, t.Vouts...)
				err := b.Put(t.TxHash, newOutputs.Serialize())
				if err != nil {
					log.Panicf("put txhash failed: %v\n", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panicf("update UTXOs failed: %v\n", err)
	}
}

// ResetUTXOSet 重置UTXO集合
func (s *UTXOSet) ResetUTXOSet() {
	// 首次创建时 创建UTXO
	err := s.Chain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoTableName))
		if b != nil {
			err := tx.DeleteBucket([]byte(utxoTableName))
			if err != nil {
				log.Panicf("delete the utxo table failed: %v\n", err)
			}
		}
		bucket, err := tx.CreateBucket([]byte(utxoTableName))
		if err != nil {
			log.Panicf("delete the utxo table failed: %v\n", err)
		}
		if bucket != nil {
			txOutputMap := s.Chain.FindUTXOMap() // 查找当前所有UTXO
			for keyHash, outputs := range txOutputMap {
				// 将所有UTXO存储
				txHash, _ := hex.DecodeString(keyHash)
				fmt.Printf("txHash: %x\n", txHash)
				// 存入UTXO table
				err = bucket.Put(txHash, outputs.Serialize())
				if err != nil {
					log.Panicf("put utxo into table failed: %v\n", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panicf("update database failed: %v\n", err)
	}
}

func (s *UTXOSet) FindUTXOWithAddress(address string) []*UTXO {
	var utxos []*UTXO
	err := s.Chain.DB.View(func(tx *bolt.Tx) error {
		// 获取UTXO table
		b := tx.Bucket([]byte(utxoTableName))
		if b != nil {
			c := b.Cursor()
			// 通过游标遍历bolt数据库中数据
			for k, v := c.First(); k != nil; k, v = c.Next() {
				txOutputs := DeserializeTxOutputs(v)
				for _, utxo := range txOutputs.Set {
					if utxo.UnLockScriptPubkeyWithAddress(address) {
						singleUTXO := UTXO{Output: utxo}
						utxos = append(utxos, &singleUTXO)
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Panicf("find the utxo of [%s] failed: %v\n", address, err)
	}
	return utxos
}

func (s *UTXOSet) GetBalance(address string) int {
	UTXOS := s.FindUTXOWithAddress(address)
	var amount int
	for _, utxo := range UTXOS {
		fmt.Printf("utxo-txhash: %x\n", utxo.TxHash)
		fmt.Printf("utxo-index: %x\n", utxo.Index)
		fmt.Printf("utxo-Ripemd160Hash: %x\n", utxo.Output.Ripemd160Hash)
		fmt.Printf("utxo-value: %x\n", utxo.Output.Value)
		amount += utxo.Output.Value
	}
	return amount
}
