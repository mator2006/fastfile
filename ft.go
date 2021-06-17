package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

var DW bool = true

func main() {

	// i, f := ProcessArgs(os.Args)
	if len(os.Args) == 2 {
		Send(os.Args[1])
	}
	if len(os.Args) == 1 {
		Receive()
	}
}

func Send(f string) {
	var ffer ff
	ffer.fsc.fn = f

	err := ffer.GetFileS()
	if err != nil {
		DebugPrint(err)
	}
	SendIP := SIP + ":" + strconv.Itoa(PortNumber)
	fb := (fmt.Sprintf("%s\n%d\n%d\n%d\n%d\nEOF\n", ffer.fsc.fn, ffer.fsc.fs, ffer.fsc.ps, ffer.fsc.pi, ffer.fsc.tc))
	conn, err := net.Dial("tcp", SendIP)
	if err != nil {
		DebugPrint(err)
		return
	}
	defer conn.Close()
	_, err = conn.Write([]byte(fb))
	if err != nil {
		DebugPrint(err)
	}
	// fmt.Printf("发送了%d个字节\n", n)

	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			DebugPrint(err) //出错退出
			return
		}
		if bytes.Contains(buf[:n], []byte("OK")) {
			break
		} else {
			DebugPrint(string(buf[:n]))
		}
	}
	ffer.s()
}

func (ffer *ff) s() {
	var vg sync.WaitGroup
	fb, err := ioutil.ReadFile(ffer.fsc.fn)
	if err != nil {
		DebugPrint(err)
		return
	}
	for i := range ffer.fsb.index {
		vg.Add(1)
		go func(nv int) {
			defer vg.Done()
			SendIP := SIP + ":" + strconv.Itoa(1024*ffer.fsc.ps+ffer.fsc.pi+nv)
			conn, err := net.Dial("tcp", SendIP)
			if err != nil {
				DebugPrint(err)
				return
			}
			defer conn.Close()

			n, err := conn.Write(fb[ffer.fsb.start[nv]:ffer.fsb.stop[nv]])
			if err != nil {
				DebugPrint(err)
			}

			_ = n
			// ffer.fsb.index[nv] = n

		}(i)
	}
	vg.Wait()
	DebugPrint(fmt.Sprintf("%+v", ffer.fsb.index))
}

func Receive() {
	var Refscer ff
	var Refsc []byte
	DebugPrint("------服务器开始监听------")
	ListenIP := LIP + ":" + strconv.Itoa(PortNumber)
	listen, err := net.Listen("tcp", ListenIP)
	if err != nil {
		DebugPrint(err)
		return
	}
	defer listen.Close() //延时关闭listen
	DebugPrint("-----等待客户连接----")
	conn, err := listen.Accept()
	if err != nil {
		DebugPrint(err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		DebugPrint(err) //出错退出
		return
	}
	if bytes.Contains(buf[:n], []byte("EOF")) {
		Refsc = buf[:n]
		Refscer = ProcessReceivefsc(Refsc)
		//fmt.Printf("[%+v]\n", Refscer.fsc)
		DebugPrint(fmt.Sprintf("------已接收到指令，开启数据通道------\n%+v", Refscer.fsc))

		var vg sync.WaitGroup

		for i := 0; i <= Refscer.fsc.tc-1; i++ {
			vg.Add(1)
			go func(nv int) {
				defer vg.Done()
				// DebugPrint(fmt.Sprintf("[%d]\t", 1024*Refscer.fsc.ps+Refscer.fsc.pi+nv))
				Refscer.r(nv)
			}(i)
		}

		time.Sleep(1 * time.Second)

		vg.Add(1)
		go func() {
			defer vg.Done()
			_, err = conn.Write([]byte("OK"))
			if err != nil {
				DebugPrint(err)
			}
			DebugPrint("------数据通道已经开启，开始传输------")
		}()

		vg.Wait()

		DebugPrint(fmt.Sprintf("通道开启/有效接收:[%d/%d]", Refscer.fsc.tc, len(Refscer.fsb.body)))

		if len(Refscer.fsb.body) == Refscer.fsc.tc {
			DebugPrint("------传输成功------")
			var fb bytes.Buffer
			for i := range Refscer.fsb.index {
				// fmt.Printf("%d\t%d\n", i, len(v))
				fb.Write(Refscer.fsb.body[i])
			}
			if len(fb.Bytes()) == Refscer.fsc.fs {
				DebugPrint(fmt.Sprintf("------文件校验成功，文件大小:[%d]------", Refscer.fsc.fs))
				err = ioutil.WriteFile(Refscer.fsc.fn, fb.Bytes(), 0644)
				if err != nil {
					err = fmt.Errorf("文件写入失败")
					DebugPrint(err)
					return
				} else {
					DebugPrint(fmt.Sprintf("------[%s]文件写入成功------", Refscer.fsc.fn))
					return
				}
			} else {
				err = fmt.Errorf("文件校验失败")
				DebugPrint(err)
				return
			}
		} else {
			if len(Refscer.fsb.index) != len(Refscer.fsb.body) {
				err = fmt.Errorf("接收数据不符index/body:[%d/%d]", len(Refscer.fsb.index), len(Refscer.fsb.body))
				DebugPrint(err)
			}
			DebugPrint("------传输失败------")
			return
		}
	}
}

func (ffer *ff) r(nv int) {

	ListenIP := LIP + ":" + strconv.Itoa(1024*ffer.fsc.ps+ffer.fsc.pi+nv)
	listen, err := net.Listen("tcp", ListenIP)
	if err != nil {
		DebugPrint(err)
		return
	}
	defer listen.Close()

	conn, err := listen.Accept()
	if err != nil {
		DebugPrint(err)
	}
	defer conn.Close()

	buf := make([]byte, fileslice)
	var n int
	for {
		n, err = conn.Read(buf)
		if err != nil {
			err = fmt.Errorf("%v error:%v\tn=%d\tbuf:%v:%v", listen.Addr().String(), err, n, buf[:4], buf[len(buf)-4:])
			DebugPrint(err)
			break
		}
		// if err != nil {
		// 	fmt.Println(string(buf[:10]))
		// 	for i := 0; i <= 2; i++ {
		// 		nn, _ := conn.Read(buf)
		// 		// fmt.Println(string(buf[:10]))
		// 		if nn == 0 {
		// 			continue
		// 		}
		// 		if nn == fileslice || nn == ffer.fsc.fs%fileslice {
		// 			n = nn
		// 			break
		// 		}
		// 	}
		// 	if n == 0 {
		// 		err = fmt.Errorf("%s接收出错,数据长度:%d,错误:%v", ListenIP, n, err)
		// 		DebugPrint(err) //出错退出
		// 		return
		// 	}
		// }
		if n == fileslice || n == ffer.fsc.fs%fileslice {
			fmt.Printf("%v OK!\tn=%d\tbuf:%v:%v\n", listen.Addr().String(), n, buf[:4], buf[len(buf)-4:])
			//DebugPrint(fmt.Sprintf("%s接收成功,%d", ListenIP, n))
			break
		}
	}

	ffer.fsb.index = append(ffer.fsb.index, nv)
	ffer.fsb.body = append(ffer.fsb.body, buf[:n])
}
