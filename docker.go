package main

import (
	"strings"
	"os"
	"os/exec"
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

type ContainerIO struct {
	Name string
	PodName string
	Namespace string
	PID int
	CPIDs []int
	IOWait int
	IOWaitPercent float64
	ReadCalls int
	WriteCalls int
}

type Runner struct {
	CMap map[int]*ContainerIO
	CMapPre map[int]*ContainerIO
	Ticks int
	TicksPre int
	Services string
	ServicesArray []string
	Hostname string
}

func (r *Runner) GetContainers() {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.37"))
	if err != nil {
		fmt.Println(err)
		return
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		fmt.Println(err)
		return
	}

	r.CMap = make(map[int]*ContainerIO)

	for _, container := range containers {
		name := strings.Split(container.Names[0], "/")[1]
		splitted := strings.Split(name, "_")
		if splitted[0] == "k8s" {
			if splitted[1] != "POD" {
					insp, _ := cli.ContainerInspect(context.Background(), container.ID)
					c := ContainerIO{
						Name: splitted[1],
						PID: insp.ContainerJSONBase.State.Pid,
						PodName: splitted[2],
						Namespace: splitted[3],
					}
					r.CMap[insp.ContainerJSONBase.State.Pid] = &c
			}
		}
	}
}

func (r *Runner) Parse() {
	r.ServicesArray = strings.Split(r.Services, ",")
	r.Hostname = os.Getenv("IO_HOSTNAME")
}

func (r *Runner) GetServices() {
	for _, name := range r.ServicesArray {
		out, err := exec.Command("pidof", name).Output()
    if err != nil {
        fmt.Println(err)
    } else {
			str := strings.TrimSpace(string(out))
			pid, err := strconv.Atoi(str)
			if err != nil {
	        fmt.Println(err)
	    } else {
				c := ContainerIO{
					Name: name,
					PID: pid,
					PodName: "_",
					Namespace: "_",
				}
				r.CMap[pid] = &c
			}
		}
	}
}
