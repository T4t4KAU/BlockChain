package block

// TxInput 交易输入结构
type TxInput struct {
	TxHash    []byte // 交易哈希
	Vout      int    // 索引 引用上一笔交易的输出索引号
	ScriptSig string // 用户名
}
