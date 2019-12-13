package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

//Wallets结构
//把地址和密钥对对应起来
//map[address]->walletKeyPair

type Wallets struct {
	WalletsMap map[string]*WalletKeyPair
}

//创建wallets，返回wallets实例
func NewWallets(node_ID string) (*Wallets, error) {
	var ws Wallets

	ws.WalletsMap = make(map[string]*WalletKeyPair)
	//1.把所有的钱包从本地加载出来
	err := ws.LoadFromFile(node_ID)

	//2.把实例返回

	return &ws, err
}

//Wallets对外   WalletKeyPair对内
//Wallets调用WalletKeyPair
const WalletName = "wallet_%s.dat"

func (ws *Wallets) CreateWallet(node_ID string) string /**WalletKeyPair*/ {
	//调用NewWalletKeyPair
	wallet := NewWalletKeyPair()
	//将返回的walleykeypair添加到walletmap中
	address := wallet.GetAddress()

	ws.WalletsMap[address] = wallet
	//保存到本地中
	res := ws.SaveToFile(node_ID)
	if res == false {
		fmt.Printf("创建钱包失败！\n")
		return ""
	}
	return address
}

//存入node_ID的钱包
func (ws *Wallets) SaveToFile(node_ID string) bool {
	var buffer bytes.Buffer
	WalletName := fmt.Sprintf(WalletName, node_ID)
	//将接口类型明确注册一下，否则gob编码失败
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(ws)
	if err != nil {
		fmt.Printf("钱包序列化失败,err:%v\n", err)
		return false
	}

	content := buffer.Bytes()
	//打开node_ID的钱包
	//WriteFile(filename string, data []byte, perm os.FileMode) error
	//0644表示：创建了一个普通文件，文件所有者对该文件有读写权限，用户组和其他人只有读权限，都没有执行权限
	err = ioutil.WriteFile(WalletName, content, 0644)
	if err != nil {
		fmt.Printf("钱包创建失败！\n")
		return false
	}
	return true
}

func (ws *Wallets) LoadFromFile(node_ID string) error {
	WalletName := fmt.Sprintf(WalletName, node_ID)
	//判断文件是否存在
	if _, err := os.Stat(WalletName); !IsFileExist(WalletName) {
		fmt.Printf("钱包文件不存在，准备创建!\n")
		return err
	}
	//读取文件
	//ReadFile(filename string) ([]byte, error)
	content, err := ioutil.ReadFile(WalletName)
	if err != nil {
		log.Panic(err)
		//return err
	}
	//注册接口
	gob.Register(elliptic.P256())

	//gob解码
	decoder := gob.NewDecoder(bytes.NewReader(content))
	var wallets Wallets
	err = decoder.Decode(&wallets)

	if err != nil {
		fmt.Printf("Error:%v\n", err)
		//return false
	}
	//赋值给ws
	ws.WalletsMap = wallets.WalletsMap
	return nil
}

func (ws *Wallets) ListAddress() []string {
	//遍历ws.WalletsMap结构返回key
	var addresses []string

	for address, _ := range ws.WalletsMap {
		addresses = append(addresses, address)
	}
	return addresses
}

//通过地址获得钱包
func (ws *Wallets) GetWallet(addr string) WalletKeyPair {
	return *ws.WalletsMap[addr]
}
