package main

import "time"

const (
	fileslice                        = 512 * 1024             //512k=524288
	PortNumber                       = 10050                  //Port number for signal
	ListeningStratbMultipleRandRange = 5                      //1024*this
	ListeningStratPlusRandRange      = 1000                   //1024*StratListeningbRand+this
	ListeningWaitTime                = 3 * time.Second        //Wait listening start time
	LoopWaitTime                     = 100 * time.Millisecond //Wait listening start time
	SignalByteSize                   = 512                    //SignalByteSize
	FailTryCount                     = 5                      //Fail Try Count
	SIP                              = "192.168.19.201"
	CIP                              = "192.168.19.202"
)

type fsc struct {
	fn    string //filename
	fs    int    //file size
	ps    int    //Port for Listening Strat Multiple
	pi    int    //Port for Listening Strat Plus
	tc    int    //Transmission Count
	fsber []fsb  //File body
}

type fsb struct {
	index int    //index
	start int    //byte start for file slice
	stop  int    //byte stop for file slice
	body  []byte //byte body for file slice
	size  int    //byte body size for file slice
}
