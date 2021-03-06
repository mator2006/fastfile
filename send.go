package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"time"
)

func (s *fsc) SendDatas() error {
	var err error

	fb, err := ioutil.ReadFile(s.fn)
	if err != nil {
		return err
	}
	IP("File read complete.")

	for i := 0; i < s.tc; i++ {
		go func(i int) {
			SendIP := fmt.Sprintf("%s:%d", s.par.SIP, s.ps*1024+s.pi+i)
			conn, err := net.Dial("tcp", SendIP)
			if err != nil {
				DP(err)
				return
			}
			defer conn.Close()

			n, err := conn.Write(fb[s.fsber[i].start:s.fsber[i].stop])
			if err != nil {
				DP(err)
				return
			}
			defer conn.Close()

			if int64(n) == s.fsber[i].size {
				s.ch <- "."
			} else {
				s.ch <- "!"
				// DP(fmt.Errorf("%s 发送失败，[实发送/应发送] [%d/%d]", SendIP, n, s.fsber[i].size))
			}

		}(i)
	}

	s.WaitChan()
	IPfn()
	IP("File send complete.")
	return err
}

func (s *fsc) SendSignal() error {
	var err error
	conn, err := net.Dial("tcp", s.par.SIP+":"+s.par.PortNumber)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write([]byte(fmt.Sprintf("%s\n%d\n%d\n%d\n%d\nEOF\n", s.fn, s.fs, s.ps, s.pi, s.tc)))
	if err != nil {
		return err
	}
	defer conn.Close()
	IP("File transmission control data sended.")

	tb := make([]byte, s.par.SignalByteSize)
	for {
		_, err = conn.Read(tb)
		if err != nil {
			DP(err) //不退出
		}
		defer conn.Close()
		if bytes.Contains(tb, []byte("OK")) {
			IP("Get signal for server ready,send start.")
			break
		}
		time.Sleep(s.par.LoopWaitTime)
	}
	return err
}

func (s *fsc) Readfsc() error {
	var err error
	filestat, err := os.Stat(s.fn)
	if err != nil {
		return err
	}
	s.fs = int64(filestat.Size())
	rand.Seed(time.Now().UnixNano())
	s.ps = rand.Intn(s.par.ListeningStratbMultipleRandRange)
	if s.ps == 0 {
		s.ps += 11
	} else {
		s.ps += 10 //低端口更可能被占用，使用高端口
	}

	rand.Seed(time.Now().UnixNano())
	s.pi = rand.Intn(s.par.ListeningStratPlusRandRange)

	s.DefinitionFileSlice()

	s.tc = int(s.fs / s.par.FileSliceSize) //多出来的不够一块，也占一块
	if s.fs%s.par.FileSliceSize != 0 {
		s.tc += 1
	}

	s.ch = make(chan string)

	for i := 0; i < s.tc; i++ {
		var tv1 fsb
		tv1.index = i
		tv1.start = int64(i) * s.par.FileSliceSize

		tv1.stop = tv1.start + s.par.FileSliceSize
		if tv1.stop > s.fs {
			tv1.stop = s.fs
		}

		tv1.size = tv1.stop - tv1.start
		s.fsber = append(s.fsber, tv1)
	}

	return err
}
