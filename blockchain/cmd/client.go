package cmd

import (
	"blockchain/block"
	"flag"
	"fmt"
	"log"
	"os"
)

const message = " ____  _            _     ____ _           _       \n| __ )| | ___   ___| | __/ ___| |__   __ _(_)_ __  \n|  _ \\| |/ _ \\ / __| |/ / |   | '_ \\ / _` | | '_ \\ \n| |_) | | (_) | (__|   <| |___| | | | (_| | | | | |\n|____/|_|\\___/ \\___|_|\\_\\\\____|_| |_|\\__,_|_|_| |_|\n                                                   \n"

// Client 客户端对象
type Client struct {
	BlockChain *block.Chain
}

func PrintUsage() {
	fmt.Println("Usage: ")
	// 创建区块链
	fmt.Println("\tcreatechain -address address -- create blockchain")
	// 添加区块
	fmt.Println("\taddblock -data data -- add a block")
	// 打印区块链完整信息
	fmt.Println("\tprintchain -- print blockchain")
	// 获取余额信息
	fmt.Println("\tgetbalance -- get balance")

	// 通过命令行转账
	fmt.Println("\t-from FROM -to TO -amount AMOUNT -- Initiate a transfer")
	fmt.Println("\t\tdescription of the transfer parameters:")
	fmt.Println("\t\t-to TO -- The destination address of the transfer")
	fmt.Println("\t\tAMOUNT amount -- The amount transferred")
}

func (cli *Client) Send(from, to, amount []string) {
	if !block.IsDBExists() {
		fmt.Println("db not exists")
		os.Exit(1)
	}
	chain := block.GetBlockChainObject()
	defer chain.DB.Close()
	chain.MineNewBlock(from, to, amount)
}

// CreateBlockChain 初始化区块链
func (cli *Client) CreateBlockChain(address string) {
	cli.BlockChain = block.CreateBlockChainWithGenesisBlock(address)
}

// AddBlock 添加区块
func (cli *Client) AddBlock(txs []*block.Transaction) {
	cli.BlockChain = block.GetBlockChainObject()
	cli.BlockChain.AddBlock(txs)
}

// PrintChain 打印完整区块链信息
func (cli *Client) PrintChain() {
	cli.BlockChain = block.GetBlockChainObject()
	cli.BlockChain.PrintChain()
}

func IsValidArgs() {
	if len(os.Args) < 2 {
		PrintUsage()
		// 直接退出
		os.Exit(1)
	}
}

// 获取账户余额
func (cli *Client) getBalance(from string) {
	// 查找指定地址的UTXO
	chain := block.GetBlockChainObject()
	defer chain.DB.Close()
	fmt.Printf("balance of address[%s]: %d\n",
		from, chain.GetBalance(from))
}

// Run 命令行
func (cli *Client) Run() {
	IsValidArgs() // 检测命令行参数个数
	fmt.Printf(message)

	AddBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)                       // 新建相关命令 添加区块
	PrintChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)                   // 输出区块链完整信息
	CreateChainWithGenesisBlockCmd := flag.NewFlagSet("createchain", flag.ExitOnError) // 创建区块链
	SendCmd := flag.NewFlagSet("send", flag.ExitOnError)                               // 发起交易
	GetBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)                   // 查询余额命令

	flagAddBlockArg := AddBlockCmd.String("data", "", "add block")                              //  数据参数处理
	flagCreateChainArg := CreateChainWithGenesisBlockCmd.String("address", "", "miner address") // 创建区块链的矿工地址 接收奖励

	// 发起交易参数
	flagSendFromArg := SendCmd.String("from", "", "The source address of the transfer")  // 交易源地址
	flagSendToArg := SendCmd.String("to", "", "The destination address of the transfer") // 交易目标地址
	flagSendAmountArg := SendCmd.String("amount", "", "The amount transferred")          // 交易额度

	// 查询余额命令行参数
	flagGetBalanceArg := GetBalanceCmd.String("address", "", "The address to query")

	// 判断参数
	switch os.Args[1] {
	case "getbalance": // 获取余额
		if err := GetBalanceCmd.Parse(os.Args[2:]); err != nil {
			log.Panicf("parse cmd get balance failed: %v\n", err)
		}
	case "send": // 发起交易参数
		if err := SendCmd.Parse(os.Args[2:]); err != nil {
			log.Panicf("parse sendCmd failed %v", err)
		}
	case "addblock": // 添加区块
		if err := AddBlockCmd.Parse(os.Args[2:]); err != nil {
			log.Panicf("parse addBlockCmd failed %v\n", err)
		}
	case "printchain": // 输出公链
		if err := PrintChainCmd.Parse(os.Args[2:]); err != nil {
			log.Panicf("parse printChainCmd failed %v\n", err)
		}
	case "createchain": // 创建区块链
		if err := CreateChainWithGenesisBlockCmd.Parse(os.Args[2:]); err != nil {
			log.Panicf("parse createBLCWithGenesisBlockCmd failed %v\n", err)
		}
	default:
		// 没有传递任何命令或者传递的命令不在命令列表中
		PrintUsage()
		os.Exit(1)
	}

	if GetBalanceCmd.Parsed() {
		if *flagGetBalanceArg == "" {
			fmt.Println("Input the address to query")
			os.Exit(1)
		}
		cli.getBalance(*flagGetBalanceArg)
	}

	if SendCmd.Parsed() {
		if *flagSendFromArg == "" {
			fmt.Println("The source address cannot be empty")
			PrintUsage()
			os.Exit(1)
		}
		if *flagSendToArg == "" {
			fmt.Println("The destination address cannot be empty")
			PrintUsage()
			os.Exit(1)
		}
		if *flagSendAmountArg == "" {
			fmt.Println("The transfer amount cannot be empty")
			PrintUsage()
			os.Exit(1)
		}
		fmt.Printf("\tFROM:[%s]\n", block.JsonToSlice(*flagSendFromArg))
		fmt.Printf("\tTO:[%s]\n", block.JsonToSlice(*flagSendToArg))
		fmt.Printf("\tAMOUNT:[%s]\n", block.JsonToSlice(*flagSendAmountArg))
		cli.Send(block.JsonToSlice(*flagSendFromArg),
			block.JsonToSlice(*flagSendToArg), block.JsonToSlice(*flagSendAmountArg))
	}

	if AddBlockCmd.Parsed() {
		if *flagAddBlockArg == "" {
			PrintUsage()
			os.Exit(1)
		}
		cli.AddBlock([]*block.Transaction{})
	}
	if PrintChainCmd.Parsed() {
		cli.PrintChain()
	}
	if CreateChainWithGenesisBlockCmd.Parsed() {
		if *flagCreateChainArg == "" {
			PrintUsage()
			os.Exit(1)
		}
		cli.CreateBlockChain(*flagCreateChainArg)
	}
}
