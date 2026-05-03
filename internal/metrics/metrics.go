package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/baditaflorin/calendar-ridicare-gunoi-arad/internal/domain"
)

type Metrics struct {
	registry       *prometheus.Registry
	HTTPRequests   *prometheus.CounterVec
	HTTPDuration   *prometheus.HistogramVec
	ETLRuns        *prometheus.CounterVec
	ETLDuration    prometheus.Histogram
	ETLLastSuccess prometheus.Gauge
	ParserIssues   prometheus.Gauge
	DBRows         *prometheus.GaugeVec
}

func New() *Metrics {
	registry := prometheus.NewRegistry()
	m := &Metrics{
		registry: registry,
		HTTPRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gunoi_arad_http_requests_total",
			Help: "Total HTTP requests by method, route, and status.",
		}, []string{"method", "route", "status"}),
		HTTPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "gunoi_arad_http_request_duration_seconds",
			Help:    "HTTP request latency by method and route.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route"}),
		ETLRuns: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gunoi_arad_etl_runs_total",
			Help: "Total ETL runs by result.",
		}, []string{"result"}),
		ETLDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "gunoi_arad_etl_duration_seconds",
			Help:    "ETL run duration.",
			Buckets: []float64{1, 2.5, 5, 10, 30, 60, 120},
		}),
		ETLLastSuccess: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gunoi_arad_etl_last_success_timestamp_seconds",
			Help: "Unix timestamp of the last successful ETL run.",
		}),
		ParserIssues: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gunoi_arad_parser_issues",
			Help: "Number of parser issues from the most recent imported dataset.",
		}),
		DBRows: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "gunoi_arad_db_rows",
			Help: "SQLite row counts by table.",
		}, []string{"table"}),
	}
	registry.MustRegister(m.HTTPRequests, m.HTTPDuration, m.ETLRuns, m.ETLDuration, m.ETLLastSuccess, m.ParserIssues, m.DBRows)
	return m
}

func (m *Metrics) Registry() *prometheus.Registry {
	return m.registry
}

func (m *Metrics) ObserveETL(result string, elapsed time.Duration, issues int) {
	m.ETLRuns.WithLabelValues(result).Inc()
	if elapsed > 0 {
		m.ETLDuration.Observe(elapsed.Seconds())
	}
	if result == "success" {
		m.ETLLastSuccess.Set(float64(time.Now().Unix()))
	}
	m.ParserIssues.Set(float64(issues))
}

func (m *Metrics) ObserveCounts(counts domain.Counts) {
	m.DBRows.WithLabelValues("places").Set(float64(counts.Places))
	m.DBRows.WithLabelValues("rules").Set(float64(counts.Rules))
	m.DBRows.WithLabelValues("events").Set(float64(counts.Events))
	m.DBRows.WithLabelValues("sources").Set(float64(counts.Sources))
	m.DBRows.WithLabelValues("parse_issues").Set(float64(counts.ParseIssues))
}

func StatusCodeLabel(status int) string {
	return strconv.Itoa(status)
}
