package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	logsexporter "github.com/jarek-kac/prometheus-exporter/logs-exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
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

	exporte := logsexporter.NewExp()
	prometheus.MustRegister(exporte)
	prometheus.MustRegister(version.NewCollector("apache_exporter"))

	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(*promPort, nil))
}
