package block

// 实现区块的基本结构

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

// Block 区块基本结构和功能管理
type Block struct {
	TimeStamp     int64          // 区块时间戳
	Hash          []byte         // 哈希值
	PrevBlockHash []byte         // 前区块哈希
	Height        int64          // 区块高度
	Txs           []*Transaction // 交易数据
	Nonce         int64          // POW哈希变化值
}

// NewBlock 创建一个区块
func NewBlock(height int64, prevBlockHash []byte, txs []*Transaction) *Block {
	var block Block
	// 生成一个区块
	block = Block{
		TimeStamp:     time.Now().Unix(),
		Hash:          nil,
		PrevBlockHash: prevBlockHash,
		Height:        height,
		Txs:           txs,
	}

	// 生成哈希
	block.SetHash()
	pow := NewProofOfWork(&block)
	hash, nonce := pow.Run()
	block.Hash = hash
	block.Nonce = int64(nonce)
	return &block
}

// SetHash 计算区块哈希
func (b *Block) SetHash() {
	timeStampBytes := IntToHex(b.TimeStamp)
	heightBytes := IntToHex(b.Height)
	// 调用SHA256实现哈希生成
	blockBytes := bytes.Join([][]byte{
		heightBytes, timeStampBytes,
		b.PrevBlockHash, b.HashTransaction(),
	}, []byte{})
	hash := sha256.Sum256(blockBytes)
	b.Hash = hash[:]
}

// CreateGenesisBlock 生成创世区块
func CreateGenesisBlock(txs []*Transaction) *Block {
	return NewBlock(1, nil, txs)
}

// Serialize 区块结构序列化
func (b *Block) Serialize() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer) // 创建编码对象
	if err := encoder.Encode(b); err != nil {
		log.Panicf("encode block error:%v", err)

	}
	return buffer.Bytes()
}

// DeserializeBlock 区块数据反序列化
func DeserializeBlock(blockBytes []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes)) // 创建解码对象
	if err := decoder.Decode(&block); err != nil {
		log.Panicf("decode block error:%v", err)
	}
	return &block
}

// PrintBlock 打印区块信息
func (b *Block) PrintBlock() {
	// 输出区块详情
	fmt.Printf("Hash:%x\n", b.Hash)
	fmt.Printf("PrevBlockHash:%x\n", b.PrevBlockHash)
	fmt.Printf("TimeStamp:%v\n", b.TimeStamp)
	fmt.Printf("Height:%v\n", b.Height)
	fmt.Printf("Nonce: %v\n", b.Nonce)

	for _, tx := range b.Txs {
		fmt.Printf("\ttx-hash: %x\n", tx.TxHash)
		fmt.Printf("\tinput...\n")
		for _, vin := range tx.Vins {
			fmt.Printf("\t\tvin-txHash: %x\n", vin.TxHash)
			fmt.Printf("\t\tvin-vout: %v\n", vin.Vout)
			fmt.Printf("\t\tvin-scriptSig: %s\n", vin.ScriptSig)
		}
		fmt.Printf("\toutput...\n")
		for _, vout := range tx.Vouts {
			fmt.Printf("\t\tvout-value: %d\n", vout.Value)
			fmt.Printf("\t\tvout-scriptPubkey: %s\n", vout.ScriptPubkey)
		}
	}
}

// HashTransaction 将指定区块交易结构序列化
func (b *Block) HashTransaction() []byte {
	var txHashes [][]byte
	// 将指定区块中所有交易哈希进行拼接
	for _, tx := range b.Txs {
		txHashes = append(txHashes, tx.TxHash)
	}
	txHash := sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}
