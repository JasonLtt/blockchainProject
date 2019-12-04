package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

func main() {
	//创建数据库
	db, err := bolt.Open("test.db", 0600, nil)//0600可读可写
	if err != nil {
		log.Panic()
	}

	defer db.Close()

	//创建bucket
	db.Update(func(tx *bolt.Tx) error {
		b1 := tx.Bucket([]byte("bucket1"))

		if b1 == nil {
			//说明bucket1不存在，需要创建一个
			b1, err = tx.CreateBucket([]byte("bucket1"))
			if err != nil {
				log.Panic(err)
			}
		}

		//bucket创建完成
		//写入数据Put，读取数据Get

		//写入数据
		err = b1.Put([]byte("name1"), []byte("Lily"))
		if err != nil {
			fmt.Printf("写入数据失败name1:Lily\n")
		}

		err = b1.Put([]byte("name2"), []byte("Jim"))
		if err != nil {
			fmt.Printf("写入数据失败name2:Jim\n")
		}

		//读取数据
		name1 := b1.Get([]byte("name1"))
		name2 := b1.Get([]byte("name2"))
		name3 := b1.Get([]byte("name3"))

		//打印数据
		fmt.Printf("name1:%s\n", name1)
		fmt.Printf("name2:%s\n", name2)
		fmt.Printf("name3:%s\n", name3)

		return nil
	})
}
