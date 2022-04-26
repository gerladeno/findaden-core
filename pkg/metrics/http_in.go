package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type HTTPIn struct {
	Uptime          prometheus.GaugeFunc
	ReqTotal        *prometheus.CounterVec
	ReqBytesTotal   *prometheus.CounterVec
	ReqErrorsTotal  *prometheus.CounterVec
	RespTotal       *prometheus.CounterVec
	RespBytesTotal  *prometheus.CounterVec
	RespErrorsTotal *prometheus.CounterVec
	RespTimeHist    *prometheus.HistogramVec
	RespTimeTotal   *prometheus.GaugeVec
}

func NewHTTPIn(host, ip, port string) *HTTPIn { //nolint:funlen
	constLabels := prometheus.Labels{"http_in_host": host, "http_in_ip": ip, "http_in_port": port}
	started := time.Now()
	return &HTTPIn{
		Uptime: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name:        "http_in_uptime",
				Help:        "Seconds since the HTTP listener has started",
				ConstLabels: constLabels,
			}, func() float64 {
				return time.Since(started).Seconds()
			}),
		ReqTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "http_in_requests_total",
			Help:        "Total amount of incoming HTTP requests on the process",
			ConstLabels: constLabels,
		}, []string{
			"http_in_source_ip",
			"http_in_source_port",
			"http_in_method",
			"http_in_url",
		}),
		ReqBytesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "http_in_request_bytes_total",
			Help:        "Total amount of incoming HTTP requests on the process",
			ConstLabels: constLabels,
		}, []string{
			"http_in_source_ip",
			"http_in_source_port",
			"http_in_method",
			"http_in_url",
		}),
		ReqErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "http_in_request_errors_total",
			Help:        "How many errors came from client, partitioned",
			ConstLabels: constLabels,
		}, []string{
			"http_in_source_ip",
			"http_in_source_port",
			"http_in_method",
			"http_in_url",
			"http_in_response_error",
		}),
		RespTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "http_in_responses_total",
			Help:        "How many HTTP responses gone, partitioned",
			ConstLabels: constLabels,
		}, []string{
			"http_in_source_ip",
			"http_in_source_port",
			"http_in_method",
			"http_in_url",
			"http_in_response_code",
		}),
		RespBytesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "http_in_response_bytes_total",
			Help:        "Content length or response size",
			ConstLabels: constLabels,
		}, []string{
			"http_in_source_ip",
			"http_in_source_port",
			"http_in_method",
			"http_in_url",
			"http_in_response_code",
		}),
		RespErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "http_in_response_errors_total",
			Help:        "Total amount of outgoing errors on each response",
			ConstLabels: constLabels,
		}, []string{
			"http_in_source_ip",
			"http_in_source_port",
			"http_in_method",
			"http_in_url",
			"http_in_response_code",
			"http_in_response_error",
		}),
		RespTimeHist: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:        "http_in_response_time_hist",
			Help:        "Total amount of time spent on the response",
			ConstLabels: constLabels,
			Buckets:     []float64{50, 100, 300, 1000, 5000},
		}, []string{
			"http_in_source_ip",
			"http_in_source_port",
			"http_in_method",
			"http_in_url",
			"http_in_response_code",
		}),
		RespTimeTotal: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:        "http_in_response_time_total",
			Help:        "Total amount of time spent on the response",
			ConstLabels: constLabels,
		}, []string{
			"http_in_source_ip",
			"http_in_source_port",
			"http_in_method",
			"http_in_url",
			"http_in_response_code",
		}),
	}
}

var httpInOnce sync.Once

func (h *HTTPIn) AutoRegister() *HTTPIn {
	httpInOnce.Do(func() {
		h.mustRegister(prometheus.DefaultRegisterer)
	})
	return h
}

func (h *HTTPIn) mustRegister(registerer prometheus.Registerer) {
	registerer.MustRegister(
		h.Uptime,
		h.ReqTotal,
		h.ReqBytesTotal,
		h.ReqErrorsTotal,
		h.RespTotal,
		h.RespBytesTotal,
		h.RespErrorsTotal,
		h.RespTimeHist,
		h.RespTimeTotal,
	)
}
