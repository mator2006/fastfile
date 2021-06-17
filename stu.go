package main

import "time"

const (
	fileslice                        = 512 * 1024      //512k=524288
	PortNumber                       = 10050           //Port number for send and listening
	ListeningStratbMultipleRandRange = 10              //1024*this
	ListeningStratPlusRandRange      = 100             //1024*StratListeningbRand+this
	ListeningWaitTime                = 3 * time.Second //Wait listening start time
	SIP                              = "192.168.19.201"
	LIP                              = "192.168.19.201"
)

type ff struct {
	fsc struct {
		fn string //filename
		fs int    //file size
		ps int    //Port for Listening Strat Multiple
		pi int    //Port for Listening Strat Plus
		tc int    //Transmission Count
	}
	fsb struct {
		index []int    //index
		start []int    //byte start for file slice
		stop  []int    //byte stop for file slice
		body  [][]byte //byte body for file slice
	}
}
type fsn struct {
	index int
	size  int
}
