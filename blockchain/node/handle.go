package node

import (
	"blockchain/block"
	"blockchain/utils"
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"net"
)

// 数据请求进行处理

// HandleConn 处理请求
func HandleConn(conn net.Conn, chain *block.Chain) {
	req, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panicf("receive a request failed: %v", err)
	}
	command := utils.BytesToCommand(req[:12])
	fmt.Printf("receive a command: %s\n", command)

	// 判断命令
	switch command {
	case VERSION:
		HandleVersion(req, chain)
	case GETDATA:
		HandleGetData(req, chain)
	case GETBLOCKS:
		HandleGetBlocks(req, chain)
	case CMDINV:
		HandleInv(req)
	case CMDBLOCK:
		HandleBlock(req, chain)
	default:
		fmt.Println("unknown command")
	}
}

// HandleVersion 处理版本验证
func HandleVersion(req []byte, chain *block.Chain) {
	fmt.Println("the request of version handle...")
	var buff bytes.Buffer
	var data Version

	// 解析请求
	dataBytes := req[12:]
	buff.Write(dataBytes)
	decoder := gob.NewDecoder(&buff)
	if err := decoder.Decode(&data); err != nil {
		log.Panicf("decode the version struct failed: %v\n", err)
	}

	// 生成version结构
	versionHeight := data.Height
	height := chain.GetHeight() // 获取当前高度
	fmt.Printf("height: %v versionHeight: %v\n", height, versionHeight)
	if height > int64(versionHeight) {
		SendVersion(data.AddrFrom, chain)
	} else if height < int64(versionHeight) {
		// 当前节点区块高度小于发送方的versionHeight
		// 向发送方发起同步数据请求
		SendGetBlocks(data.AddrFrom) // 发送请求同步数据
	}
}

// HandleGetData 处理数据获取请求
func HandleGetData(req []byte, chain *block.Chain) {
	fmt.Println("the request of get block handle...")
	var buff bytes.Buffer
	var data GetData

	dataBytes := req[12:]
	buff.Write(dataBytes)
	decoder := gob.NewDecoder(&buff)
	if err := decoder.Decode(&data); err != nil {
		log.Panicf("decode the getData struct failed: %v", err)
	}
	blockBytes := chain.GetBlock(data.ID) // 获取到区块数据
	SendBlock(data.AddrFrom, blockBytes)
}

func HandleGetBlocks(req []byte, c *block.Chain) {
	fmt.Println("the request of get blocks handle...")
	var buff bytes.Buffer
	var data GetBlocks

	dataBytes := req[12:]
	buff.Write(dataBytes)
	decoder := gob.NewDecoder(&buff)
	if err := decoder.Decode(&data); err != nil {
		log.Panicf("decode the getblocks struct failed: %v\n", err)
	}
	hashes := c.GetBlockHashes() // 获取区块链所有区块哈希
	SendInv(data.AddrFrom, hashes)
}

// HandleInv 处理INV请求
func HandleInv(req []byte) {
	fmt.Println("the request of inv handle...")
	var buff bytes.Buffer
	var data Inv
	dataBytes := req[12:]
	buff.Write(dataBytes)
	decoder := gob.NewDecoder(&buff)
	if err := decoder.Decode(&data); err != nil {
		log.Panicf("decode the inv struct failed: %v\n", err)
	}

	// 同步哈希对应的区块数据
	for _, hash := range data.Hashes {
		SendGetData(data.AddrFrom, hash)
	}
}

// HandleBlock 处理请求
func HandleBlock(req []byte, c *block.Chain) {
	fmt.Println("the request of handle block handle...")
	var buff bytes.Buffer
	var data BlockData

	dataBytes := req[12:]
	buff.Write(dataBytes)
	decoder := gob.NewDecoder(&buff)
	if err := decoder.Decode(&data); err != nil {
		log.Panicf("decode the block data struct failed: %v\n", err)
	}

	// 接收到的区块添加到区块链
	blockBytes := data.Block
	newBlock := block.DeserializeBlock(blockBytes)
	c.AddBlock(newBlock)
	utxoSet := block.UTXOSet{Chain: c}
	utxoSet.UpdateUTXOSet()
}
