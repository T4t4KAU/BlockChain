package block

// 每一笔比特币交易会创造输出 并被比特币账簿记录(除特例)
// 几乎所有输出都能创造出一定数量的可用于支付的BTC即UTXO
// 产生的UTXO被整个网络识别 拥有者可在未来交易中使用
// UTXO在UTXO集中被每一个全节点BTC客户端追踪 新的交易从UTXO集中消耗一个或多个输出

// TxOutput 交易输出
type TxOutput struct {
	Value        int // 金额
	ScriptPubkey string
}

// CheckPubkeyWithAddress 验证当前UTXO是否属于指定的地址
func (txOutput *TxOutput) CheckPubkeyWithAddress(address string) bool {
	return address == txOutput.ScriptPubkey
}
