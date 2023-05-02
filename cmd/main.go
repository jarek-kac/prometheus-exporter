package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	logsexporter "github.com/jarek-kac/prometheus-exporter/logs-exporter"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var (
		targetHost = flag.String("target.host", "localhost", "nginx address with basic_status page")
		targetPort = flag.Int("target.port", 8080, "nginx port with basic_status page")
		targetPath = flag.String("target.path", "/status", "URL path to scrap metrics")
		promPort   = flag.String("prom.port", ":9150", "port to expose prometheus metrics")
		logPath    = flag.String("target.log", "access.log", "path to access.log")
	)
	flag.Parse()
	fmt.Print(*targetHost, *targetPort, targetPath, promPort, logPath)

	go logsexporter.GetLogMetrics(*logPath)

	recordMetrics()
	//reg := prometheus.NewRegistry()

	//promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	http.Handle("/metrics", promhttp.Handler())
	//http.Handle("/metrics", promHandler)
	log.Fatal(http.ListenAndServe(*promPort, nil))
}

func recordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
)
