package main

import (
	"bytes"
	"fmt"
	"log"
	"time"
)

//func (cli *CLI) AddBlock(txs []*Transaction) {
//	cli.bc.AddBlock(txs)
//	fmt.Printf("添加区块成功!\n")
//}

func (cli *CLI) CreateBlockChain(node_ID string, addr string) {
	if !IsValidAddress(addr) {
		fmt.Printf("无效地址！\n")
		return
	}
	bc := CreateBlockChain(addr, node_ID)
	//if bc==nil{
	//	fmt.Print("创建失败")
	//}
	if bc != nil {
		defer bc.db.Close()
	}
	fmt.Print("创建区块链成功！\n")
}

func (cli *CLI) GetBalance(addr string, node_ID string) {
	if !IsValidAddress(addr) {
		fmt.Printf("无效地址！\n")
		return
	}

	bc := NewBlockChain(node_ID)
	if bc == nil {
		return
	}

	utxoset := UTXOset{bc}
	defer bc.db.Close()
	bc.GetBalance(addr, &utxoset)
}

func (cli *CLI) PrintChain(node_ID string) {
	bc := NewBlockChain(node_ID)
	if bc == nil {
		return
	}
	defer bc.db.Close()
	it := bc.NewIterator()

	for {
		block := it.Next()
		fmt.Printf("+++++++++++++++++++++++++++++++++++++++++\n")
		fmt.Printf("Version:%d\n", block.Version)
		fmt.Printf("PrevBlockHash:%x\n", block.PrevBlockHash)
		fmt.Printf("MerKleRoot:%x\n", block.MerKleRoot)
		//更改时间格式
		TimeFormat := time.Unix(int64(block.TimeStamp), 0).Format("2006-01-02 15:04:05")
		fmt.Printf("TimeStamp:%s\n", TimeFormat)

		fmt.Printf("Difficulty:%d\n", block.Difficulty)
		fmt.Printf("Nonce:%d\n", block.Nonce)
		fmt.Printf("Hash:%x\n", block.Hash)
		fmt.Printf("Height:%d\n", block.Height)
		fmt.Printf("Transaction:\n")
		for _, tx := range block.Transactions {
			//fmt.Printf("Data:%s\n", block.Transactions[0].TXInputs[0].PubKey)
			fmt.Println(tx)
		}

		pow := NewProofOfWork(block)
		fmt.Printf("IsValid:%v\n", pow.IsValid())

		if bytes.Equal(block.PrevBlockHash, []byte{}) {
			fmt.Printf("区块链遍历结束\n")
			break
		}
	}
}

func (cli *CLI) Send(from, to string, amount float64, miner string, data string, node_ID string, mineNow bool) {
	if !IsValidAddress(from) {
		fmt.Printf("from: %s 是无效地址！\n", from)
		return
	}

	if !IsValidAddress(to) {
		fmt.Printf("to: %s 是无效地址！\n", to)
		return
	}

	if mineNow {
		if !IsValidAddress(miner) {
			fmt.Printf("miner: %s 是无效地址！\n", miner)
			return
		}
	}
	bc := NewBlockChain(node_ID)
	if bc == nil {
		return
	}
	utxoset := UTXOset{bc}
	defer bc.db.Close()

	//创建钱包
	wallets, err := NewWallets(node_ID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)
	//1.创建普通交易
	tx := NewTransaction(&wallet, to, amount, bc, &utxoset)

	if tx != nil {
		if mineNow {
			//2.创建挖矿交易
			coinbase := NewCoinBaseTx(miner, data)
			txs := []*Transaction{coinbase}
			txs = append(txs, tx)

			//3.添加到区块
			bc.MineBlock(txs)
			fmt.Printf("挖矿成功！")
		} else {
			sendTx(knownNodes[0], tx)
		}
	} else {
		fmt.Printf("发现无效交易，过滤!\n")
	}
}

func (cli *CLI) CreateWallet(node_ID string) {
	//	w := NewWalletKeyPair()
	ws, _ := NewWallets(node_ID)
	address := ws.CreateWallet(node_ID)

	fmt.Printf("新的钱包地址为：%s\n", address)
}

func (cli *CLI) ListAddress(node_ID string) {
	ws, err := NewWallets(node_ID)
	if err != nil {
		log.Panic(err)
	}

	addresses := ws.ListAddress()
	for _, address := range addresses {
		fmt.Printf("address : %s\n", address)
	}
}

func (cli *CLI) PrintTx(node_ID string) {
	bc := NewBlockChain(node_ID)
	if bc == nil {
		return
	}
	defer bc.db.Close()
	it := bc.NewIterator()

	for {
		block := it.Next()
		fmt.Printf("\n+++++++++++++++++++ 新的区块 ++++++++++++++++++++\n")
		for _, tx := range block.Transactions {
			fmt.Printf("tx : %v\n", tx)
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) stratNode(nodeID, minerAddress string) {
	fmt.Sprintf("开始节点:%s\n", nodeID)
	if len(minerAddress) > 0 {
		if IsValidAddress(minerAddress) {
			fmt.Printf("挖矿奖励的节点：%s\n", minerAddress)
		} else {
			fmt.Printf("挖矿节点错误！\n")
		}
	}
	StartServer(nodeID, minerAddress)
}

func (cli *CLI) reindexUtxo(node_ID string) {
	bc := NewBlockChain(node_ID)
	utxoset := UTXOset{bc}
	utxoset.Reindex()

	count := utxoset.countTransaction()
	fmt.Printf("重置UTXO池成功！此时拥有 %d 笔交易在UTXO池中！\n", count)
}
