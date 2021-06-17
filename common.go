package main

import (
	"fmt"
	"time"
)

func ProcessArgs(args []string) (int, string) {
	if len(args) == 2 {
		return 1, args[1]
	}
	return 0, ""
}

func DebugPrint(pc interface{}) {
	t := time.Now().Format("2006/01/02 15:04:05")
	if DW {
		fmt.Println(t, pc)
	}
}
