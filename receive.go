package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
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
	}
	r.Processfscdata(buf)
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

func (r *fsc) SendReadySignials(i int) bool {
	var b bool
	SendIP := CIP + ":" + strconv.Itoa(PortNumber)
	conn, err := net.Dial("tcp", SendIP)
	if err != nil {
		DebugPrint(fmt.Errorf("%v,第[%d]次重试", err, i+1))
		return b
	}
	defer conn.Close()

	_, err = conn.Write([]byte("OK"))
	if err != nil {
		DebugPrint(err)
		return b
	}
	defer conn.Close()
	return true
}

func (r *fsc) Getdatas() {
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

			conn, err := listen.Accept()
			if err != nil {
				DebugPrint(err)
			}
			defer conn.Close()

			blocksize := fileslice
			if nv == r.tc-1 {
				blocksize = r.fs % fileslice
			}
			var s int
			buf := make([]byte, blocksize)
			for {
				n, err := conn.Read(buf)
				if err != nil {
					conn.Close()
					if s == FailTryCount {
						DebugPrint(fmt.Errorf("%s 数据读取错误，实读取/应读取:[%d/%d]", ListenIP, n, blocksize))
						break
					}
					time.Sleep(LoopWaitTime)
					s++
					// DebugPrint(err)
				}
				defer conn.Close()

				if n == blocksize {
					var fsber fsb
					fsber.index = nv
					fsber.body = buf[:n]
					fsber.size = blocksize
					r.fsber = append(r.fsber, fsber)
					break
				}
			}
		}(i)
	}
	vg.Wait()
}

func (r *fsc) WriteFile() {
	var s int
	var buf bytes.Buffer
	if len(r.fsber) == r.tc {
		for _, v := range r.fsber {
			if len(v.body) == v.size {
				s++
			}
		}
		if s == len(r.fsber) {
			fmt.Println("校验成功")
			for i := 1; i <= r.tc-1; i++ {
				for _, v := range r.fsber {
					if v.index == i {
						buf.Write(v.body)
					}
				}
			}
			err := ioutil.WriteFile(r.fn, buf.Bytes(), 0644)
			if err != nil {
				DebugPrint(err)
				return
			} else {
				fmt.Println("文件写入成功")
			}
		}
	} else {
		fmt.Println("校验失败")
		return
	}
}
