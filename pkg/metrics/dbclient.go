package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type DBClient struct {
	Connections  prometheus.GaugeFunc
	ErrsTotal    *prometheus.CounterVec
	TimeTotal    *prometheus.CounterVec
	BytesTotal   *prometheus.CounterVec
	RecordsTotal *prometheus.CounterVec
}

type ConnectionStat interface{}

func NewDBClient(db, host, port string, fn func() float64) *DBClient {
	constLabels := prometheus.Labels{"db": db, "db_host": host, "db_port": port}
	return &DBClient{
		Connections: prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name:        "db_client_connections_total",
			Help:        "Amount of open connections",
			ConstLabels: constLabels,
		}, fn),
		ErrsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "db_client_query_errors_total",
			Help:        "How many requests send to remote http service",
			ConstLabels: constLabels,
		}, []string{"db_query"}),
		TimeTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "db_client_query_time_total",
			Help:        "How many requests send to remote http service",
			ConstLabels: constLabels,
		}, []string{"db_query"}),
		BytesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "db_client_query_bytes_total",
			Help:        "How many requests send to remote http service",
			ConstLabels: constLabels,
		}, []string{"db_query"}),
		RecordsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "db_client_query_records_total",
			Help:        "How many requests send to remote http service",
			ConstLabels: constLabels,
		}, []string{"db_query"}),
	}
}

var dbClientOnce sync.Once

func (c *DBClient) AutoRegister() *DBClient {
	dbClientOnce.Do(func() {
		c.mustRegister(prometheus.DefaultRegisterer)
	})
	return c
}

func (c *DBClient) mustRegister(registerer prometheus.Registerer) *DBClient {
	registerer.MustRegister(c.Connections, c.ErrsTotal, c.TimeTotal, c.BytesTotal, c.RecordsTotal)
	return c
}
