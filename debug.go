package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ProcessArgs(args []string) (int, string) {
	if len(args) > 1 {
		return 0, args[1]
	}
	return 1, "IMG_0151.CR2"
}

func DebugPrint(pc interface{}) {
	t := time.Now().Format("2006/01/02 15:04:05")
	if DW {
		fmt.Println(t, pc)
	}
}

func ProcessReceivefsc(RecivedBytes []byte) ff {
	var refscer ff
	var sss []string
	s := string(RecivedBytes)
	if s == "" {
		return refscer
	}
	ss := strings.Split(s, "\n")
	if len(ss) < 5 {
		return refscer
	}
	for _, v := range ss {
		v = strings.ReplaceAll(v, "\n", "")
		v = strings.ReplaceAll(v, "\r", "")
		v = strings.ReplaceAll(v, " ", "")
		sss = append(sss, v)
	}
	refscer.fsc.fn = sss[0]
	refscer.fsc.fs, _ = strconv.Atoi(sss[1])
	refscer.fsc.ps, _ = strconv.Atoi(sss[2])
	refscer.fsc.pi, _ = strconv.Atoi(sss[3])
	refscer.fsc.tc, _ = strconv.Atoi(sss[4])
	return refscer
}
