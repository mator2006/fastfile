package main

import (
	"os"
)

var (
	DebugPrintSwitch bool = true
	InfoPrintSwitch  bool = true
)

func main() {

	Mode, SIP, Filename := ProcessArgs(os.Args)

	switch Mode {
	case 0:
		Send(SIP, Filename)
	case 1:
		Receive()
	case 2:
		Usage(os.Args[0])
	default:
		DP("Error")
		return
	}

}

func Send(SIP string, Filename string) {
	var s fsc
	s.Init()
	s.par.SIP = StrFormat(SIP)
	s.fn = StrFormat(Filename)

	var err error

	err = s.Readfsc()
	if err != nil {
		DP(err)
		return
	}

	err = s.SendSignal()
	if err != nil {
		DP(err)
		return
	}

	err = s.SendDatas()
	if err != nil {
		DP(err)
		return
	}

}

func Receive() {
	var r fsc
	r.Init()

	err := r.Listening()
	if err != nil {
		DP(err)
		return
	}

	err = r.Writefile()
	if err != nil {
		DP(err)
		return
	}
	IP("file write complete.")
}
