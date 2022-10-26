package block

// UTXO： 未话费的交易输出
// 每一个交易代表UTXO集的变化 用户比特币余额指用户钱包中可用UTXO总和
// UTXO具有不可分割性 比特币交易必须从用户可用的UTXO中创建出来
// 一笔交易会消耗先前已经被记录的UTXO 同时创建新的UTXO以供未来被消耗

// UTXO查找三要素：
// 1. 输入引用的交易哈希
// 2. 输出索引
// 3. 交易输出

// UTXO 结构
type UTXO struct {
	TxHash []byte    // UTXO对应哈希
	Index  int       // 所属交易输出列表中索引
	Output *TxOutput // 交易输出
}
