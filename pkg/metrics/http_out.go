package metrics

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gerladeno/homie-core/pkg/common"
	"github.com/prometheus/client_golang/prometheus"
)

type HTTPOut struct {
	DNSResolveTime      *prometheus.GaugeVec
	ConnectTime         *prometheus.CounterVec
	HandshakeTime       *prometheus.CounterVec
	ReqTotal            *prometheus.CounterVec
	ReqBytesTotal       *prometheus.CounterVec
	ReqErrorsTotal      *prometheus.CounterVec
	RespTotal           *prometheus.CounterVec
	RespBytesTotal      *prometheus.CounterVec
	RespErrorsTotal     *prometheus.CounterVec
	RespTimeToFirstByte *prometheus.GaugeVec
	RespTimeTotal       *prometheus.CounterVec
}

func NewHTTPOut(host string) *HTTPOut { //nolint:funlen
	constLabels := prometheus.Labels{"host": host}
	return &HTTPOut{
		DNSResolveTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        "http_out_dns_time",
				Help:        "Time that spends on DNS resolve",
				ConstLabels: constLabels,
			},
			[]string{
				"http_out_host",
				"http_out_port",
			}),
		ConnectTime: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "http_out_connect_time_total",
				Help:        "Time that spends on connect to the remote endpoint",
				ConstLabels: constLabels,
			},
			[]string{
				"http_out_host",
				"http_out_port",
			}),
		HandshakeTime: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "http_out_handshake_time_total",
				Help:        "Time that spends on handshake to the remote endpoint in a case of TLS",
				ConstLabels: constLabels,
			},
			[]string{
				"http_out_host",
				"http_out_port",
			}),
		ReqTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "http_out_requests_total",
				Help:        "Total amount of outgoing requests on each HTTP connection",
				ConstLabels: constLabels,
			},
			[]string{
				"http_out_host",
				"http_out_port",
				"http_out_method",
				"http_out_url",
			}),
		ReqBytesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "http_out_request_bytes_total",
				Help:        "Size of outgoing request on each HTTP connection",
				ConstLabels: constLabels,
			},
			[]string{
				"http_out_host",
				"http_out_port",
				"http_out_method",
				"http_out_url",
			}),
		ReqErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "http_out_request_errors_total",
				Help:        "Amount of errors on outgoing request on each HTTP connection",
				ConstLabels: constLabels,
			},
			[]string{
				"http_out_host",
				"http_out_port",
				"http_out_method",
				"http_out_url",
			}),
		RespTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "http_out_responses_total",
				Help:        "Total amount of responses on each request",
				ConstLabels: constLabels,
			},
			[]string{
				"http_out_host",
				"http_out_port",
				"http_out_method",
				"http_out_url",
				"http_out_response_code",
			}),
		RespBytesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "http_out_response_bytes_total",
				Help:        "Content length or response size",
				ConstLabels: constLabels,
			},
			[]string{
				"http_out_host",
				"http_out_port",
				"http_out_method",
				"http_out_url",
				"http_out_response_code",
			}),
		RespErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "http_out_response_errors_total",
				Help:        "Total amount of outgoing errors on each response",
				ConstLabels: constLabels,
			},
			[]string{
				"http_out_host",
				"http_out_port",
				"http_out_method",
				"http_out_url",
				"http_out_response_code",
			}),
		RespTimeToFirstByte: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        "http_out_response_first_byte_time",
				Help:        "Time that spends on getting first byte from the remote endpoint",
				ConstLabels: constLabels,
			},
			[]string{
				"http_out_host",
				"http_out_port",
				"http_out_method",
				"http_out_url",
				"http_out_response_code",
			}),
		RespTimeTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        "http_out_response_time_total",
				Help:        "Total time that spends on getting data from the remote endpoint",
				ConstLabels: constLabels,
			},
			[]string{
				"http_out_host",
				"http_out_port",
				"http_out_method",
				"http_out_url",
				"http_out_response_code",
			}),
	}
}

var httpOutOnce sync.Once

func (h *HTTPOut) AutoRegister() *HTTPOut {
	httpOutOnce.Do(func() {
		h.mustRegister(prometheus.DefaultRegisterer)
	})
	return h
}

func (h *HTTPOut) mustRegister(registerer prometheus.Registerer) {
	registerer.MustRegister(
		h.DNSResolveTime,
		h.HandshakeTime,
		h.ConnectTime,
		h.ReqTotal,
		h.ReqBytesTotal,
		h.ReqErrorsTotal,
		h.RespTotal,
		h.RespBytesTotal,
		h.RespErrorsTotal,
		h.RespTimeToFirstByte,
		h.RespTimeTotal,
	)
}

func (h *HTTPOut) DoAndCollect(c *http.Client, req *http.Request, dest interface{}) (*http.Response, error) {
	labelValues := []string{req.URL.Hostname(), req.URL.Port(), req.Method, replaceAccounts(replaceUUIDs(req.URL.Path))}
	h.ReqTotal.WithLabelValues(labelValues...).Inc()
	if req.ContentLength > 0 {
		h.ReqBytesTotal.WithLabelValues(labelValues...).Add(float64(req.ContentLength))
	}
	started := time.Now()
	resp, err := c.Do(req)
	if err != nil {
		h.ReqErrorsTotal.WithLabelValues(labelValues...).Inc()
		h.RespTotal.WithLabelValues(append(labelValues, err.Error())...).Inc()
		h.RespTimeTotal.WithLabelValues(append(labelValues, err.Error())...).Add(time.Since(started).Seconds())
		return resp, err
	}
	if resp.StatusCode != http.StatusOK {
		h.ReqErrorsTotal.WithLabelValues(labelValues...).Inc()
		h.RespTotal.WithLabelValues(append(labelValues, resp.Status)...).Inc()
		h.RespTimeTotal.WithLabelValues(append(labelValues, resp.Status)...).Add(time.Since(started).Seconds())
		b, e := ioutil.ReadAll(resp.Body)
		if e != nil {
			return nil, fmt.Errorf("err reading response body: %w", e)
		}
		return nil, fmt.Errorf("err unexpected response, code: %d, resp: %s, req: %s", resp.StatusCode, string(b), req.URL.String())
	}
	labelValues = append(labelValues, strconv.Itoa(resp.StatusCode))
	h.RespTotal.WithLabelValues(labelValues...).Inc()
	h.RespTimeTotal.WithLabelValues(labelValues...).Add(time.Since(started).Seconds())
	cr := &common.CountingReader{Reader: resp.Body}
	if err = json.NewDecoder(cr).Decode(dest); err != nil {
		h.RespErrorsTotal.WithLabelValues(labelValues...).Inc()
		err = fmt.Errorf("err decoding response: %w", err)
	}
	h.RespBytesTotal.WithLabelValues(labelValues...).Add(float64(cr.BytesRead))
	return resp, err
}
