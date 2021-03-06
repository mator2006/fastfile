package main

import "time"

const (
	M1          = 1 * 1024 * 1024        //1M
	G1          = 1 * 1024 * 1024 * 1024 //1G                                                                                                                                                                                //1T
	ListeningIP = "0.0.0.0"              //ListeningIP                                                                                                                                                                                //1T
	UsageText   = `
	Usage:

	[f] -h                      print usage
	[f]                         enter listenning Mode
	[f] ipadress filename       enter send mode,ipadress is server ip,filename is send file name
	
	CopyRight Mator by 2021.06`  //Usage
)

type fsc struct {
	par   ProgParameter //Program Parameter
	fn    string        //filename
	fs    int64         //file size
	ps    int           //Port for Listening Strat Multiple
	pi    int           //Port for Listening Strat Plus
	tc    int           //Transmission Count
	fsber []fsb         //File body
	ch    chan string   //channel
}

type fsb struct {
	index int     //index
	start int64   //byte start for file slice
	stop  int64   //byte stop for file slice
	body  *[]byte //byte body for file slice
	size  int64   //byte body size for file slice
}

type ProgParameter struct {
	PortNumber                       string        //Port number for signal
	ListeningStratbMultipleRandRange int           //1024*this
	ListeningStratPlusRandRange      int           //1024*StratListeningbRand+this
	ListeningWaitTime                time.Duration //Wait listening start time
	LoopWaitTime                     time.Duration //Wait listening start time
	FileSliceSize                    int64         //File slice size
	SignalByteSize                   int           //SignalByteSize
	FailTryCount                     int           //Fail Try Count
	SIP                              string        //Server IP
}
