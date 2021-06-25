package main

import (
	"crypto/md5"
	"fmt"
	"path"
	"strings"
	"time"
)

func ProcessArgs(args []string) (int, string, string) {
	switch len(args) {
	case 1:
		IP("Listening Mode")
		return 1, "", ""
	case 2:
		if strings.ReplaceAll(args[1], " ", "") == "-h" || strings.ReplaceAll(args[1], " ", "") == "--help" {
			DP("Help Mode")
			return 2, "", ""
		}
		DP("Error Mode")
		return 3, "", ""
	case 3:
		DP("Send Mode")
		return 0, args[1], args[2]
	default:
		DP("Error Mode")
		return 3, "", ""
	}
}

func DP(pc interface{}) {
	t := time.Now().Format("2006/01/02 15:04:05")
	if DebugPrintSwitch {
		fmt.Println(t, pc)
	}
}

func IP(pc interface{}) {
	t := time.Now().Format("2006/01/02 15:04:05")
	if InfoPrintSwitch {
		fmt.Println(t, pc)
	}
}

func IPfn() {
	if InfoPrintSwitch {
		fmt.Printf("\n")
	}
}

func pathc(filename string) string {
	return path.Base(strings.ReplaceAll(filename, `\`, `/`))
}

func StrFormat(instr string) string {
	instr = strings.ReplaceAll(instr, "\n", "")
	instr = strings.ReplaceAll(instr, "\r", "")
	instr = strings.ReplaceAll(instr, " ", "")
	// instr = strings.ReplaceAll(instr, "\n", "")
	return instr
}

func (p *fsc) Init() {
	p.par.PortNumber = "10050"

	p.par.ListeningStratPlusRandRange = 1000
	p.par.ListeningStratbMultipleRandRange = 5

	p.par.ListeningWaitTime = 5 * time.Second
	p.par.LoopWaitTime = 100 * time.Millisecond

	p.par.SignalByteSize = 512

	p.par.FailTryCount = 5
}

func (s *fsc) V() {
	fmt.Printf("\n\n%+v\n\nPress enter key to contiue.", s.par)
	_, _ = fmt.Scanln()
	fmt.Printf("\n\n{filename:%s filesize:%d PortM:%d PortI:%d TC:%d channel_lenth:%d}\n\nPress enter key to contiue.\n\n", s.fn, s.fs, s.ps, s.pi, s.tc, len(s.ch))
	_, _ = fmt.Scanln()
	var tvbs []byte
	for _, v := range s.fsber {
		vt1 := *v.body
		fmt.Printf("[%d]\t[%d]\t[%d:%d]\t%s\t%v\n", v.index, v.size, v.start, v.stop, fmt.Sprintf("%x", md5.Sum(*v.body)), vt1[:4])
		tvbs = append(tvbs, *v.body...)
	}
	fmt.Printf("\nPress enter key to contiue.")
	_, _ = fmt.Scanln()
	fmt.Printf("Received data total is [%d],MD5 hash is [%x]\n", len(tvbs), md5.Sum(tvbs))
}

func Usage(Args string) {
	Args = pathc(Args)
	text := strings.ReplaceAll(UsageText, "[f]", Args)
	fmt.Println(text)
}

func (ds *fsc) DefinitionFileSlice() {
	switch {
	case ds.fs >= 100*M1 && ds.fs < G1: //根据文件大小调整分块大小
		ds.par.FileSliceSize = 5 * M1
	case ds.fs >= G1:
		ds.par.FileSliceSize = 10 * M1
	default:
		ds.par.FileSliceSize = 512 * 1024
	}
}

func (ss *fsc) WaitChan() {

	for i := 0; i < ss.tc; i++ {
		x := <-ss.ch
		if x != "" {
			if InfoPrintSwitch {
				bar(i, ss.tc)
			}
		}
	}
}

func bar(n, s int) {
	var str1 = "-"
	var str2 = ">"

	var barlenth float32 = 80
	nf := float32(n + 1)
	sf := float32(s)
	barsinit := strings.Repeat(str1, int(barlenth))

	bars := strings.Replace(barsinit, str1, str2, int(nf/sf*barlenth))

	fmt.Printf("\r[%s%s] [%d/%d]", str2, bars[1:], n+1, s)
}
