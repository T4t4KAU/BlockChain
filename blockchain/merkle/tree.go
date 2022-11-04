package merkle

import "crypto/sha256"

// Merkle树是一种二叉树结构 由一个根节点 一组中间节点和叶节点组成树
// 所有节点都存储了哈希值
// 叶节点: 对于一个区块而言，每一笔交易数据 进行哈希运算后 得到的哈希值就是叶节点
// 中间节点: 子节点两两匹配 子节点哈希值合并成新的字符串 对合并结果再次进行哈希运算 得到的哈希值
// 根节点：有且只有一个 为终止节点
// Merkle树是从下往上逐层计算的
// 每个中间节点是根据相邻的两个叶子节点组合计算得出的
// 根节点是根据两个中间节点组合计算得出的

type Tree struct {
	Root *Node // 根节点
}

type Node struct {
	Left  *Node // 左节点
	Right *Node // 右节点
	Data  []byte
}

// MakeNode 创建Merkle节点
func MakeNode(left, right *Node, data []byte) *Node {
	node := &Node{}
	// 判断叶子节点
	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		node.Data = hash[:]
	} else {
		// 将左右节点数据拼接取哈希值
		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		node.Data = hash[:]
	}
	node.Left = left
	node.Right = right
	return node
}

// NewTree 创建Merkle树
func NewTree(txHashes [][]byte) *Tree {
	var nodes []Node
	if len(txHashes)%2 != 0 {
		txHashes = append(txHashes, txHashes[len(txHashes)-1])
	}

	// 生成叶子节点
	for _, data := range txHashes {
		node := MakeNode(nil, nil, data)
		nodes = append(nodes, *node)
	}

	// 自下而上计算中间节点 直至根节点
	for i := 0; i < len(txHashes); i++ {
		var parentNodes []Node
		for j := 0; j < len(nodes); j += 2 {
			// 通过相邻两个节点生成一个父节点
			node := MakeNode(&nodes[j], &nodes[j+1], nil)
			parentNodes = append(parentNodes, *node) // 放入父节点
		}
		if len(parentNodes)%2 != 0 {
			parentNodes = append(parentNodes, parentNodes[len(parentNodes)-1])
		}
		nodes = parentNodes
	}
	tree := Tree{&nodes[0]}
	return &tree // 返回根节点
}
