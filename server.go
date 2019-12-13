package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3001"}
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction)

type addr struct {
	AddrList []string
}

//获取version信息
type version struct {
	Version    int
	BestHeight uint64
	AddrFrom   string
}

//获取所有区块
type getblocks struct {
	AddrFrom string
}

type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
}

//向其他节点展示当前节点有什么块和交易
type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type block struct {
	AddrFrom string
	Block    []byte //todo
}

type Tx struct {
	AddrFrom    string
	Transaction []byte //todo
}

type Addr struct{
	AddrList []string
}

// 中心节点启动监听，监听来自节点的version请求
// 启动时创建一个BlockChain，用nodeID来区分不同的数据库，因为模拟的是不同的节点各自独立的数据库
func StartServer(nodeID, minerAdd string) {
	nodeAddress = fmt.Sprintf("localhost:%s\n", nodeID)
	miningAddress = minerAdd
	ln, err := net.Listen(protocol, nodeAddress)
	if err == nil {
		log.Panic(err)
	}
	defer ln.Close()

	bc := NewBlockChain(nodeID)
	//如果该节点不是中心化节点的话knownNodes[0]，就要向中心化节点发送自己当前的version
	//实际比特币中knownNodes数组是一个 DNS里面随机获取到的几个全节点，都可以作为中心化节点
	if nodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}
	//开启监听 ln由 localhost:nodeID组成，各自循环监听不同的消息
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc)
	}
}

//如果当前节点不是中心节点，则向中心节点发送version信息
//检查自己的区块链是否已经过时
//addr：中心节点的地址
//bc：nodeID的区块链
func sendVersion(addr string, bc *BlockChain) {
	bestHeight := bc.GetBestHeight()
	//指定命令名{nodeVersion：1，BestHeight:当前区块链的长度，nodeAddress:发送给addr的version信息的nodeID}
	payload := gobEncode(version{nodeVersion, bestHeight, nodeAddress})
	//前12个字节指定了命令名（这里是“version”），后面的字节包含gob编码的信息结构
	request := append(CommandToBytes("version"), payload...)

	sendData(addr, request)
}

//通过BytesToCommand提取命令，并在handleConnection选择对应的处理器处理命令
func handleConnection(conn net.Conn, bc *BlockChain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := BytesToCommand(request[:commandLength])
	fmt.Printf("收到 %s 命令!\n", command)
	switch command {
	case "addr":
		handleAddr(request)

	case "block":
		handleBlock(request, bc)

	case "inv":
		handleInv(request, bc)

	case "getblocks":
		handleGetBlocks(request, bc)

	case "getdata":
		handleGetData(request, bc)

	case "tx":
		handleTx(request, bc)

	case "version":
		handleVersion(request, bc)

	default:
		fmt.Println("命令错误，请检查!\n") //todo
	}
	conn.Close()
}

//添加Addr的命令
func handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	knownNodes = append(knownNodes, payload.AddrList...)
	fmt.Printf("现在有 %d 个已知节点！\n", len(knownNodes))
	requestBlocks()
}


func sendAddr(addr string){
	nodes:=Addr{knownNodes}
	nodes.AddrList=append(nodes.AddrList,nodeAddress)
	payload:=gobEncode(nodes)
	request:=append(CommandToBytes("addr"),payload...)

	sendData(addr,request)
}

func handleBlock(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	//todo
	blockData := payload.Block
	block := Deserialize(blockData)

	fmt.Println("收到一个新区块！\n")
	bc.AddBlocks(block)

	fmt.Printf("添加的区块为：%s\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOset := UTXOset{bc}
		//更新UTXO池
		UTXOset.Reindex()
	}
}

//获取当前节点的块和交易
//不包含完整的区块链和交易，只是哈希
//Type字段表明是块还是交易
//todo
//总是需要要从中心服务器全部挨个获取，在比对
func handleInv(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("收到的区块： %d %s!\n", len(payload.Items), payload.Type)
	//在我们的版本中，不会发送多重哈希的inv
	//区块
	if payload.Type == "block" {
		//获取远程的返回过来的全部blockHash数组
		blocksInTransit = payload.Items

		//获取最后一个块的数据
		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}

		//排除掉最后一个块，因为前面已经通过 sendGetData 获取
		blocksInTransit = newInTransit
	}
	//交易
	if payload.Type == "tx" {
		//只取第一个哈希
		txID := payload.Items[0]
		//在内存池中检查是否已经有这个哈希，如果没有，发送getdata消息
		if mempool[hex.EncodeToString(txID)].TXid == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

func sendGetData(addr, kind string, txid []byte) {
	payload := gobEncode(getdata{nodeAddress, kind, txid})
	request := append(CommandToBytes("getdata"), payload...)

	sendData(addr, request)
}

//获取所有块哈希的列表
//实际bitcoin中比这个复杂，需要从单点下载十几GB的数据
func handleGetBlocks(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	//blocks的的第一个区块hash值是区块链的最后一个区块的区块hash
	blocks := bc.GetBlockHashes()
	// 中心服务器 会向 客户端服务器 发送所有区块的BlockHash
	sendInv(payload.AddrFrom, "block", blocks)
}

func sendInv(addr, kind string, items [][]byte) {
	inventory := inv{addr, kind, items}
	payload := gobEncode(inventory)
	request := append(CommandToBytes("inv"), payload...)

	sendData(addr, request)
}

//获取某个块或交易的请求
func handleGetData(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}
		sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]
		//todo
		//未能检查是否该交易存在
		sendTx(payload.AddrFrom, &tx)
	}
}

