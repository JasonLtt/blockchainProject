package main

import (
	"./base58"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
	"strings"
)

//交易输入（TXInput）
//1.应用utxo的交易ID
//2.对应交易ID下的索引（交易的具体位置）
//3.解锁脚本（签名，公钥）
//
//交易输出（TXOutput）
//1.接收的金额
//2.锁定脚本（对方公钥的哈希，这个哈希可以通过地址反推，所以转账时知道地址即可）
//
//交易ID
//一般是交易结构的哈希值

type TXInput struct {
	TXID  []byte //交易ID
	Index int64  //索引
	//	Address string //解锁脚本（签名，公钥）
	Signature []byte //交易签名

	PubKey []byte //公钥本身，不是哈希
}

type TXOutput struct {
	Value float64 //接收金额
	//	Address string  //锁定脚本

	PubKeyHash []byte //公钥哈希
}

//给定转账地址，得到这个地址的哈希
//锁定output
func (output *TXOutput) Lock(address string) {
	//address->public key hash
	decodeInfo := base58.Decode(address)

	pubKeyHash := decodeInfo[1 : len(decodeInfo)-4]

	output.PubKeyHash = pubKeyHash

}

//创建新的交易输出 包含转账金额和公钥哈希（“锁定脚本”）
func NewTXOuput(value float64, address string) TXOutput {
	output := TXOutput{Value: value}
	output.Lock(address)
	return output
}

type Transaction struct {
	TXid      []byte     //交易id
	TXInputs  []TXInput  //所有过的input
	TXOutputs []TXOutput //所有的output
}

func (tx *Transaction) SetTXID() {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	hash := sha256.Sum256(buffer.Bytes())
	tx.TXid = hash[:]
}

//挖矿奖励
const reward = 12.5

//实现挖矿交易
//特点：只有输出，没有有效的输入(不要id，索引，签名）
func NewCoinBaseTx(miner string, data string) *Transaction {
	//TODO
	inputs := []TXInput{{nil, -1, nil, []byte(data)}}
	//outputs := []TXOutput{{12.5, miner}}
	output := NewTXOuput(reward, miner)
	outputs := []TXOutput{output}

	tx := Transaction{nil, inputs, outputs}
	tx.SetTXID()

	return &tx
}

func (tx *Transaction) IsCoinbase() bool {
	//1.只有一个input
	//2.引用的id是nil
	//3.引用的索引是-1
	inputs := tx.TXInputs
	if len(inputs) == 1 && inputs[0].TXID == nil && inputs[0].Index == -1 {
		return true
	}
	return false
}

//创建普通交易
func NewTransaction(from string, to string, amount float64, bc *BlockChain) *Transaction {
	//1.打开钱包
	ws := NewWallets()

	wallet := ws.WalletsMap[from]
	if wallet == nil {
		fmt.Printf("%s 的私钥不存在，交易创建失败!\n", from)
		return nil
	}
	//2.获取公钥私钥
	privateKey := wallet.PrivateKey //此时不需要，步骤三签名时需要
	publicKey := wallet.PublicKey
	pubKeyHash := HashPubKey(publicKey)

	//1.遍历账本，找到属于付款人的合适的金额，把这个outputs找到
	utxos := make(map[string][]int64) //标识能用的utxo
	var resValue float64              //这些utxo存储的金额
	utxos, resValue = bc.FindNeedUtxos(pubKeyHash, amount)

	//2.如果钱不足以转账,交易失败
	if resValue < amount {
		fmt.Printf("余额不足，交易失败!\n")
		return nil
	}

	var inputs []TXInput
	var outputs []TXOutput

	//3.否则,将outputs转成inputs
	for txid, indexes := range utxos {
		for _, i := range indexes {
			input := TXInput{[]byte(txid), i, nil, publicKey}
			inputs = append(inputs, input)
		}
	}

	//4.创建输出,创建一个属于收款人的output
	//output := TXOutput{amount, to}
	output := NewTXOuput(amount, to)
	outputs = append(outputs, output)

	//5.如果有找零,创建属于付款人的output
	if resValue > amount {
		//output1 := TXOutput{resValue - amount, from}
		output1 := NewTXOuput(resValue-amount, from)
		outputs = append(outputs, output1)
	}

	//创建交易
	tx := Transaction{nil, inputs, outputs}

	//6.设置交易id
	tx.SetTXID()

	//把查找引用交易的环节放到blockchain中，同时在blockchain进行调用签名
	//付款人在创建交易时，已经得到了所有引用的output的详细信息
	//但是我们不去使用，因为矿工没有这部分信息，矿工进行校验需要遍历账本找到所有引用交易
	//为了统一操作，再进行查询一次，进行签名
	bc.SignTransaction(&tx, privateKey)

	//7.返回交易结构
	return &tx
}

