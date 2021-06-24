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

	r.WaitChan() //利用通道阻塞，等待任务完成
	IP("data channel build complete.")

	_, err = conn.Write([]byte("OK"))
	if err != nil {
		return err
	}
	IP("Ready signal sended.")

	r.WaitChan() //利用通道阻塞，等待任务完成
	IPfn()
	IP("Data reception completed.")

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

			r.ch <- ""

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
						DP(err) //不退出循环
						return
					}
				}
				if int64(len(tb)) == blocksize {
					var fsber fsb
					fsber.index = i

					fsber.body = &tb
					fsber.size = blocksize
					r.fsber = append(r.fsber, fsber)
					r.ch <- "."
					break
				}
			}
		}(i)
	}

	return err
}

func (r *fsc) Writefile() error {
	var err error
	r.fn = pathc(r.fn) //去掉路径
	f, err := os.Create(r.fn)
	if err != nil {
		DP(err)
		return err
	}
	defer f.Close()
	BufferedWriter := bufio.NewWriter(f)
	IP("File writing...")

	for i := 0; i < r.tc; i++ {
		for _, v := range r.fsber {
			if v.index == i {
				_, err := BufferedWriter.Write(*v.body)
				if err != nil {
					return err
				}

				break
			}
		}
	}
	BufferedWriter.Flush()
	IP("File write complete.")
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
	} else {
		err = fmt.Errorf("Invalid file transmission control data !")
		return err
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

	r.DefinitionFileSlice()

	//r.ch = make(chan fsb, r.tc*3)
	r.ch = make(chan string)

	return err
}
