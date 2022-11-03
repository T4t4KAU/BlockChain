package cmd

import (
	"blockchain/block"
	"blockchain/utils"
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

func IsValidArgs() {
	if len(os.Args) < 2 {
		PrintUsage()
		// 直接退出
		os.Exit(1)
	}
}

func PrintUsage() {
	fmt.Println("Usage: ")
	fmt.Println("\tcreatewallet -- create wallet")
	fmt.Println("\taccounts -- list accounts")
	// 创建区块链
	fmt.Println("\tcreatechain -address address -- create blockchain")
	// 添加区块
	fmt.Println("\taddblock -data data -- add a block")
	// 打印区块链完整信息
	fmt.Println("\tprintchain -- print blockchain")
	// 获取余额信息
	fmt.Println("\tgetbalance -address address -- get balance of address")

	// 通过命令行转账
	fmt.Println("\tsend -from FROM -to TO -amount AMOUNT -- Initiate a transfer")
	fmt.Println("\tdescription of the transfer parameters:")
	fmt.Println("\t\t-from FROM -- the source address of the transfer")
	fmt.Println("\t\t-to TO -- The destination address of the transfer")
	fmt.Println("\t\t-amount AMOUNT -- The amount transferred")
	fmt.Println("\tutxo -test METHOD -- test methods of UTXO table")
	fmt.Println("\t\tMETHOD -- name of method")
	fmt.Println("\t\t\tbalance -- find all the UTXOs")
	fmt.Println("\t\t\treset -- reset UTXO table")
	fmt.Println("\tsetid -port PORT")
}

// Run 命令行
func (cli *Client) Run() {
	IsValidArgs() // 检测命令行参数个数
	fmt.Printf(message)

	GetAccountsCmd := flag.NewFlagSet("accounts", flag.ExitOnError)
	CreateWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)               // 创建钱包
	AddBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)                       // 新建相关命令 添加区块
	PrintChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)                   // 输出区块链完整信息
	CreateChainWithGenesisBlockCmd := flag.NewFlagSet("createchain", flag.ExitOnError) // 创建区块链
	SendCmd := flag.NewFlagSet("send", flag.ExitOnError)                               // 发起交易
	GetBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)                   // 查询余额命令
	UTXOTestCmd := flag.NewFlagSet("utxo", flag.ExitOnError)
	StartNodeCmd := flag.NewFlagSet("start", flag.ExitOnError)

	flagAddBlockArg := AddBlockCmd.String("data", "", "add block")                              //  数据参数处理
	flagCreateChainArg := CreateChainWithGenesisBlockCmd.String("address", "", "miner address") // 创建区块链的矿工地址 接收奖励

	// 发起交易参数
	flagSendFromArg := SendCmd.String("from", "", "The source address of the transfer")  // 交易源地址
	flagSendToArg := SendCmd.String("to", "", "The destination address of the transfer") // 交易目标地址
	flagSendAmountArg := SendCmd.String("amount", "", "The amount transferred")          // 交易额度

	// 查询余额命令行参数
	flagGetBalanceArg := GetBalanceCmd.String("address", "", "The address to query")
	flagUTXOArg := UTXOTestCmd.String("method", "", "UTXO table related actions")
	// flagStartNodeArg := StartNodeCmd.String("start", "", "start node")

	// 判断参数
	switch os.Args[1] {
	case "start":
		if err := StartNodeCmd.Parse(os.Args[2:]); err != nil {
			log.Panicf("parse cmd start node server failed: %v\n", err)
		}
	case "utxo":
		err := UTXOTestCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicf("parse cmd operate utxo table failed! %v\n", err)
		}
	case "accounts":
		err := GetAccountsCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicf("parse cmd create account failed: %v\n", err)
		}
	case "createwallet": // 创建钱包
		err := CreateWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicf("parse cmd of create wallet failed: %v\n", err)
		}
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

	nodeId := os.Getenv("NODE_ID")
	fmt.Println("NODE ID:", nodeId)

	// 解析命令行参数

	if StartNodeCmd.Parsed() {
		cli.startNode(nodeId)
	}

	if UTXOTestCmd.Parsed() {
		switch *flagUTXOArg {
		case "balance":
			cli.TestFindUTXOMap()
		case "reset":
			cli.TestResetUTXO(nodeId)
		}
	}

	if GetAccountsCmd.Parsed() {
		cli.GetAccounts(nodeId)
	}

	if CreateWalletCmd.Parsed() {
		cli.CreateWallets(nodeId)
	}

	if GetBalanceCmd.Parsed() {
		if *flagGetBalanceArg == "" {
			fmt.Println("Input the address to query")
			os.Exit(1)
		}
		cli.getBalance(*flagGetBalanceArg, nodeId)
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
		fmt.Printf("FROM: %s\n", utils.JsonToSlice(*flagSendFromArg))
		fmt.Printf("TO: %s\n", utils.JsonToSlice(*flagSendToArg))
		fmt.Printf("AMOUNT: %s\n", utils.JsonToSlice(*flagSendAmountArg))
		cli.Send(utils.JsonToSlice(*flagSendFromArg),
			utils.JsonToSlice(*flagSendToArg),
			utils.JsonToSlice(*flagSendAmountArg), nodeId)
	}

	if AddBlockCmd.Parsed() {
		if *flagAddBlockArg == "" {
			PrintUsage()
			os.Exit(1)
		}
		cli.AddBlock([]*block.Transaction{}, nodeId)
	}
	if PrintChainCmd.Parsed() {
		cli.PrintChain(nodeId)
	}
	if CreateChainWithGenesisBlockCmd.Parsed() {
		if *flagCreateChainArg == "" {
			PrintUsage()
			os.Exit(1)
		}
		cli.CreateBlockChain(*flagCreateChainArg, nodeId)
	}
}
