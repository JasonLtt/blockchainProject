package main

import (
	"fmt"
	"os"
	"strconv"
)

//使用命令行
//cli：command line

//添加区块：bc.AddBlock(data),data 通过osArgs获得
//打印区块：遍历区块，不需要外部输入数据

type CLI struct {
	//bc *BlockChain//CLI中不需要保存区块链实例，所有命令在自己调用之前，自己获取区块链实例
}

func (cli *CLI) printUsage() {
	fmt.Println("./blockchainProject.exe  createBlockChain 地址                                          [创建区块链]")
	fmt.Println("./blockchainProject.exe  printChain                                                     [打印区块]")
	fmt.Println("./blockchainProject.exe  getBalance 地址                                                [获取地址的余额]")
	fmt.Println("./blockchainProject.exe  send From To Amount Data Miner(选填)                           [(暂且必须填写Miner)转账命令]")
	fmt.Println("./blockchainProject.exe  createWallet                                                   [创建钱包]")
	fmt.Println("./blockchainProject.exe  listAddress                                                    [打印所有钱包地址]")
	fmt.Println("./blockchainProject.exe  printTx                                                        [打印所有交易]")
	fmt.Println("./blockchainProject.exe  startNode                                                      [(待调试)StartNode]")
	fmt.Println("./blockchainProject.exe  reindexUtxo                                                    [重新建立UTXO池]")
}

func (cli *CLI) Run() {
	node_ID := "3001"
	cmds := os.Args

	if len(cmds) < 2 {
		cli.printUsage()
		os.Exit(1)
	}

	switch cmds[1] {
	case "createBlockChain":
		if len(cmds) != 3 {
			cli.printUsage()
			os.Exit(1)
		}
		fmt.Printf("创建区块链命令被调用!\n")
		addr := cmds[2]
		cli.CreateBlockChain(node_ID, addr)

	case "printChain":
		fmt.Printf("打印区块链命令被调用\n")
		cli.PrintChain(node_ID)

	case "getBalance":
		fmt.Printf("获取余额的命令被调用\n")
		cli.GetBalance(cmds[2], node_ID)

	case "send":
		fmt.Printf("转账命令被调用\n")
		if len(cmds) < 6 || len(cmds) > 7 {
			fmt.Printf("send命令发现无效参数，请检查！\n")
			cli.printUsage()
			os.Exit(1)
		}
		from := cmds[2]
		to := cmds[3]
		amount, _ := strconv.ParseFloat(cmds[4], 64)
		data := cmds[5]
		if len(cmds)==7 {
			miner := cmds[6]
			cli.Send(from, to, amount, miner, data, node_ID, true)
		} else {
			//todo
			//暂且无法使用
			cli.Send(from, to, amount, "nil", data, node_ID, false)
		}

	case "createWallet":
		fmt.Printf("创建钱包命令被调用\n")
		cli.CreateWallet(node_ID)

	case "listAddress":
		fmt.Printf("打印钱包地址命令被调用!\n")
		cli.ListAddress(node_ID)

	case "printTx":
		fmt.Printf("打印交易命令被调用!\n")
		cli.PrintTx(node_ID)

		//todo
		//暂且无法使用
	case "startNode":
		fmt.Printf("startnode命令被调用！\n")
		if len(cmds) !=3 {
			fmt.Println("错误!")
			cli.printUsage()
			os.Exit(1)
		}
		//暂且先不从环境变量获取node_ID，先直接赋予
		//nodeId := cmds[2]
		miner := cmds[2]
		cli.stratNode(node_ID, miner)

	case "reindexUtxo":
		fmt.Printf("重建建立UTXO池命令被调用！\n")
		cli.reindexUtxo(node_ID)

	default:
		fmt.Printf("无效命令，请检查!\n")
		cli.printUsage()
	}
}