func sendBlock(addr string, b *Block) {
	data := block{nodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(CommandToBytes("block"), payload...)

	sendData(addr, request)
}

//向中心服务器发送交易添加到交易池的操作
func sendTx(addr string, tx *Transaction) {
	data := Tx{nodeAddress, tx.Serialize()}
	payload := gobEncode(data)
	request := append(CommandToBytes("tx"), payload...)
	sendData(addr, request)
}

func handleTx(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload Tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := DeserializeTransaction(txData)
	mempool[hex.EncodeToString(tx.TXid)] = tx
	//检查该节点是否为中心节点
	//中心节点暂且不会挖矿，只会将新的交易推送给网络中其他节点
	//todo
	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddrFrom {
				sendInv(node, "tx", [][]byte{tx.TXid})
			}
		}
	} else {
		//挖矿节点进行挖矿
		if len(mempool) >= 2 && len(miningAddress) > 0 {
			//将内存池内的交易数量大于等于2笔，且存在挖矿节点
			//则挖矿进行挖矿打包
		MineTransaction:
			var txs []*Transaction

			for id := range mempool {
				tx := mempool[id]
				//将新交易放入内存池中，需要先进行验证
				if bc.VerifyTransaction(&tx) {
					fmt.Printf("---- 交易有效 :%x\n", tx.TXid)
					txs = append(txs, &tx)
				} else {
					fmt.Printf("发现无效交易：%x\n", tx.TXid)
				}
			}

			if len(txs) == 0 {
				fmt.Println("所有的交易都不合法！等待新的交易......\n")
				return
			}

			cbTx := NewCoinBaseTx(miningAddress, "")
			txs = append(txs, cbTx)

			newBlock := bc.MineBlock(txs)
			utxoset := UTXOset{bc}
			utxoset.Reindex()

			fmt.Println("挖到新区块!\n")
			//当一笔交易被挖出来后，则从内存池中移除
			for _, tx := range txs {
				txID := hex.EncodeToString(tx.TXid)
				delete(mempool, txID)
			}

			//将新块的hash发送给其他节点
			//todo
			//后期改进：节点验证
			for _, node := range knownNodes {
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}
			//若果内存池中还有交易，返回上述进行打包交易
			if len(mempool) > 0 {
				goto MineTransaction
			}
		}
	}

}

//version命令处理器
func handleVersion(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload version

	//获取version信息
	buff.Write(request[commandLength:])
	//gob解码
	dec := gob.NewDecoder(&buff)
	//传给payload
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	//获取自身节点的区块链长度
	myBestHeight := bc.GetBestHeight()
	//获取version信息里面的区块链长度
	foreignerBestHeight := payload.BestHeight

	//自身节点区块链长度与version信息的区块链长度比较
	//如果中心服务器的数据不是最新的，发起一个获取最新blocks的请求给普通服务器
	if myBestHeight < foreignerBestHeight {
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bc)
	}

	//如果payload.AddrFrom不在knownNodes中
	//将payload.AddrFrom添加进入knownNodes
	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}

}

//判断addr是不是在knownNodes中
func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}
	return false
}

//
func requestBlocks() {
	for _, node := range knownNodes {
		sendGetBlocks(node)
	}
}

//获取最新的blocks的操作
//address为中心服务器的地址
func sendGetBlocks(address string) {
	payload := gobEncode(getblocks{nodeAddress})
	//命令更改为getblocks
	request := append(CommandToBytes("getblocks"), payload...)

	sendData(address, request)
}

//todo
func sendData(address string, data []byte) {
	//拨号不成功，替换knownNodes中的节点
	// Conn is a generic stream-oriented network connection
	conn, err := net.Dial(protocol, address)
	if err != nil {
		fmt.Printf("%s 节点不可用！\n", address)
		var updataNodes []string

		for _, node := range knownNodes {
			if node != address {
				updataNodes = append(updataNodes, node)
			}
		}
		knownNodes = updataNodes

		return
	}
	defer conn.Close()
	//io.Copy实现了两个文件指针之间的内容拷贝
	//拷贝文件，将副本从src复制到dst中，直到src达到文件末尾或发生错误
	//然后返回复制的字节数和复制遇到的第一个问题
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

//将命令名转换成字节流
func CommandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

//当一个节点接收到一个命令，通过BytesToCommand提取命令名
func BytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		//b！=0
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("命令：%s\n", command)
}

//gob编码
func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}
