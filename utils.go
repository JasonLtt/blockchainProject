package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
)

func uintToByte(num uint64) []byte {
	//TODO
	//使用binary.Write来进行编码
	var buffer bytes.Buffer
	//编码一定要进行错误检查，非null则错误
	err := binary.Write(&buffer, binary.BigEndian, num) //binary.BigEndian为大端对齐
	if err != nil {
		log.Panic()
	}
	return buffer.Bytes()
}

//判断文件是否存在
func IsFileExist(fileName string) bool {
	_,err:=os.Stat(fileName)
	if os.IsNotExist(err){
		return false
	}
	return true
}
