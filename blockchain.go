package main

import (
	"./base58"
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

//创建区块链
//type BlockChain struct {
//	Blocks []*Block
//}
//
////实现创建区块链的方法
//func NewBlochChain() *BlockChain {
//	//在创建的时候添加一个区块：创世块
//	genesisBlock := NewBlock(genesisInfo, []byte{0x0000000000000000})
//
//	bc := BlockChain{Blocks: []*Block{genesisBlock}}
//	return &bc
//}
//
////添加区块
//func (bc *BlockChain) AddBlock(data string) {
////	//创建一个区块
////	//找到区块链中最后一个区块的哈希值
////	//此处哈希值不需要自己添加，只需要传输即可
////	lastBlock := bc.Blocks[len(bc.Blocks)-1]
////	prevHash := lastBlock.Hash
////	//找到上一个区块哈希值后，将要打包的数据打包成一个区块
////	block := NewBlock(data, prevHash)
////	//将打包好的区块加入区块链
////	bc.Blocks = append(bc.Blocks, block)
////}

//改写后的创建区块
// v3版本
type BlockChain struct {
	db   *bolt.DB //句柄
	tail []byte   //最后一个区块的哈希值
}

const blockChainName = "blockChain.db"
const blockBucketName = "blockbucket"
const lashHashKey = "lastHashKey"

//创建区块链
func CreateBlockChain(miner string) *BlockChain {

	if IsFileExist(blockChainName) {
		fmt.Print("区块链已经存在，不需要重复创建！\n")
		return nil
	}

	db, err := bolt.Open(blockChainName, 0600, nil) //0600可读可写
	if err != nil {
		log.Panic()
	}

	//	defer db.Close()

	var tail []byte

	db.Update(func(tx *bolt.Tx) error {

		b, err := tx.CreateBucket([]byte(blockBucketName))
		if err != nil {
			log.Panic(err)
		}

		//创建创世块
		//创世块中只有一个挖矿交易，只有CoinBase

		coinbase := NewCoinBaseTx(miner, genesisInfo)
		//genesisBlock := NewBlock(genesisInfo, []byte{})
		genesisBlock := NewBlock([]*Transaction{coinbase}, []byte{})

		b.Put(genesisBlock.Hash, genesisBlock.Serialize())
		b.Put([]byte(lashHashKey), genesisBlock.Hash)

		//只是为了检验 才打印出来
		//blockInfo:=b.Get(genesisBlock.Hash)
		//block:=Deserialize(blockInfo)
		//fmt.Printf("解码后的block数据:%s\n",block)

		tail = genesisBlock.Hash

		return nil
	})

	return &BlockChain{db, tail}
}

//返回区块链实例
func NewBlockChain() *BlockChain {
	if !IsFileExist(blockChainName) {
		fmt.Print("区块链不存在，请先创建！\n")
		return nil
	}
	//1.获得数据库的句柄，打开数据库，读写数据
	// 判断是否有bucket，如果没有，就创建bucket
	// 写入创世块
	// 写入lastHashKey
	// 更新tail为最后一个区块的哈希
	// 返回bc实例
	db, err := bolt.Open(blockChainName, 0600, nil) //0600可读可写
	if err != nil {
		log.Panic()
	}

	var tail []byte

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucketName))

		if b == nil {
			fmt.Printf("区块链Bucket为空，请检查！\n")
			os.Exit(1)
		}
		tail = b.Get([]byte(lashHashKey))

		return nil
	})
	return &BlockChain{db, tail}
}

//添加区块
func (bc *BlockChain) AddBlock(txs []*Transaction) {
	//矿工得到交易，第一时间对交易进行验证
	validTXs := []*Transaction{}
	for _, tx := range txs {
		if bc.VerifyTransaction(tx) {
			fmt.Printf("---- 交易有效 :%x\n",tx.TXid)
			validTXs = append(validTXs, tx)
		} else {
			fmt.Printf("发现无效交易：%x\n", tx.TXid)
		}

	}

	bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucketName))

		if b == nil {
			//说明bucket不存在，就退出
			fmt.Printf("bucket不存在，请检查\n")
			os.Exit(1)
		}

		//创建创世块
		block := NewBlock(validTXs, bc.tail)
		b.Put(block.Hash, block.Serialize())
		b.Put([]byte(lashHashKey), block.Hash) //更新最后一个哈希

		bc.tail = block.Hash //更新尾值
		return nil
	})
}

