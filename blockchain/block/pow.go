package block

import (
	"blockchain/utils"
	"bytes"
	"crypto/sha256"
	"math/big"
)

// 共识算法管理 POW实例
// POW: 工作量证明

const targetBit = 16 // 目标难度值

// ProofOfWork POW结构
type ProofOfWork struct {
	Block  *Block   // 要共识验证的区块
	Target *big.Int // 目标难度的哈希(大数存储)
}

// NewProofOfWork 创建一个POW对象
func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	// 数据总长度为8 满足前两位为0
	// 1 * 2 << (8-2) = 64
	target = target.Lsh(target, 256-targetBit)
	return &ProofOfWork{Block: block, Target: target}
}

// Run 计算哈希值
func (p *ProofOfWork) Run() ([]byte, int) {
	var nonce = 0 // 碰撞次数
	var hashInt big.Int
	var hash [32]byte
	// 反复计算哈希值 直到符合要求
	for {
		// 生成准备数据
		dataBytes := p.prepareData(int64(nonce))
		hash = sha256.Sum256(dataBytes) // 计算SHA256值
		hashInt.SetBytes(hash[:])
		// 检测生成的哈希值是否符合条件
		if p.Target.Cmp(&hashInt) == 1 {
			// 找到符合条件的哈希值
			break
		}
		nonce++
	}
	return hash[:], nonce
}

// 生成准备数据
func (p *ProofOfWork) prepareData(nonce int64) []byte {
	// 拼接区块属性
	timeStampBytes := utils.IntToHex(p.Block.TimeStamp)
	heightBytes := utils.IntToHex(p.Block.Height)
	// 拼接区块数据 用于后续做哈希计算
	// 将区块高度 时间戳 前区块哈希 序列化交易数据 目标难度值 碰撞次数拼接
	data := bytes.Join([][]byte{
		heightBytes, timeStampBytes,
		p.Block.PrevBlockHash,
		p.Block.HashTransaction(),
		utils.IntToHex(targetBit), utils.IntToHex(nonce),
	}, []byte{})

	return data
}
