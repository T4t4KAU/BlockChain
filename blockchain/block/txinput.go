package block

// 交易输入将UTXO(通过引用)标记为被消费 通过解锁脚本提供所有权证明
// 要构建一个交易 一个钱包从它控制的UTXO中选择足够的价值来执行被请求的付款
// 对于用于付款的每个UTXO 钱包将创建一个指向UTXO的输入

// TxInput 交易输入结构
type TxInput struct {
	TxHash    []byte // 交易哈希
	Vout      int    // 输出索引 引用上一笔交易的输出索引号
	ScriptSig string
}

// CheckPubkeyWithAddress 验证引用的地址是否匹配
func (txInput *TxInput) CheckPubkeyWithAddress(address string) bool {
	return address == txInput.ScriptSig
}
