package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
)

// 一笔交易有一个哈希值唯一标识
// 一笔交易可能有多个输入或输出
// 一笔交易的输出会成为另一笔交易的输入

// Transaction 交易结构
type Transaction struct {
	TxHash []byte      // 交易哈希
	Vins   []*TxInput  // 输入列表
	Vouts  []*TxOutput // 输出列表
}

// NewCoinbaseTransaction 创建Coinbase交易
// Coinbase Transaction: 币基交易
// 每个区块的第一笔交易 由挖矿奖励产生 比特币由此在挖矿中被创造
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

// NewSimpleTransaction 生成普通转账交易
func NewSimpleTransaction(from string, to string, amount int) *Transaction {
	var txInputs []*TxInput   // 输入列表
	var txOutputs []*TxOutput // 输出列表

	// 输人
	txInput := &TxInput{
		[]byte("c9efa576aac56bf0920a840b0ec6dacfdae7083809b40df46711b19bc29392fb"), 0, from}
	txInputs = append(txInputs, txInput)

	txOutput := &TxOutput{10 - amount, from}
	txOutputs = append(txOutputs, txOutput)

	// 输出(找零)
	if amount < 10 {
		txOutput = &TxOutput{10 - amount, from}
		txOutputs = append(txOutputs, txOutput)
	}
	tx := Transaction{nil, txInputs, txOutputs}
	tx.HashTransaction()
	return &tx
}

// IsCoinbaseTransaction 判断指定的交易是否为coinbase交易
func (tx *Transaction) IsCoinbaseTransaction() bool {
	return tx.Vins[0].Vout == -1 && len(tx.Vins[0].TxHash) == 0
}
