package main

import (
	"bytes"
	"fmt"
	"strings"
)

func main() {
	strA := []string{"hello", "world", "itcast"}

	strRes := strings.Join(strA, "=")

	fmt.Printf("strRes:%s\n", strRes)
	joinRes := bytes.Join([][]byte{[]byte("hello"), []byte("world"), []byte("itcast")}, []byte("="))
	fmt.Printf("JoinRes:%s\n", joinRes)
}
