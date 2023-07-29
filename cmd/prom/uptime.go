package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type uptimeCollector time.Time

var (
	startTimeDesc = prometheus.NewDesc(
		"node_start_time_seconds",
		"Start time of the node since unix epoch in seconds.",
		nil,
		nil,
	)
	currentTimeDesc = prometheus.NewDesc(
		"node_current_time_seconds",
		"Current time of the node since unix epoch in seconds.",
		nil,
		nil,
	)
)

func NewUptimeCollector() prometheus.Collector {
	return uptimeCollector(time.Now())
}

func (c uptimeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- startTimeDesc
	ch <- currentTimeDesc
}

func (c uptimeCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		startTimeDesc,
		prometheus.CounterValue,
		float64(time.Time(c).UnixMilli())/1000,
	)
	ch <- prometheus.MustNewConstMetric(
		currentTimeDesc,
		prometheus.CounterValue,
		float64(time.Now().UnixMilli())/1000,
	)
}
