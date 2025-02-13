package main

import (
	"log"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	// 定义指标
	totalSol = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gecko_proxy",
			Help: "gecko proxy",
		},
		[]string{"user", "is_error", "sort"},
	)

	page404 = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gecko_404",
			Help: "gecko 404 page",
		},
		[]string{"page", "sort"},
	)

	qps = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proxy_qps",
			Help: "proxy qps",
		},
		[]string{"method", "status"},
	)

	latencyHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_latency_seconds",
			Help:    "Histogram of request latencies",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status"},
	)
)

func init() {
	// 注册指标
	prometheus.MustRegister(totalSol)
	prometheus.MustRegister(page404)
	prometheus.MustRegister(qps)
	prometheus.MustRegister(latencyHistogram)
}

func emitGeckoProxyStatus(user, sort, isError string) {
	totalSol.WithLabelValues(user, isError, sort).Inc()
}

func emitQps(method, status string) {
	qps.WithLabelValues(method, status).Inc()
}

func emitLatency(latency float64, method, status string) {
	latencyHistogram.WithLabelValues(method, status).Observe(latency)
}

func emitPage404(page, sort string) {
	page404.WithLabelValues(page, sort).Inc()
}

func pushData(task string) {
	// 获取主机名
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal("Could not get hostname: ", err)
	}

	// 将指标推送到 Pushgateway
	err = push.New(pushGatewayURL, hostname+task).
		Collector(totalSol).
		Collector(page404).
		Collector(qps).
		Collector(latencyHistogram).
		Push()
	if err != nil {
		log.Printf("Could not push metrics to Pushgateway: %v", err)
	} else {
		log.Println("Pushed metrics to Pushgateway successfully.")
	}
}

func test() {
	// 模拟推送数据
	emitGeckoProxyStatus("user1", "sort1", "no_error")
	emitQps("GET", "200")
	emitLatency(0.123, "GET", "200")
	emitPage404("404", "sort2")

	// 推送数据到 Pushgateway
	pushData("task")
}
