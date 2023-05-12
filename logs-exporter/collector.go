package filemetrics

import (
	"log"
	"regexp"
	"strconv"

	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
)

var exp = regexp.MustCompile(`^(?P<remote>[^ ]*) (?P<host>[^ ]*) (?P<user>[^ ]*) \[(?P<time>[^\]]*)\] \"(?P<method>\w+)(?:\s+(?P<path>[^\"]*?)(?:\s+\S*)?)?\" (?P<status_code>[^ ]*) (?P<size>[^ ]*)(?:\s"(?P<referer>[^\"]*)") "(?P<agent>[^\"]*)" (?P<urt>[^ ]*)$`)

type Metrics struct {
	size     prometheus.Counter
	duration *prometheus.HistogramVec
	requests *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	m := &Metrics{
		size: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "nginx",
			Name:      "size_bytes_total",
			Help:      "Total bytes sent to the clients.",
		}),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "nginx",
			Name:      "jkk_http_requests_total",
			Help:      "Total number of requests.",
		}, []string{"status_code", "method", "path"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "nginx",
			Name:      "http_request_duration_seconds",
			Help:      "Duration of the request.",
			// Optionally configure time buckets
			// Buckets:   prometheus.LinearBuckets(0.01, 0.05, 20),
			Buckets: prometheus.DefBuckets,
		}, []string{"status_code", "method", "path"}),
	}
	return m
}

func (m *Metrics) Describe(ch chan<- *prometheus.Desc) {
	m.size.Describe(ch)
	m.requests.Describe(ch)
	m.duration.Describe(ch)
}

func (m *Metrics) Collect(ch chan<- prometheus.Metric) {

	var path string = "access.log"
	t, err := tail.TailFile(path, tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		log.Fatalf("tail.TailFile failed: %s", err)
	}

	for line := range t.Lines {
		match := exp.FindStringSubmatch(line.Text)
		result := make(map[string]string)

		for i, name := range exp.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}

		s, err := strconv.ParseFloat(result["size"], 64)
		if err != nil {
			continue
		}
		m.size.Add(s)

		//m.requests.With(prometheus.Labels{"method": result["method"], "status_code": result["status_code"], "path": result["path"]}).Add(1)
		m.requests.WithLabelValues(result["method"], result["status_code"], result["path"]).Add(1)

		u, err := strconv.ParseFloat(result["urt"], 64)
		if err != nil {
			continue
		}
		//m.duration.With(prometheus.Labels{"method": result["method"], "status_code": result["status_code"], "path": result["path"]}).Observe(u)
		m.duration.WithLabelValues(result["method"], result["status_code"], result["path"]).Observe(u)

	}

}
