package node

import (
	"blockchain/block"
	"blockchain/utils"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

// SendMessage 向指定地址发送数据
func SendMessage(to string, msg []byte) {
	fmt.Println("send request to server...")
	conn, err := net.Dial("tcp", to)
	if err != nil {
		log.Panicf("connect to server[%s] failed: %v\n", to, err)
	}
	_, err = io.Copy(conn, bytes.NewReader(msg))
	if err != nil {
		log.Panicf("add the data to conn failed: %v\n", err)
	}
}

// SendVersion 区块链版本验证
func SendVersion(toAddress string, chain *block.Chain) {
	// 获取当前节点区块高度
	height := chain.GetHeight()
	versionData := Version{Height: int(height), AddrFrom: nodeAddr} // 组装生成version
	data := utils.GobEncoder(versionData)
	request := append(utils.CommandToBytes(VERSION), data...)
	SendMessage(toAddress, request)
}

// SendGetData 发送获取指定区块请求
func SendGetData(toAddress string, hash []byte) {
	data := utils.GobEncoder(GetData{AddrFrom: nodeAddr, ID: hash})
	req := append(utils.CommandToBytes(GETDATA), data...)
	SendMessage(toAddress, req)
}

// SendGetBlocks 从指定结点同步数据
func SendGetBlocks(address string) {
	data := utils.GobEncoder(GetBlocks{AddrFrom: nodeAddr})
	req := append(utils.CommandToBytes(GETBLOCKS), data...)
	SendMessage(address, req)
}

// SendInv 向其他节点展示
func SendInv(toAddress string, hashes [][]byte) {
	data := utils.GobEncoder(Inv{AddrFrom: nodeAddr, Hashes: hashes})
	req := append(utils.CommandToBytes(CMDINV), data...)
	SendMessage(toAddress, req)
}

// SendBlock 发送区块信息
func SendBlock(toAddress string, block []byte) {
	data := utils.GobEncoder(BlockData{AddrFrom: nodeAddr, Block: block})
	req := append(utils.CommandToBytes(CMDBLOCK), data...)
	SendMessage(toAddress, req)
}
