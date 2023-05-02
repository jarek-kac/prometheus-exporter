package filemetrics

import (
	"log"
	"regexp"
	"strconv"

	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func GetLogMetrics(logPath string) {
	tailAccessLogFile(logPath)
}

var exp = regexp.MustCompile(`^(?P<remote>[^ ]*) (?P<host>[^ ]*) (?P<user>[^ ]*) \[(?P<time>[^\]]*)\] \"(?P<method>\w+)(?:\s+(?P<path>[^\"]*?)(?:\s+\S*)?)?\" (?P<status_code>[^ ]*) (?P<size>[^ ]*)(?:\s"(?P<referer>[^\"]*)") "(?P<agent>[^\"]*)" (?P<urt>[^ ]*)$`)

func tailAccessLogFile(path string) {
	m := NewMetrics( /*reg*/ )
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

		m.requests.With(prometheus.Labels{"method": result["method"], "status_code": result["status_code"], "path": result["path"]}).Add(1)

		u, err := strconv.ParseFloat(result["urt"], 64)
		if err != nil {
			continue
		}
		m.duration.With(prometheus.Labels{"method": result["method"], "status_code": result["status_code"], "path": result["path"]}).Observe(u)

	}

}

type metrics struct {
	size     prometheus.Counter
	duration *prometheus.HistogramVec
	requests *prometheus.CounterVec
}

func NewMetrics( /*reg prometheus.Registerer*/ ) *metrics {
	m := &metrics{
		size: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "nginx",
			Name:      "size_bytes_total",
			Help:      "Total bytes sent to the clients.",
		}),
		requests: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "nginx",
			Name:      "jkk_http_requests_total",
			Help:      "Total number of requests.",
		}, []string{"status_code", "method", "path"}),
		duration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "nginx",
			Name:      "http_request_duration_seconds",
			Help:      "Duration of the request.",
			// Optionally configure time buckets
			// Buckets:   prometheus.LinearBuckets(0.01, 0.05, 20),
			Buckets: prometheus.DefBuckets,
		}, []string{"status_code", "method", "path"}),
	}
	//reg.MustRegister(m.size, m.requests, m.duration)
	return m
}
