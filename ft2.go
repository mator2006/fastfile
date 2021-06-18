package main

import (
	"fmt"
	"os"
	"time"
)

var DW bool = true

func main() {
	Doselect, fn := ProcessArgs(os.Args)
	if Doselect == 0 {
		Receive()
	}
	if Doselect == 1 {
		Send(fn)
	}
}

func Receive() {
	var rfscer fsc

	rfscer.Getsignalr()

	rfscer.WriteFile()
}

func Send(fn string) {
	var fscer fsc

	if fn == "" {
		DebugPrint(fmt.Errorf("File name is wrong !"))
		return
	}
	fscer.fn = fn
	fscer.Initfsc()

	var s int
	for {
		if fscer.Sendfsc() {
			break
		}
		if s == FailTryCount-1 {
			DebugPrint(fmt.Errorf("Can not connection server !"))
			return
		}
		s++
		time.Sleep(ListeningWaitTime)
	}

	fscer.Senddatas()
}
