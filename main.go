package main

import (
	// "fmt"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	io_wait = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "io_wait",
			Help: "I/O wait in ticks",
		},
		[]string{"io_container_name", "io_pod_name", "io_namespace", "io_hostname"},
	)
	io_rcalls = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "io_rcalls",
			Help: "I/O read calls",
		},
		[]string{"io_container_name", "io_pod_name", "io_namespace", "io_hostname"},
	)
	io_wcalls = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "io_wcalls",
			Help: "I/O write calls",
		},
		[]string{"io_container_name", "io_pod_name", "io_namespace", "io_hostname"},
	)
)

func main() {
	addr := flag.String("web.listen-address", ":9600", "Address on which to expose metrics")
	interval := flag.Int("interval", 10, "Interval fo metrics collection in seconds")
	services := flag.String("services", "etcd", "Additional services")
	flag.Parse()
	prometheus.MustRegister(io_wait)
	prometheus.MustRegister(io_rcalls)
	prometheus.MustRegister(io_wcalls)
	http.Handle("/metrics", prometheus.Handler())
	go run(int(*interval), *services)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func run(interval int, services string) {
	r := Runner{Services: services}
	r.Parse()
	for {
		r.Copy()
		r.RunJob()
		for _, c := range r.CMap {
			io_wait.With(prometheus.Labels{"io_container_name": c.Name, "io_pod_name": c.PodName, "io_namespace": c.Namespace, "io_hostname": r.Hostname}).Set(float64(c.IOWaitPercent))
			io_rcalls.With(prometheus.Labels{"io_container_name": c.Name, "io_pod_name": c.PodName, "io_namespace": c.Namespace, "io_hostname": r.Hostname}).Set(float64(c.ReadCalls))
			io_wcalls.With(prometheus.Labels{"io_container_name": c.Name, "io_pod_name": c.PodName, "io_namespace": c.Namespace, "io_hostname": r.Hostname}).Set(float64(c.WriteCalls))
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
