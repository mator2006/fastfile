package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
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

	for { //等待通道建立完成，发送OK信号
		if len(r.ch) == r.tc {
			_, err = conn.Write([]byte("OK"))
			if err != nil {
				DP(err)
				return err
			}
			IP("Ready signal sended.")
			break
		}
		time.Sleep(r.par.LoopWaitTime)
	}

	for { //等待数据传输完成
		if len(r.ch) == r.tc*2 {
			if DebugPrintSwitch {
				fmt.Printf("\n")
			}
			IP("file transmission complete.")
			break
		}
		time.Sleep(r.par.LoopWaitTime)
	}

	return err
}

func (r *fsc) ReciveData() error {
	var err error

	for i := 0; i < r.tc; i++ {
		go func(i int) {
			ListenIP := fmt.Sprintf("%s:%s", ListeningIP, strconv.Itoa(1024*r.ps+r.pi+i))
			listen, err := net.Listen("tcp", ListenIP)
			if err != nil {
				DP(err)
				return
			}
			defer listen.Close()

			var vt1 fsb
			r.ch <- vt1

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
				if int64(len(tb)) == blocksize {
					if InfoPrintSwitch {
						fmt.Printf(".")
					}
					var fsber fsb
					fsber.index = i

					fsber.body = &tb
					fsber.size = blocksize
					r.fsber = append(r.fsber, fsber)
					r.ch <- fsber
					break
				}
			}
		}(i)
	}

	return err
}

func (r *fsc) Writefile() error {
	var err error
	f, err := os.Create(r.fn)
	if err != nil {
		DP(err)
		return err
	}
	defer f.Close()
	BufferedWriter := bufio.NewWriter(f)

	for i := 0; i < r.tc; i++ {
		for _, v := range r.fsber {
			if v.index == i {
				n, err := BufferedWriter.Write(*v.body)
				if err != nil {
					DP(err)
					return err
				}
				if int64(n) == v.size {
					if InfoPrintSwitch {
						fmt.Printf(".")
					} else {
						if InfoPrintSwitch {
							fmt.Printf("!")
						}
					}
				}
				break
			}
		}
	}
	if InfoPrintSwitch {
		fmt.Printf("\n")
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
		r.fs, err = strconv.ParseInt(tv2[1], 10, 64)
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
		return err
	case r.fs == 0:
		err = fmt.Errorf("Invalid data for filesize!")
		return err
	case r.ps == 0:
		err = fmt.Errorf("Invalid data for Port start!")
		return err
	case r.tc == 0:
		err = fmt.Errorf("Invalid data for transmission count!")
		return err
	}

	switch {
	case r.fs >= 100*M1 && r.fs < G1:
		r.par.FileSliceSize = 5 * M1
	case r.fs >= G1:
		r.par.FileSliceSize = 10 * M1
	default:
		r.par.FileSliceSize = 512 * 1024
	}

	r.ch = make(chan fsb, r.tc*3)

	return err
}
