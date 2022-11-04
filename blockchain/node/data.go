package node

type Version struct {
	Height   int    // 当前节点区块高度
	AddrFrom string // 当前节点地址
}

type BlockData struct {
	AddrFrom string
	Block    []byte
}

type GetBlocks struct {
	AddrFrom string
}

type GetData struct {
	AddrFrom string
	ID       []byte // 区块哈希
}

type Inv struct {
	AddrFrom string // 当前节点地址
	Hashes   [][]byte
}
