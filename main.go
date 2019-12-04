package main

// v1版本
// 简单区块字段：prevehash，hash，data（使用数据流代替交易）
// 添加区块
// 打印区块
//
// v2版本
// 区块字段全部填上
// 加入工作量证明pow
//   NewProofOfWork()
// 提供计算不断计算hash的哈数
//   Run（）
// 验证随机值
// 提供校验函数
//   IsValid（）
//
// v3版本
// bolt数据库
// 使用数据库改写区块链结构
// 添加命令行：创建区块链、添加区块、打印区块
//
// v4版本
// 引入UTXO
// 添加命令行：getBalance（余额），CreateBlockChain（创建区块链），send（转账）
//
// v5版本
// 引入签名
//1.创建密钥对->公钥->地址
//  椭圆曲线算法生成私钥,私钥生成公钥,公钥成功地址,wallet.dat保存所有公钥、私钥、地址
//2.使用地址、公钥私钥改写代码
//3.交易签名校验

func main() {
	//block := NewBlock(genesisInfo, []byte{0x0000000000000000}) //创世块
	//bc := NewBlockChain("班长")
	//defer bc.db.Close()
	cli := CLI{}
	cli.Run()
}

	//bc.AddBlock("hahhahahahha")
	//
	//it := bc.NewIterator()
	//
	//for {
	//	block := it.Next()
	//	fmt.Printf("+++++++++++++++++++++++++++++++++++++++++\n")
	//	fmt.Printf("Version:%d\n", block.Version)
	//	fmt.Printf("PrevBlockHash:%x\n", block.PrevBlockHash)
	//	fmt.Printf("MerKleRoot:%x\n", block.MerKleRoot)
	//	//更改时间格式
	//	TimeFormat := time.Unix(int64(block.TimeStamp), 0).Format("2006-01-02 15:04:05")
	//	fmt.Printf("TimeStamp:%s\n", TimeFormat)
	//
	//	fmt.Printf("Difficulty:%d\n", block.Difficulty)
	//	fmt.Printf("Nonce:%d\n", block.Nonce)
	//	fmt.Printf("Hash:%x\n", block.Hash)
	//	fmt.Printf("Data:%s\n", block.Data)
	//
	//	pow := NewProofOfWork(block)
	//	fmt.Printf("IsValid:%v\n", pow.IsValid())
	//
	//	if bytes.Equal(block.PrevBlockHash, []byte{}) {
	//		fmt.Printf("区块链遍历结束\n")
	//		break
	//	}
	//}

	//简单版本
	//bc.AddBlock("哈哈哈哈哈")
	//bc.AddBlock("他来了他来了")
	//for i, block := range bc.Blocks {
	//	fmt.Printf("++++++++++++ %d ++++++++++++\n", i)
	//	fmt.Printf("Version:%d\n", block.Version)
	//	fmt.Printf("PrevBlockHash:%x\n", block.PrevBlockHash)
	//	fmt.Printf("MerKleRoot:%x\n", block.MerKleRoot)
	//	//更改时间格式
	//	TimeFormat := time.Unix(int64(block.TimeStamp), 0).Format("2006-01-02 15:04:05")
	//	fmt.Printf("TimeStamp:%s\n", TimeFormat)
	//
	//	fmt.Printf("Difficulty:%d\n", block.Difficulty)
	//	fmt.Printf("Nonce:%d\n", block.Nonce)
	//	fmt.Printf("Hash:%x\n", block.Hash)
	//	fmt.Printf("Data:%s\n", block.Data)
	//
	//	pow := NewProofOfWork(block)
	//	fmt.Printf("IsValid:%v\n", pow.IsValid())
	//}
//}
