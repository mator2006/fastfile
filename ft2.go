package main

import (
	"fmt"
	"os"
	"sync"
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
	var vg sync.WaitGroup

	vg.Add(1)
	go func() {
		defer vg.Done()
		rfscer.Getsignalr()
	}()

	vg.Wait()

	for i := 0; i <= FailTryCount-1; i++ {
		if rfscer.SendReadySignials(i) {
			break
		}
		if i == FailTryCount-1 {
			DebugPrint(fmt.Errorf("Can not Send reday signal"))
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	rfscer.Getdatas()
	rfscer.WriteFile()
}

func Send(fn string) {
	var fscer fsc
	var vg sync.WaitGroup
	if fn == "" {
		DebugPrint(fmt.Errorf("file name is wrong"))
		return
	}
	fscer.fn = fn
	fscer.Initfsc()
	vg.Add(1)
	go func() {
		defer vg.Done()
		fscer.Getsignals()
	}()
	for i := 0; i <= FailTryCount-1; i++ {
		if fscer.Sendfsc(i) {
			break
		}
		if i == FailTryCount-1 {
			DebugPrint(fmt.Errorf("Can not Send file send control infomation"))
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	vg.Wait()
	fscer.Senddatas()
}
