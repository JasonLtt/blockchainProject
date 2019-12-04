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

const Usage = `
    ./blockchainProject.exe createBlockChain 地址             创建区块链
	./blockchainProject.exe printChain                        打印区块
	./blockchainProject.exe getBalance 地址                   获取地址的余额
	./blockchainProject.exe send From To Amount Miner Data    转账命令
	./blockchainProject.exe createWallet                      创建钱包
	./blockchainProject.exe listAddress                       打印所有钱包地址
	./blockchainProject.exe printTx                           打印所有交易
`

type CLI struct {
	//bc *BlockChain//CLI中不需要保存区块链实例，所有命令在自己调用之前，自己获取区块链实例

}

func (cli *CLI) Run() {
	cmds := os.Args

	if len(cmds) < 2 {
		fmt.Printf(Usage)
		os.Exit(1)
	}

	switch cmds[1] {
	case "createBlockChain":
		if len(cmds) != 3 {
			fmt.Printf(Usage)
			os.Exit(1)
		}
		fmt.Printf("创建区块链命令被调用!\n")
		addr := cmds[2]
		cli.CreateBlockChain(addr)
	//case "addBlock":
	//	if len(cmds) != 3 {
	//		fmt.Printf(Usage)
	//		os.Exit(1)
	//	}
	//	fmt.Printf("添加区块命令被调用，数据：%s\n", cmds[2])
	// := cmds[2]//TODO
	//cli.AddBlock()

	case "printChain":
		fmt.Printf("打印区块链命令被调用\n")
		cli.PrintChain()

	case "getBalance":
		fmt.Printf("获取余额的命令被调用\n")
		cli.GetBalance(cmds[2])

	case "send":
		fmt.Printf("转账命令被调用\n")
		if len(cmds) != 7 {
			fmt.Printf("send命令发现无效参数，请检查！\n")
			fmt.Printf(Usage)
			os.Exit(1)
		}
		from := cmds[2]
		to := cmds[3]
		amount, _ := strconv.ParseFloat(cmds[4], 64)
		miner := cmds[5]
		data := cmds[6]
		cli.Send(from, to, amount, miner, data)

	case "createWallet":
		fmt.Printf("创建钱包命令被调用\n")
		cli.CreateWallet()

	case "listAddress":
		fmt.Printf("打印钱包地址命令被调用!\n")
		cli.ListAddress()
	case "printTx":
		fmt.Printf("打印交易命令被调用!\n")
		cli.PrintTx()

	default:
		fmt.Printf("无效命令，请检查\n")
		fmt.Printf(Usage)
	}
}
