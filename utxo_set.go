package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

const utxoBucket = "chainstate"

type UTXOset struct {
	bc *BlockChain
}

//提供某一个地址，找到这个地址下所有utxos
//在 func (bc *BlockChain) GetBalance(address string) 中调用
func (u UTXOset) FindMyUtxos(pubKeyHash []byte /*address string*/) []UTXOInfo /*[]TXOutput*/ {
	fmt.Printf("FindMyUtxos\n")
	//var UTXOs []TXOutput
	var UTXOInfos []UTXOInfo

	it := u.bc.NewIterator()
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

func (u UTXOset) FindNeedUtxos(pubKeyHash []byte, amount float64) (map[string][]int64, float64) {
	needUtxos := make(map[string][]int64)

	var resValue float64

	utxoinfos := u.FindMyUtxos(pubKeyHash)

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

//这是寻找这个区块链的所有utxos
//构造一个utxos池
func (u UTXOset) Reindex() {
	db := u.bc.db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	UTXO := u.bc.FindUtxos()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}

//通过传入的block中transaction更新utxoset
//该block为区块链中链尾的区块
func (u UTXOset) UTXOsUpdata(block *Block) {
	db := u.bc.db

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				for _, input := range tx.TXInputs {
					//updataOuts := TXOutputs{}
					updataOuts := Transaction{}
					outsBytes := b.Get(input.TXID)
					outs := DeserializeTransaction(outsBytes)

					for outIdx, output := range outs.TXOutputs {
						if int64(outIdx) != input.Index {
							updataOuts.TXOutputs = append(updataOuts.TXOutputs, output)
						}
					}
					if len(updataOuts.TXOutputs) == 0 {
						err := b.Delete(input.TXID)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := b.Put(input.TXID, updataOuts.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}
				}
			}
			newOutputs := Transaction{}
			for _, out := range tx.TXOutputs {
				newOutputs.TXOutputs = append(newOutputs.TXOutputs, out)
			}

			err := b.Put(tx.TXid, newOutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

//计算utxo池中有多少交易
func (u *UTXOset) countTransaction() int {
	db := u.bc.db
	counter := 0

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return counter
}
