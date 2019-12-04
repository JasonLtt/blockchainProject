package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

//1. 定义结构（区块头的字段比正常的少）
//>1. 前区块哈希
//>2. 当前区块哈希
//>3. 数据

//2. 创建区块
//3. 生成哈希
//4. 引入区块链
//5. 添加区块
//6. 重构代码

const genesisInfo = "The Times 09/Nov/2019 Chancellor on brink of second bailout for banks"

//定义区块
type Block struct {
	Version uint64 //区块版本号

	PrevBlockHash []byte //前区块哈希

	MerKleRoot []byte //先为空，v4版本再添加

	TimeStamp uint64 //时间戳，从1970.1.1至今的秒数

	Difficulty uint64 //挖矿的难度值，v2时使用

	Nonce uint64 //随机数

	//Data []byte //数据，当前使用数据流，v4时使用交易来代替
	Transactions []*Transaction

	Hash []byte //当前区块的哈希，区块本不存在哈希，只是为了方便，将哈希添加进来
}

//模拟MerKleRoot，只是简单的全部拼接，不是利用二叉树进行hash运算
//func (block *Block) HashTransaction() {
//	//将所有交易的ID拼接，整体做hash运算，作为MerKleRoot
//	var hashes []byte
//	for _, tx := range block.Transactions {
//		txid := tx.TXid
//		hashes = append(hashes, txid...)
//	}
//	hash := sha256.Sum256(hashes)
//	block.MerKleRoot = hash[:]
//}

//计算MerkleRoot，利用二叉树进行hash运算
func (block *Block) HashTransaction() {
	//将所有交易的ID拼接，整体做hash运算，作为MerKleRoot
	var data [][]byte
	for _, tx := range block.Transactions {
		txid := tx.TXid
		data = append(data, txid)
	}
	root:=NewMerkleTree(data)
	block.MerKleRoot =root.RootNode.Data[:]
}

//创建区块
func NewBlock(txs []*Transaction /*data string*/, prevblockhash []byte) *Block {
	block := Block{
		Version:       00,
		PrevBlockHash: prevblockhash,
		MerKleRoot:    []byte{},
		TimeStamp:     uint64(time.Now().Unix()),
		Difficulty:    Bits,
		//	Nonce:         10,
		//Data: []byte(data),
		Transactions: txs,
		Hash:         []byte{},
	}

	block.HashTransaction()

	//	block.SetHash()
	//切换成工作了证明计算hash
	pow := NewProofOfWork(&block)
	hash, nonce := pow.Run()
	block.Hash = hash
	block.Nonce = nonce
	return &block
}

//序列化
func (block *Block) Serialize() []byte {
	var buffer bytes.Buffer
	//定义编码器
	encoder := gob.NewEncoder(&buffer)
	//编码器对结构进行编码，一定要进行校验
	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return buffer.Bytes()
}

//反序列化
func Deserialize(data []byte) *Block {
	//fmt.Printf("解码传入的数据：%x\n",data)

	var block Block
	//创建解码器
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}

////实现哈希
//func (block *Block) SetHash() {
//	//var data []byte //定义全局变量
//	//data = append(data, uintToByte(block.Version)...)
//	//data = append(data, block.PrevBlockHash...)
//	//data = append(data, block.MerKleRoot...)
//	//data = append(data, uintToByte(block.TimeStamp)...)
//	//data = append(data, uintToByte(block.Difficulty)...)
//	//data = append(data, uintToByte(block.Nonce)...)
//	//data = append(data, block.Data...)
//
//	tmp := [][]byte{
//		uintToByte(block.Version),
//		block.PrevBlockHash,
//		block.MerKleRoot,
//		uintToByte(block.TimeStamp),
//		uintToByte(block.Difficulty),
//		uintToByte(block.Nonce),
//		block.Data,
//	}
//	data := bytes.Join(tmp, []byte{})
//	hash := sha256.Sum256(data)
//	block.Hash = hash[:]
//}
