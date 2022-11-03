package node

import (
	"blockchain/utils"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

// SendMessage 发送消息
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

func SendVersion(address string) {
	// 获取当前节点区块高度
	height := 1
	versionData := Version{Height: height, AddrFrom: nodeAddr} // 组装生成version
	data := utils.GobEncoder(versionData)
	request := append(utils.CommandToBytes(VERSION), data...)
	SendMessage(address, request)
}

// SendGetData 发送获取指定区块请求
func SendGetData(address string, hash []byte) {
	data := utils.GobEncoder(GetData{AddrFrom: nodeAddr, ID: hash})
	req := append(utils.CommandToBytes(GETDATA), data...)
	SendMessage(address, req)
}

func SendGetBlocks(address string) {
	data := utils.GobEncoder(GetBlocks{AddrFrom: nodeAddr})
	req := append(utils.CommandToBytes(GETBLOCKS), data...)
	SendMessage(address, req)
}

func SendInv(address string, hashes [][]byte) {
	data := utils.GobEncoder(Inv{AddrFrom: nodeAddr, Hashes: hashes})
	req := append(utils.CommandToBytes(CMDINV), data...)
	SendMessage(address, req)
}

// SendBlock 发送区块信息
func SendBlock(address string, block []byte) {
	data := utils.GobEncoder(BlockData{AddrFrom: nodeAddr, Block: block})
	req := append(utils.CommandToBytes(CMDBLOCK), data...)
	SendMessage(address, req)
}
