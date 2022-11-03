package block

import (
	"blockchain/utils"
	"blockchain/wallet"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"time"
)

// 交易实际上解锁指定地址的output
// 然后重新分配它们的值 加锁到新的output中
// 为了交易安全 必须加密(签名)的数据:
// 1. 保存在已解锁output中的公钥哈希
// 2. 保存在新生成output中的公钥哈希
// 3. 新生成output所包含的value

// 一笔交易有一个哈希值唯一标识
// 一笔交易可能有多个输入或输出
// 一笔交易的输出会成为另一笔交易的输入
// 多笔转账交易可以同时被包含在一个区块中

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
		[]byte{}, -1, nil, nil,
	}
	txOutput := NewTxOutput(10, address) // 挖矿奖励
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

	// 添加时间戳标识
	stamp := time.Now().UnixNano()
	txHashBytes := bytes.Join([][]byte{result.Bytes(), utils.IntToHex(stamp)}, []byte{})
	// 生成哈希值
	hash := sha256.Sum256(txHashBytes)
	tx.TxHash = hash[:]
}

// NewSimpleTransaction 普通转账交易
func NewSimpleTransaction(from string, to string, amount int,
	chain *Chain, txs []*Transaction, nodeId string) *Transaction {
	var txInputs []*TxInput   // 输入列表
	var txOutputs []*TxOutput // 输出列表

	// 查找转账者的UTXO
	money, spendableUTXO := chain.FindSpendableUTXO(from, amount, txs)
	fmt.Printf("money: %v\n", money)

	// 获取钱包集合对象
	wallets := wallet.NewWallets(nodeId)
	w := wallets.Wallets[from] // 获取转账者钱包对象

	// 遍历spendableUTXO中的交易输出
	for txHash, indexArray := range spendableUTXO {
		txHashesBytes, err := hex.DecodeString(txHash)
		if err != nil {
			log.Panicf("decode string to bytes failed %v\n", err)
		}
		// 遍历索引列表 将交易输出索引拼接到当前交易输入列表中
		for _, index := range indexArray {
			txInput := &TxInput{txHashesBytes, index, nil, w.PublicKey}
			txInputs = append(txInputs, txInput)
		}
	}

	// 生成一笔交易输出 所属于转账目标
	txOutput := NewTxOutput(amount, to)
	txOutputs = append(txOutputs, txOutput)

	// 找零会生成一笔交易输出 所属于转账者
	if money > amount {
		txOutput = NewTxOutput(money-amount, from)
		txOutputs = append(txOutputs, txOutput)
	} else {
		log.Panicf("insufficient balance...")
	}

	tx := Transaction{nil, txInputs, txOutputs}
	tx.HashTransaction()

	// 使用私钥对交易进行签名
	chain.SignTransaction(&tx, w.PrivateKey)

	return &tx
}

// IsCoinbaseTransaction 判断指定的交易是否为coinbase交易
func (tx *Transaction) IsCoinbaseTransaction() bool {
	return tx.Vins[0].Vout == -1 && len(tx.Vins[0].TxHash) == 0
}

// Sign 对交易进行签名
func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTxs map[string]Transaction) {
	// 检查tx中每一个输入所引用的交易哈希是否包含在prevTxs中
	// 如果没有包含 表明该交易被人篡改
	for _, vin := range tx.Vins {
		if prevTxs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
			log.Panicf("previous transaction is not correct\n")
		}
	}

	// 提取要签名的属性
	txCopy := tx.TrimmedCopy()
	for id, vin := range txCopy.Vins {
		prevTx := prevTxs[hex.EncodeToString(vin.TxHash)]
		txCopy.Vins[id].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash
		txCopy.TxHash = txCopy.Hash()
		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.TxHash) // 调用核心函数
		if err != nil {
			log.Panicf("sign to transaction[%x] failed: %v", tx.TxHash, err)
		}
		sign := append(r.Bytes(), s.Bytes()...)
		tx.Vins[id].Signature = sign // 将签名赋值
	}
}

// TrimmedCopy 拷贝交易 生成一个专门用于交易签名的副本
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []*TxInput
	var outputs []*TxOutput

	for _, vin := range tx.Vins {
		inputs = append(inputs, &TxInput{vin.TxHash, vin.Vout, nil, nil})
	}
	for _, vout := range tx.Vouts {
		outputs = append(outputs, &TxOutput{vout.Value, vout.Ripemd160Hash})
	}
	txCopy := Transaction{tx.TxHash, inputs, outputs}
	return txCopy
}

// Hash 生成交易哈希
func (tx *Transaction) Hash() []byte {
	txCopy := tx
	txCopy.TxHash = []byte{}
	hash := sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

// Serialize 交易序列化
func (tx *Transaction) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(tx); err != nil {
		log.Panicf("serialize the transaction to byte failed: %v\n", err)
	}
	return result.Bytes()
}

// Verify 验证交易
func (tx *Transaction) Verify(prevTxs map[string]Transaction) bool {
	// 检查能否查找到交易哈希
	for _, vin := range tx.Vins {
		if prevTxs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
			log.Panicf("verify transaction failed\n")
		}
	}
	// 提取相同交易签名
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	// 遍历交易输入 对每笔输入所引用的输出进行验证
	for id, vin := range tx.Vins {
		prevTx := prevTxs[hex.EncodeToString(vin.TxHash)]
		txCopy.Vins[id].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash
		// 由要验证的数据生成的交易哈希 须与签名完全一致
		txCopy.TxHash = txCopy.Hash()
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:sigLen/2])
		s.SetBytes(vin.Signature[sigLen/2:])

		x := big.Int{}
		y := big.Int{}
		pubKeyLen := len(vin.PublicKey)
		x.SetBytes(vin.PublicKey[:pubKeyLen/2])
		y.SetBytes(vin.PublicKey[pubKeyLen/2:])
		rawPublicKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPublicKey, txCopy.TxHash, &r, &s) {
			return false
		}
	}
	return true
}
