package node

type BlockData struct {
	AddrFrom string
	Block    []byte
}

type GetBlocks struct {
	AddrFrom string
}

type GetData struct {
	AddrFrom string
	ID       []byte
}

type Inv struct {
	AddrFrom string // 当前节点地址
	Hashes   [][]byte
}
