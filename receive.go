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
			}
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
