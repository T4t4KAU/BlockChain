package cmd

import (
	"blockchain/block"
	"blockchain/node"
	"blockchain/wallet"
	"fmt"
	"log"
	"os"
)

// 实现命令行完整逻辑

// Send 发起转账
func (cli *Client) Send(from, to, amount []string, nodeId string) {
	if !block.IsDBExists(nodeId) {
		fmt.Println("database not exists")
		os.Exit(1)
	}
	chain := block.GetBlockChainObject(nodeId)
	defer chain.DB.Close()
	if len(from) != len(to) || len(from) != len(amount) {
		fmt.Println("the sender and receiver are inconsistent...")
		os.Exit(1)
	}
	chain.MineNewBlock(from, to, amount, nodeId) // 挖掘一个新区块储存转账信息

	utxoSet := &block.UTXOSet{Chain: chain}
	utxoSet.UpdateUTXOSet()
}

// CreateBlockChain 初始化区块链
func (cli *Client) CreateBlockChain(address string, nodeId string) {
	cli.BlockChain = block.CreateBlockChainWithGenesisBlock(address, nodeId)
	defer cli.BlockChain.DB.Close()

	utxoSet := &block.UTXOSet{Chain: cli.BlockChain}
	utxoSet.ResetUTXOSet()
}

// AddBlock 添加区块
func (cli *Client) AddBlock(txs []*block.Transaction, nodeId string) {
	cli.BlockChain = block.GetBlockChainObject(nodeId)
	latestBlock := cli.BlockChain.GetLatestBlock()
	newBlock := block.NewBlock(latestBlock.Height+1, latestBlock.Hash, txs)
	cli.BlockChain.AddBlock(newBlock)
}

// PrintChain 打印完整区块链信息
func (cli *Client) PrintChain(nodeId string) {
	cli.BlockChain = block.GetBlockChainObject(nodeId)
	cli.BlockChain.PrintChain(nodeId)
}

// CreateWallets 创建钱包集合
func (cli *Client) CreateWallets(nodeId string) {
	wallets := wallet.NewWallets(nodeId) // 创建一个集合对象
	wallets.CreateWallet(nodeId)
}

// GetAccounts 获取账户列表
func (cli *Client) GetAccounts(nodeId string) {
	wallets := wallet.NewWallets(nodeId)
	fmt.Println("account list:")
	for key := range wallets.Wallets {
		fmt.Printf(" [%s]\n", key)
	}
}

// TestResetUTXO 重置UTXO
func (cli *Client) TestResetUTXO(nodeId string) {
	chain := block.GetBlockChainObject(nodeId)
	defer chain.DB.Close()
	utxoSet := block.UTXOSet{Chain: chain}
	utxoSet.ResetUTXOSet()
}

func (cli *Client) TestFindUTXOMap() {

}

// 获取账户余额 即计算地址对应账户下的UTXO交易输出和
func (cli *Client) getBalance(from string, nodeId string) {
	// 查找指定地址的UTXO
	chain := block.GetBlockChainObject(nodeId)
	defer chain.DB.Close()
	utxoSet := block.UTXOSet{Chain: chain}
	amount := utxoSet.GetBalance(from)
	fmt.Printf("balance of address[%s]: %d\n", from, amount)
}

// SetNodeId 设置端口号
func (cli *Client) SetNodeId(nodeId string) {
	if nodeId == "" {
		fmt.Println("please set port")
		os.Exit(1)
	}
	err := os.Setenv("NODE_ID", nodeId)
	if err != nil {
		log.Fatalf("set env failed: %v\n", err)
	}
	fmt.Println("set NODE_ID:", nodeId)
}

// 启动节点
func (cli *Client) startNode(nodeId string) {
	node.StartServer(nodeId)
}