//第一个参数是私钥
//第二个参数是这个交易的input所引用的所有交易
func (tx *Transaction) Sign(privKey *ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	fmt.Printf("对交易进行签名...\n")

	if tx.IsCoinbase() {
		return
	}

	//1.拷贝交易txCopy，做相应裁剪，将每一个input的pubKey和sig设置为nil
	txCopy := tx.TrimmedCopy()

	for i, input := range txCopy.TXInputs {
		//2.遍历txCopy，把这个input所引用的output的公钥哈希复制给pubKey
		prevTX := prevTXs[string(input.TXID)]
		output := prevTX.TXOutputs[input.Index]

		//input.PubKey = output.PubKeyHash
		txCopy.TXInputs[i].PubKey = output.PubKeyHash
		//签名是对数据的hash进行签名
		//我们的额数据都在交易中，我们要求交易的哈希
		//Transaction的SetID函数就是对交易的哈希
		//所以我们可以使用交易id作为我们要签名的内容

		//3.生成要签名的数据（哈希）
		txCopy.SetTXID()
		signData := txCopy.TXid
		//清理数据
		txCopy.TXInputs[i].PubKey = nil

		fmt.Printf("要签名的数据，signData：%x\n", signData)
		//4.对数据进行签名
		r, s, err := ecdsa.Sign(rand.Reader, privKey, signData)
		if err != nil {
			fmt.Printf("交易签名失败! err:%v\n", err)
		}
		//5.拼接r，s为字节流
		signature := append(r.Bytes(), s.Bytes()...)
		//6.赋值给原始交易的Signature字段
		tx.TXInputs[i].Signature = signature
	}
}

//裁剪
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, input := range tx.TXInputs {
		input1 := TXInput{input.TXID, input.Index, nil, nil}
		inputs = append(inputs, input1)
	}
	outputs = tx.TXOutputs

	tx1 := Transaction{tx.TXid, inputs, outputs}
	return tx1
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	fmt.Printf("对交易进行校验...\n")
	//1.拷贝修剪副本
	txCopy := tx.TrimmedCopy()
	//2.遍历原始交易
	for i, input := range tx.TXInputs {
		//3.遍历原始交易的inputs所引用的前交易prevTX
		prevTX := prevTXs[string(input.TXID)]
		output := prevTX.TXOutputs[input.Index]

		//4.找到output的公钥哈希，赋值给txCopy对应的input
		txCopy.TXInputs[i].PubKey = output.PubKeyHash

		//5.还原签名的数据
		txCopy.SetTXID()
		//清理,置空
		txCopy.TXInputs[i].PubKey = nil
		verifyData := txCopy.TXid
		fmt.Printf("verifyData:%x\n", verifyData)

		//6.校验
		//还原签名为r，s
		signature := input.Signature
		//公钥字节流
		pubKeyBytes := input.PubKey

		r := big.Int{}
		s := big.Int{}
		rData := signature[:len(signature)/2]
		sData := signature[len(signature)/2:]
		r.SetBytes(rData)
		s.SetBytes(sData)

		//还原公钥的curve，x，y
		x := big.Int{}
		y := big.Int{}
		xData := pubKeyBytes[:len(pubKeyBytes)/2]
		yData := pubKeyBytes[len(pubKeyBytes)/2:]
		x.SetBytes(xData)
		y.SetBytes(yData)

		curve := elliptic.P256()

		publicKey := ecdsa.PublicKey{curve, &x, &y}
		//公钥，数据，r，s
		//Verify(pub *PublicKey, hash []byte, r, s *big.Int) bool
		if !ecdsa.Verify(&publicKey, verifyData, &r, &s) {
			return false
		}
	}
	return true
}

func (tx *Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction: %x", tx.TXid))

	for i, input := range tx.TXInputs {
		lines = append(lines, fmt.Sprintf("      Input: %d", i))
		lines = append(lines, fmt.Sprintf("         TXID:      %x", input.TXID))
		lines = append(lines, fmt.Sprintf("         Out:       %d", input.Index))
		lines = append(lines, fmt.Sprintf("         Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("         PubKey:    %x", input.PubKey))
	}
	for i, output := range tx.TXOutputs {
		lines = append(lines, fmt.Sprintf("      Output: %d", i))
		lines = append(lines, fmt.Sprintf("         Value:     %f", output.Value))
		lines = append(lines, fmt.Sprintf("         Script:    %x", output.PubKeyHash))
	}
	return strings.Join(lines, "\n")

}

