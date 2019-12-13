package main

import (
	"bufio"
	"fmt"
	"os"
)

const WalletName="wallet.dat"

func main(){


	file, err := os.Open(WalletName)
	if err!= nil {
		fmt.Println("failed to open")
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for  {
		str, err := reader.ReadString('\n') //每次读取一行
		if err!= nil {
			break // 读完或发生错误
		}
		fmt.Printf(str)
	}

}
