package node

import (
	"blockchain/block"
	"fmt"
	"log"
	"net"
	"strconv"
)

// 节点之间须要进行数据同步
//

const (
	PROTOCOL  = "tcp"
	VERSION   = "version"
	GETBLOCKS = "getblocks"
	CMDINV    = "inv"
	GETDATA   = "getdata"
	CMDBLOCK  = "block"
)

var (
	port       = 3000
	knownNodes = []string{"localhost:" + strconv.Itoa(port)} // 主节点地址
	nodeAddr   string                                        // 节点地址
)

// StartServer 启动服务
func StartServer(nodeId string) {
	nodeAddr = fmt.Sprintf("localhost:%s", nodeId)
	// 监听节点
	listen, err := net.Listen(PROTOCOL, nodeAddr)
	if err != nil {
		log.Panicf("listen address of %s failed: %v\n", nodeAddr, err)
	}
	defer listen.Close()

	// 主节点负责保存数据 钱包节点负责发送请求
	// 判断是否为主节点 非主节点则发送请求 同步数据

	chain := block.GetBlockChainObject(nodeId)
	if nodeAddr != knownNodes[0] {
		SendVersion(knownNodes[0], chain)
	}

	for {
		conn, e := listen.Accept()
		if e != nil {
			log.Panicf("accept connect failed: %v\n", err)
		}
		go HandleConn(conn, chain) // 使用线程单独处理请求
	}
}
