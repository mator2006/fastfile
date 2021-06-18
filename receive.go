package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (r *fsc) Getsignalr() {
	ListenIP := SIP + ":" + strconv.Itoa(PortNumber)
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

	buf := make([]byte, SignalByteSize)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			DebugPrint(err)
		}
		if bytes.Contains(buf[:n], []byte("EOF")) {
			DebugPrint("Get file transmission contron signal")
			break
		}
		time.Sleep(LoopWaitTime)
	}
	r.Processfscdata(buf)

	ch := make(chan string, r.tc*3)

	go func() {
		r.Getdatas(ch)
	}()

	for {
		if len(ch) == r.tc {
			DebugPrint("全部通道开启完成，通知对方开始发送")
			_, err = conn.Write([]byte("OK"))
			if err != nil {
				DebugPrint(err)
				return
			}
			break
		}
		// DebugPrint(fmt.Sprintf("传输通道开启 [%d/%d]", len(ch), r.tc))
		time.Sleep(LoopWaitTime / 10)
	}

	for {
		if len(ch) == r.tc*2 {
			DebugPrint(fmt.Sprintf("传输完成 [%d/%d]", len(ch)-r.tc, r.tc))
			DebugPrint("传输全部完成")
			break
		}
		DebugPrint(fmt.Sprintf("传输完成 [%d/%d]", len(ch)-r.tc, r.tc))
		time.Sleep(LoopWaitTime) //???这个地方去掉休眠就会卡住
	}

}

func (r *fsc) Processfscdata(indata []byte) {
	var tv1, tv2 []string
	tv1 = strings.Split(string(indata), "\n")
	for _, v := range tv1 {
		v = strings.ReplaceAll(v, "\n", "")
		v = strings.ReplaceAll(v, "\r", "")
		v = strings.ReplaceAll(v, " ", "")
		if v == "" {
			continue
		}
		tv2 = append(tv2, v)
	}
	if len(tv2) >= 5 {
		r.fn = tv2[0]
		r.fs, _ = strconv.Atoi(tv2[1])
		r.ps, _ = strconv.Atoi(tv2[2])
		r.pi, _ = strconv.Atoi(tv2[3])
		r.tc, _ = strconv.Atoi(tv2[4])
	}
}

func (r *fsc) Getdatas(ch chan string) {
	var vg sync.WaitGroup

	for i := 0; i <= r.tc-1; i++ {
		vg.Add(1)
		go func(nv int) {
			defer vg.Done()
			ListenIP := SIP + ":" + strconv.Itoa(1024*r.ps+r.pi+nv)
			listen, err := net.Listen("tcp", ListenIP)
			if err != nil {
				DebugPrint(err)
				return
			}
			defer listen.Close()

			go func() {
				ch <- ListenIP
			}()

			conn, err := listen.Accept()
			if err != nil {
				DebugPrint(err)
			}
			defer conn.Close()

			blocksize := fileslice
			if nv == r.tc-1 {
				blocksize = r.fs % fileslice
			}

			for {
				tb, err := ioutil.ReadAll(conn)
				if err != nil {
					conn.Close()
					time.Sleep(FailTryCount * LoopWaitTime)
					tb, _ = ioutil.ReadAll(conn)
				}
				defer conn.Close()
				if len(tb) == blocksize {
					var fsber fsb
					fsber.index = nv

					fsber.body = tb
					fsber.size = blocksize
					r.fsber = append(r.fsber, fsber)

					break
				}
				time.Sleep(LoopWaitTime)
			}
			ch <- fmt.Sprintf("%s:[ok]", ListenIP)
		}(i)
	}

	vg.Wait()

}

func (r *fsc) WriteFile() {
	var fb []byte

	if len(r.fsber) == r.tc {
		for i := 0; i <= r.tc-1; i++ {
			for _, v := range r.fsber {
				if v.index == i {
					fb = append(fb, v.body...)
				}
			}
		}
		r.fn = strings.ReplaceAll(r.fn, `\`, `/`)
		r.fn = path.Base(r.fn)
		err := ioutil.WriteFile(r.fn, fb, 0644)
		if err != nil {
			DebugPrint(err)
			return
		} else {
			DebugPrint("文件写入成功")
		}
	} else {
		DebugPrint("校验失败")
		r.V()
		return
	}
}

func (r *fsc) V() {
	for i := 0; i <= r.tc-1; i++ {
		for _, v := range r.fsber {
			if v.index == i {
				fmt.Printf("[total size:%d] [index:%d] [llsize:%d] [sjsize:%d] [%v:%v]\n", r.fs, v.index, v.size, len(v.body), v.body[:10], v.body[len(v.body)-10:])
			}
		}
	}
}