//迭代器作用：之前创建的区块保存，添加的新的区块直接加到后面，实现不断挖矿的功能
//定义一个区块链的迭代器，包含db，current
type BlockChainIterator struct {
	db      *bolt.DB
	current []byte //当前所指向区块的哈希值
}

//创建迭代器，使用bc进行初始化
func (bc *BlockChain) NewIterator() *BlockChainIterator {
	return &BlockChainIterator{bc.db, bc.tail}
}

func (it *BlockChainIterator) Next() *Block {

	var block Block

	it.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucketName))
		if b == nil {
			fmt.Printf("bucket不存在，请检查\n")
			os.Exit(1)
		}
		//获取当前区块的信息
		blockInfo := b.Get(it.current)
		//反序列化
		block = *Deserialize(blockInfo)
		//将current指向前一个区块的哈希值
		it.current = block.PrevBlockHash

		return nil
	})
	return &block
}

//优化FindMyUtxos和FindNeedUtxos两个函数
//
//1.FindMyUtxos：找到所有utxo（只要output）
//2.FindNeedUtxos：找到需要的utxo（只要output的定位）

//定义已经包含output的定位信息的结构
type UTXOInfo struct {
	TXID   []byte
	index  int64
	Output TXOutput
}

func (bc *BlockChain) FindMyUtxos(pubKeyHash []byte /*address string*/) []UTXOInfo /*[]TXOutput*/ {
	fmt.Printf("FindMyUtxos\n")
	//var UTXOs []TXOutput
	var UTXOInfos []UTXOInfo

	it := bc.NewIterator()
	//标识已经消耗过的utxo的结构，key是id，value是这个id里面的output索引的数组
	spentUTXOs := make(map[string][]int64)

	//1.遍历账本
	for {
		block := it.Next()
		//2.遍历交易
		for _, tx := range block.Transactions {
			//不是coinbase就遍历Inputs
			if tx.IsCoinbase() == false {
				for _, input := range tx.TXInputs {
					//判断当前被使用input是否为目标地址
					if bytes.Equal(HashPubKey(input.PubKey), pubKeyHash) {
						fmt.Printf("找到一个消耗过的output！index：%d\n", input.Index)
						key := string(input.TXID)
						spentUTXOs[key] = append(spentUTXOs[key], input.Index)
						//spentUTXOS[0x333]=[]int64{0,1}
						//spentUTXOS[0x222]=[]int64{0}
					}
				}
			}

		OUTPUT:
			//3.遍历output
			for i, output := range tx.TXOutputs {
				key := string(tx.TXid)
				indexes := spentUTXOs[key]
				if len(indexes) != 0 {
					fmt.Printf("当前这笔交易中有别消耗过的output！\n")
					for _, j := range indexes {
						if int64(i) == j {
							fmt.Printf("i == j,当前的output已经被消耗过，跳过不统计!\n")
							continue OUTPUT
						}
					}
				}

				//4.找到属于我的所有output
				if bytes.Equal(pubKeyHash, output.PubKeyHash) {
					//fmt.Printf("找到属于 %s 的output，i:%d\n", address, i)
					//UTXOs = append(UTXOs, output)
					utxoinfos := UTXOInfo{tx.TXid, int64(i), output}
					UTXOInfos = append(UTXOInfos, utxoinfos)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			fmt.Printf("遍历区块结束!\n")
			break
		}
	}
	//return UTXOs
	return UTXOInfos
}

func (bc *BlockChain) GetBalance(address string) {
	//此过程不要打开钱包，因为有可能查看余额的人不是地址本人
	decodeInfo := base58.Decode(address)

	pubKeyHash := decodeInfo[1 : len(decodeInfo)-4]

	utxoinfos := bc.FindMyUtxos(pubKeyHash)

	var total = 0.0
	for _, utxoinfo := range utxoinfos {
		//total += utxo.Value
		total += utxoinfo.Output.Value
	}

	fmt.Printf("%s 的余额为：%f\n", address, total)
}

func (bc *BlockChain) FindNeedUtxos(pubKeyHash []byte, amount float64) (map[string][]int64, float64) {
	needUtxos := make(map[string][]int64)

	var resValue float64

	utxoinfos := bc.FindMyUtxos(pubKeyHash)

	for _, utxoinfo := range utxoinfos {
		key := string(utxoinfo.TXID)
		needUtxos[key] = append(needUtxos[key], int64(utxoinfo.index))
		resValue += utxoinfo.Output.Value
		//2.判断金额是否满足
		if resValue >= amount {
			//a.满足，直接返回
			break
		}
	}
	return needUtxos, resValue
}

//交易签名
func (bc *BlockChain) SignTransaction(tx *Transaction, privateKey *ecdsa.PrivateKey) {
	//1.遍历账本
	prevTXs := make(map[string]Transaction)

	for _, input := range tx.TXInputs {
		prevTX := bc.FindTransaction(input.TXID)
		if prevTX == nil {
			fmt.Printf("没有找到交易:%x\n", input.TXID)
		} else {
			prevTXs[string(input.TXID)] = *prevTX
		}
	}

	tx.Sign(privateKey, prevTXs)
}

//矿工校验交易签名
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	//如果是挖矿交易，直接返回true
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)
	//1.找到交易input所引用的所有交易prevTXs
	for _, input := range tx.TXInputs {
		prevTX := bc.FindTransaction(input.TXID)
		if prevTX == nil {
			fmt.Printf("没有找到交易:%x\n", input.TXID)
		} else {
			prevTXs[string(input.TXID)] = *prevTX
		}
	}
	//2.对交易进行校验
	return tx.Verify(prevTXs)
}

