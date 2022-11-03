package block

import (
	"blockchain/wallet"
	"bytes"
	"encoding/gob"
	"log"
)

// 每一笔比特币交易会创造输出 并被比特币账簿记录(除特例)
// 几乎所有输出都能创造出一定数量的可用于支付的BTC即UTXO
// 产生的UTXO被整个网络识别 拥有者可在未来交易中使用
// UTXO在UTXO集中被每一个全节点BTC客户端追踪 新的交易从UTXO集中消耗一个或多个输出

// TxOutput 交易输出
type TxOutput struct {
	Value         int // 金额
	Ripemd160Hash []byte
}

type TxOutputs struct {
	Set []*TxOutput
}

// UnLockScriptPubkeyWithAddress 解锁账户
func (txOutput *TxOutput) UnLockScriptPubkeyWithAddress(address string) bool {
	hash160 := wallet.StringToHash160(address) // 将地址转换为哈希字节序列
	return bytes.Compare(hash160, txOutput.Ripemd160Hash) == 0
}

// NewTxOutput 创建交易输出
func NewTxOutput(value int, address string) *TxOutput {
	txOutput := &TxOutput{}
	hash160 := wallet.StringToHash160(address)
	txOutput.Value = value
	txOutput.Ripemd160Hash = hash160
	return txOutput
}

// Serialize 交易输出序列化
func (set *TxOutputs) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(set); err != nil {
		log.Panicf("serialize the utxo failed: %v\n", err)
	}
	return result.Bytes()
}

func DeserializeTxOutputs(txOutputsBytes []byte) *TxOutputs {
	var txOutputs TxOutputs
	decoder := gob.NewDecoder(bytes.NewReader(txOutputsBytes))
	if err := decoder.Decode(&txOutputs); err != nil {
		log.Panicf("deserialize the struct utxo failed: %v\n", err)
	}
	return &txOutputs
}
