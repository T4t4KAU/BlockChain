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

// HandleConn 处理请求
func HandleConn(conn net.Conn, chain *block.Chain) {
	req, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panicf("receive a request failed: %v", err)
	}
	command := utils.BytesToCommand(req[:12])
	fmt.Printf("receive a command: %s\n", command)
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

func HandleVersion(req []byte, chain *block.Chain) {
	fmt.Println("the request of version handle...")
	var buff bytes.Buffer
	var data Version

	dataBytes := req[12:]
	buff.Write(dataBytes)
	decoder := gob.NewDecoder(&buff)

	if err := decoder.Decode(&data); err != nil {
		log.Panicf("decode the version struct failed: %v\n", err)
	}
	versionHeight := data.Height
	height := chain.GetHeight()
	fmt.Printf("height: %v versionHeight: %v\n", height, versionHeight)
	if height > int64(versionHeight) {

	} else if height < int64(versionHeight) {
		SendGetBlocks(data.AddrFrom)
	}
}

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
	blockBytes := chain.GetBlock(data.ID)
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
	hashes := c.GetBlockHashes()
	SendInv(data.AddrFrom, hashes)
}

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
	SendGetData(data.AddrFrom, data.Hashes[0])
}

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

	blockBytes := data.Block
	newBlock := block.DeserializeBlock(blockBytes)
	c.AddBlock(newBlock)
	utxoSet := block.UTXOSet{Chain: c}
	utxoSet.UpdateUTXOSet()
}
