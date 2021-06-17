package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

func (s *fsc) Initfsc() {
	f, err := os.Stat(s.fn)
	if err != nil {
		DebugPrint(fmt.Errorf("%v:%v", s.fn, err))
		return
	}

	s.fs = int(f.Size())

	rand.Seed(time.Now().UnixNano())
	s.ps = rand.Intn(ListeningStratbMultipleRandRange)
	if s.ps == 0 {
		s.ps += 1
	}

	rand.Seed(time.Now().UnixNano())
	s.pi = rand.Intn(ListeningStratPlusRandRange)

	s.tc = s.fs / fileslice
	if s.fs%fileslice != 0 {
		s.tc += 1
	}

}

func (s *fsc) Getsignals() {
	ListenIP := CIP + ":" + strconv.Itoa(PortNumber)
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
		if bytes.Contains(buf[:n], []byte("OK")) {
			DebugPrint("Get OK signal")
			break
		}
	}
}

func (s *fsc) Sendfsc(i int) bool {
	var b bool
	SendIP := SIP + ":" + strconv.Itoa(PortNumber)
	conn, err := net.Dial("tcp", SendIP)
	if err != nil {
		DebugPrint(fmt.Errorf("%v,第[%d]次重试", err, i+1))
		return b
	}
	defer conn.Close()

	_, err = conn.Write([]byte(fmt.Sprintf("%s\n%d\n%d\n%d\n%d\nEOF\n", s.fn, s.fs, s.ps, s.pi, s.tc)))
	if err != nil {
		DebugPrint(err)
		return b
	}
	defer conn.Close()
	return true
}

func (s *fsc) Senddatas() {
	s.ReadFile()
	var vg sync.WaitGroup
	for _, v := range s.fsber {
		vg.Add(1)
		go func(fsber fsb) {
			defer vg.Done()
			SendIP := SIP + ":" + strconv.Itoa(1024*s.ps+s.pi+fsber.index)
			conn, err := net.Dial("tcp", SendIP)
			if err != nil {
				DebugPrint(err)
				return
			}
			defer conn.Close()

			n, err := conn.Write(fsber.body)
			if err != nil {
				DebugPrint(err)
			}
			defer conn.Close()

			if n == fsber.size {
				//DebugPrint(fmt.Sprintf("%s 发送成功，发送了[%d]byte", SendIP, fsber.size))
			} else {
				DebugPrint(fmt.Errorf("%s 发送失败，[应发送/实发送] [%d/%d]", SendIP, n, fsber.size))
			}
		}(v)
	}
	vg.Wait()
}

func (s *fsc) ReadFile() {
	fb, err := ioutil.ReadFile(s.fn)
	if err != nil {
		DebugPrint(err)
		return
	}
	for i := 0; i <= s.tc-1; i++ {
		var tv1 fsb
		tv1.index = i
		tv1.start = i * fileslice

		tv1.stop = tv1.start + fileslice
		if len(fb) < tv1.stop {
			tv1.stop = len(fb)
		}

		tv1.body = fb[tv1.start:tv1.stop]
		tv1.size = len(tv1.body)
		s.fsber = append(s.fsber, tv1)
	}
}
