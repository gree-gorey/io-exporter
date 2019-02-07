package main

import (
	"fmt"
  "strings"
	"strconv"
	"io/ioutil"
  "sync"
  "os"
  "bufio"

	"github.com/Bowery/proc"
)

func (r *Runner) RunJob() {
  r.GetContainers()
	r.GetServices()
	tree, err := proc.GetPidTree(1)
	if err != nil {
		fmt.Println("unable to get proc tree", err)
	} else {
		getChildren(tree, false, &r.CMap, 0)
	}
	r.GetTicks()
  var wg sync.WaitGroup
	for _, c := range r.CMap {
		wg.Add(1)
    go func (c *ContainerIO)  {
      defer wg.Done()
      r.getIoMain(c)
    }(c)
	}
	wg.Wait()
}

func (r *Runner) Copy() {
	r.TicksPre = r.Ticks
	r.CMapPre = r.CMap
	r.CMap = nil
	r.Ticks = 0
}

func (r *Runner) GetTicks() {
	name := "/host/proc/stat"
  f, err := os.Open(name)
	if err != nil {
		fmt.Println("unable to open", name)
	}
  defer f.Close()
  reader := bufio.NewReader(f)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "cpu ") {
      items := strings.Split(scanner.Text(), " ")[1:]
			for _, iS := range items {
				i, _ := strconv.Atoi(iS)
				r.Ticks += i
			}
		}
  }
}

func (r *Runner) getIoMain(c *ContainerIO) {
  var wg sync.WaitGroup
	for _, cPID := range c.CPIDs {
		wg.Add(2)
    go func (cPID int, c *ContainerIO)  {
      defer wg.Done()
      getIo(cPID, c)
    }(cPID, c)
    go func (cPID int, c *ContainerIO)  {
      defer wg.Done()
      getIoRW(cPID, c)
    }(cPID, c)
	}
	wg.Wait()
	if _, ok := (*r).CMapPre[c.PID]; ok {
		c.IOWaitPercent = 100 * float64(c.IOWait - (*r).CMapPre[c.PID].IOWait) / float64(r.Ticks - r.TicksPre)
	} else {
		c.IOWaitPercent = 0
	}
}

func getIo(PID int, c *ContainerIO) {
	name := "/host/proc/" + strconv.Itoa(PID) + "/stat"
	dat, err := ioutil.ReadFile(name)
	if err != nil {
		fmt.Println("unable to open", name)
		return
	}
	iowaitS := strings.Split(string(dat), " ")[41]
  iowait, err := strconv.Atoi(iowaitS)
	if err != nil {
		fmt.Println(err)
	}
  c.IOWait += iowait
}

func getIoRW(PID int, c *ContainerIO) {
	name := "/host/proc/" + strconv.Itoa(PID) + "/io"
  f, err := os.Open(name)
	if err != nil {
		fmt.Println("unable to open", name)
		return
	}
  defer f.Close()
  r := bufio.NewReader(f)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "syscr") {
      rcallsS := strings.Split(scanner.Text(), ": ")[1]
      rcalls, err := strconv.Atoi(rcallsS)
    	if err != nil {
    		fmt.Println(err)
    	}
      c.ReadCalls += rcalls
		}
    if strings.Contains(scanner.Text(), "syscw") {
      wcallsS := strings.Split(scanner.Text(), ": ")[1]
      wcalls, err := strconv.Atoi(wcallsS)
    	if err != nil {
    		fmt.Println(err)
    	}
      c.WriteCalls += wcalls
		}
  }
}

func getChildren(procItem *proc.Proc, saveCh bool, save *map[int]*ContainerIO, main int) {
  if _, ok := (*save)[procItem.Pid]; ok {
    saveCh = true
    main = procItem.Pid
  } else if saveCh {
    saveCh = true
  } else {
    saveCh = false
    main = 0
  }
  if main != 0 {
    (*save)[main].CPIDs = append((*save)[main].CPIDs, procItem.Pid)
  }
  for _, ch := range procItem.Children {
    getChildren(ch, saveCh, save, main)
  }
}