//寻找input对应的output
func (bc *BlockChain) FindTransaction(txid []byte) *Transaction {
	it := bc.NewIterator()

	for {
		block := it.Next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.TXid, txid) {
				fmt.Printf("找到了所引用的交易 ：%x\n", txid)
				return tx
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return nil
}

//
//	//1.遍历账本
//	for {
//		block := it.Next()
//		//2.遍历交易
//		for _, tx := range block.Transactions {
//			//遍历Inputs
//			for _, input := range tx.TXInputs {
//				if input.Address == from {
//					fmt.Printf("找到一个消耗过的output！index：%d\n", input.Index)
//					key := string(input.TXID)
//					spentUTXOs[key] = append(spentUTXOs[key], input.Index)
//					//spentUTXOS[0x333]=[]int64{0,1}
//					//spentUTXOS[0x222]=[]int64{0}
//				}
//			}
//
//		OUTPUT:
//			//3.遍历output
//			for i, output := range tx.TXOutputs {
//				key := string(tx.TXid)
//				indexes := spentUTXOs[key]
//				if len(indexes) != 0 {
//					fmt.Printf("当前这笔交易中有别消耗过的output！\n")
//					for _, j := range indexes {
//						if int64(i) == j {
//							fmt.Printf("i == j,当前的output已经被消耗过，跳过不统计!\n")
//							continue OUTPUT
//						}
//					}
//				}
//
//				//4.找到属于我的所有output
//				if from == output.Address {
//					fmt.Printf("找到属于 %s 的output，i:%d\n", from, i)
//					//UTXOs = append(UTXOs, output)
//					//找到符合条件的output
//					//1.添加到返回结构中的needUtxos
//					needUtxos[key] = append(needUtxos[key], int64(i))
//					resValue += output.Value
//					//2.判断金额是否满足
//					if resValue >= amount {
//						//a.满足，直接返回
//						return needUtxos, resValue
//					}
//					//b.不满足，继续遍历
//				}
//			}
//		}
//
//		if len(block.PrevBlockHash) == 0 {
//			fmt.Printf("遍历区块结束!\n")
//			break
//		}
//	}
//	//+++++++++++++++++++++++++++++++++++++++
//
//	return needUtxos, resValue
//}
