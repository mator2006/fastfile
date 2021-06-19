package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

func (s *fsc) Readfsc() error {
	var err error
	filestat, err := os.Stat(s.fn)
	if err != nil {
		DP(err)
		return err
	}
	s.fs = int(filestat.Size())
	rand.Seed(time.Now().UnixNano())
	s.ps = rand.Intn(s.par.ListeningStratbMultipleRandRange)
	if s.ps == 0 {
		s.ps += 1
	}

	rand.Seed(time.Now().UnixNano())
	s.pi = rand.Intn(s.par.ListeningStratPlusRandRange)
	switch {
	case s.fs >= 100*M1 && s.fs < G1:
		s.par.FileSliceSize = M1
	case s.fs >= G1:
		s.par.FileSliceSize = 10 * M1
	default:
		s.par.FileSliceSize = 512 * 1024
	}

	s.tc = s.fs / s.par.FileSliceSize
	if s.fs%s.par.FileSliceSize != 0 {
		s.tc += 1
	}

	s.ch = make(chan string, s.tc*3)

	return err
}

func (s *fsc) SendDatas() error {
	var err error
	for {
		if len(s.ch) == 1 {
			<-s.ch
			IP("File reading and sending...")
			break
		}
		time.Sleep(s.par.LoopWaitTime)
	}
	err = s.ReadFileBody()
	if err != nil {
		DP(err)
		return err
	}
	var vg sync.WaitGroup
	for i := 0; i < s.tc; i++ {
		vg.Add(1)
		go func(i int) {
			defer vg.Done()
			SendIP := fmt.Sprintf("%s:%d", s.par.SIP, s.ps*1024+s.pi+i)
			conn, err := net.Dial("tcp", SendIP)
			if err != nil {
				DP(err)
				return
			}
			defer conn.Close()
			n, err := conn.Write(s.fsber[i].body)
			if err != nil {
				DP(err)
				return
			}
			defer conn.Close()

			if n == s.fsber[i].size {
				fmt.Printf(".")
			} else {
				fmt.Printf("!")
				DP(fmt.Errorf("%s 发送失败，[实发送/应发送] [%d/%d]", SendIP, n, s.fsber[i].size))
			}

		}(i)
	}
	vg.Wait()
	fmt.Printf("\nFile transmission complete.\n")
	return err
}

func (s *fsc) ReadFileBody() error {
	var err error
	vtb, err := ioutil.ReadFile(s.fn)
	if err != nil {
		DP(err)
		return err
	}
	for i := 0; i < s.tc; i++ {
		var tv1 fsb
		tv1.index = i
		tv1.start = i * s.par.FileSliceSize

		tv1.stop = tv1.start + s.par.FileSliceSize
		if len(vtb) < tv1.stop {
			tv1.stop = len(vtb)
		}

		tv1.body = vtb[tv1.start:tv1.stop]
		tv1.size = len(tv1.body)
		s.fsber = append(s.fsber, tv1)
	}
	return err
}

func (s *fsc) SendSignal() error {
	var err error
	conn, err := net.Dial("tcp", s.par.SIP+":"+s.par.PortNumber)
	if err != nil {
		DP(err)
		return err
	}
	defer conn.Close()

	_, err = conn.Write([]byte(fmt.Sprintf("%s\n%d\n%d\n%d\n%d\nEOF\n", s.fn, s.fs, s.ps, s.pi, s.tc)))
	if err != nil {
		DP(err)
		return err
	}
	defer conn.Close()
	IP("File send control data sended.")

	tb := make([]byte, s.par.SignalByteSize)

	_, err = conn.Read(tb)
	if err != nil {
		DP(err)
	}
	defer conn.Close()
	if bytes.Contains(tb, []byte("OK")) {
		IP("Get signal for server ready,send start.")
		s.ch <- "OK"
		return err
	}
	err = fmt.Errorf("Error signal!")
	return err
}
