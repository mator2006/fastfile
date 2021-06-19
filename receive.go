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

func (r *fsc) Listening() error {
	var err error
	ListenIP := fmt.Sprintf("%s:%s", ListeningIP, r.par.PortNumber)
	listen, err := net.Listen("tcp", ListenIP)
	if err != nil {
		DP(err)
		return err
	}
	defer listen.Close()
	IP("Listenning for file trandmission control signal.")

	conn, err := listen.Accept()
	if err != nil {
		DP(err)
		return err
	}
	defer conn.Close()

	buf := make([]byte, r.par.SignalByteSize)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			DP(err)
		}
		if bytes.Contains(buf[:n], []byte("EOF")) {
			DP("Get file transmission contron signal.")
			break
		}
		time.Sleep(r.par.LoopWaitTime)
	}

	err = r.Processfscdata(buf)
	if err != nil {
		DP(err)
		return err
	}

	go func() {
		err = r.ReciveData()
		if err != nil {
			DP(err)
			return
		}
	}()

	for {
		if len(r.ch) == r.tc {
			IP("All file data channel started.")
			break
		}
		//fmt.Printf("channel_lenth:[%d]\n", len(r.ch))
		time.Sleep(r.par.LoopWaitTime)
	}
	_, err = conn.Write([]byte("OK"))
	if err != nil {
		DP(err)
		return err
	}
	IP("Ready signal sended.")
	return err
}

func (r *fsc) ReciveData() error {
	var err error
	var vg sync.WaitGroup
	for i := 0; i < r.tc; i++ {
		vg.Add(1)
		go func(i int) {
			defer vg.Done()
			ListenIP := fmt.Sprintf("%s:%s", ListeningIP, strconv.Itoa(1024*r.ps+r.pi+i))
			listen, err := net.Listen("tcp", ListenIP)
			if err != nil {
				DP(err)
				return
			}
			defer listen.Close()

			r.ch <- fmt.Sprintf("%s:[Ready]", ListenIP)

			conn, err := listen.Accept()
			if err != nil {
				DP(err)
				return
			}
			defer conn.Close()

			blocksize := r.par.FileSliceSize
			if i == r.tc-1 {
				blocksize = r.fs % r.par.FileSliceSize
			}

			for {
				tb, err := ioutil.ReadAll(conn)
				if err != nil {
					if err != nil {
						DP(err)
						return
					}
				}
				if len(tb) == blocksize {
					var fsber fsb
					fsber.index = i

					fsber.body = tb
					fsber.size = blocksize
					r.fsber = append(r.fsber, fsber)
					r.ch <- fmt.Sprintf("[%d:%d]", len(fsber.body), fsber.size)
					break
				}
			}
		}(i)
	}
	vg.Wait()
	return err
}

func (r *fsc) Writefile() error {
	var err error
	var tvbs []byte
	for i := 0; i < r.tc; i++ {
		for _, v := range r.fsber {
			if v.index == i {
				tvbs = append(tvbs, v.body...)
				continue
			}
		}
	}

	if len(tvbs) == r.fs {
		err = ioutil.WriteFile(r.fn, tvbs, 0644)
		if err != nil {
			return err
		}
		IP("File wirte complete.")
	} else {
		err = fmt.Errorf("File verify fail!")
		r.V()
		return err
	}

	return err
}

func (r *fsc) Processfscdata(indata []byte) error {
	var err error
	var tv1, tv2 []string
	tv1 = strings.Split(string(indata), "\n")
	for _, v := range tv1 {
		v = StrFormat(v)
		if v == "" {
			continue
		}
		tv2 = append(tv2, v)
	}
	if len(tv2) >= 5 {
		r.fn = tv2[0]
		r.fs, err = strconv.Atoi(tv2[1])
		if err != nil {
			DP(err)
			return err
		}
		r.ps, err = strconv.Atoi(tv2[2])
		if err != nil {
			DP(err)
			return err
		}
		r.pi, err = strconv.Atoi(tv2[3])
		if err != nil {
			DP(err)
			return err
		}
		r.tc, err = strconv.Atoi(tv2[4])
		if err != nil {
			DP(err)
			return err
		}
	}
	switch {
	case r.fn == "":
		err = fmt.Errorf("Invalid data for filename!")
	case r.fs == 0:
		err = fmt.Errorf("Invalid data for filesize!")
	case r.ps == 0:
		err = fmt.Errorf("Invalid data for Port start!")
	case r.tc == 0:
		err = fmt.Errorf("Invalid data for transmission count!")
	}

	switch {
	case r.fs >= 100*M1 && r.fs < G1:
		r.par.FileSliceSize = M1
	case r.fs >= G1:
		r.par.FileSliceSize = 10 * M1
	default:
		r.par.FileSliceSize = 512 * 1024
	}

	r.ch = make(chan string, r.tc*3)

	return err
}
