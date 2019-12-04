package main

import (
	"fmt"
	"os"
)

func main(){
	cmds:=os.Args

	for i,cmd:=range cmds{
		fmt.Printf("cmd[%d]:%s\n",i,cmd)
	}
}
