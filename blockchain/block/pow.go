package block

import (
	"bytes"
	"crypto/sha256"
	"math/big"
)

// 共识算法管理文件 实现POW实例

const targetBit = 16 // 目标难度值

// ProofOfWork 工作量证明结构
type ProofOfWork struct {
	// 要共识验证的区块
	Block *Block
	// 目标难度的哈希(大数存储)
	Target *big.Int
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
	// 无限循环
	for {
		// 生成准备数据
		dataBytes := p.prepareData(int64(nonce))
		hash = sha256.Sum256(dataBytes)
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
	timeStampBytes := IntToHex(p.Block.TimeStamp)
	heightBytes := IntToHex(p.Block.Height)
	// 数据拼接
	data := bytes.Join([][]byte{
		heightBytes,
		timeStampBytes,
		p.Block.PrevBlockHash,
		p.Block.HashTransaction(),
		IntToHex(targetBit),
		IntToHex(nonce)}, []byte{})

	return data
}
