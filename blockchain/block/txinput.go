package block

import (
	"blockchain/crypto"
	"bytes"
)

// 交易输入将UTXO(通过引用)标记为被消费 通过解锁脚本提供所有权证明
// 要构建一个交易 一个钱包从它控制的UTXO中选择足够的价值来执行被请求的付款
// 对于用于付款的每个UTXO 钱包将创建一个指向UTXO的输入

// TxInput 交易输入结构
type TxInput struct {
	TxHash    []byte // 交易哈希
	Vout      int    // 输出索引 引用上一笔交易的输出索引号
	Signature []byte // 数字签名
	PublicKey []byte // 公钥
}

// UnLockRipemd160Hash 解锁账户
func (txInput *TxInput) UnLockRipemd160Hash(ripemd160Hash []byte) bool {
	inputRipemd160Hash := crypto.Ripemd160Hash(txInput.PublicKey)
	return bytes.Compare(inputRipemd160Hash, ripemd160Hash) == 0
}
