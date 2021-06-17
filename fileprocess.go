package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"
)

func (ff *ff) GetFileS() error {
	var err error

	fb, err := os.Stat(ff.fsc.fn)
	if err != nil {
		DebugPrint(err)
		return err
	}
	ff.fsc.fs = int(fb.Size())

	if ff.fsc.fs == 0 {
		err = fmt.Errorf("file size 0")
		DebugPrint(err)
		return err
	}

	if ff.fsc.fs%fileslice != 0 {
		ff.fsc.tc = ff.fsc.fs/fileslice + 1
	} else {
		ff.fsc.tc = ff.fsc.fs / fileslice
	}

	rand.Seed(time.Now().UnixNano())
	ff.fsc.ps = rand.Intn(ListeningStratbMultipleRandRange)
	if ff.fsc.ps == 0 {
		ff.fsc.ps += 1
	}
	ff.fsc.pi = rand.Intn(ListeningStratPlusRandRange)

	for i := 0; i <= ff.fsc.tc-1; i++ {
		ff.fsb.index = append(ff.fsb.index, i)
		ff.fsb.start = append(ff.fsb.start, i*fileslice)
		if i == ff.fsc.tc-1 {
			ff.fsb.stop = append(ff.fsb.stop, ff.fsc.fs)
		} else {
			ff.fsb.stop = append(ff.fsb.stop, i*fileslice+fileslice)
		}
	}
	return err
}

func (ffer *ff) TestWriteFile() {
	fb, err := ioutil.ReadFile(ffer.fsc.fn)
	if err != nil {
		DebugPrint(err)
	}
	var vg sync.WaitGroup
	var buf bytes.Buffer
	for i := range ffer.fsb.start {
		vg.Add(1)
		go func(nv int) {
			defer vg.Done()
			buf.Write(fb[ffer.fsb.start[nv]:ffer.fsb.stop[nv]])
		}(i)
		vg.Wait()
		// fmt.Println(len(buf.Bytes()))
		err = ioutil.WriteFile("tt.cr2", buf.Bytes(), 644)
		if err != nil {
			DebugPrint(err)
		}
	}
}
