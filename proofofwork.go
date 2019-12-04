package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

//定义一个工作量证明的结构
//a. block
//b. 目标值
// 2.提供创建pow的函数
// NewProofOfWork(参数)
// 3.提供计算 不断计算hash的hash
// Run()
// 4.提供一个校验函数

type ProofOfWork struct {
	block *Block
	//来存储哈希值
	//SetBytes:把bytes转成big.int类型
	//SetString:把string转成
	target *big.Int //系统提供，固定的
}

const Bits = 16

func NewProofOfWork(block *Block) *ProofOfWork {
	pow := ProofOfWork{
		block: block,
	}
	//写难度值，难度值应该是推导出来的，但是我们为了简化，把难度值先写成固定的，一切完成之后，再去推导
	// 0000100000000000000000000000000000000000000000000000000000000000

	//固定难度值
	//targetStr := "0001000000000000000000000000000000000000000000000000000000000000"
	//var bigIntTmp big.Int
	//bigIntTmp.SetString(targetStr, 16)
	//pow.target = &bigIntTmp

	//变动的难度值，推导
	//  0001000000000000000000000000000000000000000000000000000000000000
	//初始化
	//  0000000000000000000000000000000000000000000000000000000000000001
	//向左移动，256位
	//1 0000000000000000000000000000000000000000000000000000000000000000
	//向右移动，四次，一个16进制位代表4个2进制位
	//1 0001000000000000000000000000000000000000000000000000000000000000

	bigIntTmp := big.NewInt(1)
	//bigIntTmp.Lsh(bigIntTmp,256)
	//bigIntTmp.Rsh(bigIntTmp,16)//前面3个零
	bigIntTmp.Lsh(bigIntTmp, 256-Bits)
	pow.target = bigIntTmp

	return &pow
}

//pow的运算函数，为了获取挖矿的随机数，并且获取区块的哈希值
func (pow *ProofOfWork) Run() ([]byte, uint64) {
	//1.获取block数据
	//2.拼接nonce
	// 3.sha256
	// 4.与难度值比较
	// a.哈希值大于难度值，继续，nonce++
	// b.哈希值小于难度值，成功，退出

	var nonce uint64 //定义随机值

	var hash [32]byte //定义hash

	for {
		fmt.Printf("%x\r", hash)                     //打印hash计算的过程
		hash = sha256.Sum256(pow.prepareData(nonce)) //将拼接的block和nonce做一次hash
		var bigIntTmp big.Int
		bigIntTmp.SetBytes(hash[:]) //将hash值转成和难度值一个类型的值
		//   -1 if x <  y
		//    0 if x == y
		//   +1 if x >  y
		//
		//func (x *Int) Cmp(y *Int) (r int)
		if bigIntTmp.Cmp(pow.target) == -1 {
			fmt.Printf("挖矿成功!nonce:%d,哈希值:%x\n", nonce, hash)
			break
		} else {
			nonce++
		}
	}
	return hash[:], nonce
}

//拼接block和nonce
func (pow *ProofOfWork) prepareData(nonce uint64) []byte {
	block := pow.block
	tmp := [][]byte{
		uintToByte(block.Version),
		block.PrevBlockHash,
		block.MerKleRoot,
		uintToByte(block.TimeStamp),
		uintToByte(block.Difficulty),
		uintToByte(nonce),
	}

	//真实的比特币中，是对区块头进行hash运算，不对整个区块进行hash运算

	data := bytes.Join(tmp, []byte{})
	return data
}

func (pow *ProofOfWork) IsValid() bool {
	//将数据与nonce拼接，然后计算hash值
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	//改变格式
	var tmp big.Int
	tmp.SetBytes(hash[:])
	//将计算的hash值与难度值进行比较，若小于难度值，则返回true
	return tmp.Cmp(pow.target) == -1
}
