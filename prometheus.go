package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//单个定义指标
var (
	//自定义cpu温度数据手机(测量指标)
	cpuTemp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_temperature_celsius",
		Help: "Current temperature of the CPU.",
	})
	//自定义硬盘错误的数量(累加指标)
	hdFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hd_errors_total",
			Help: "Number of hard-disk errors.",
		},
		[]string{"device"}, //定义字段
	)
)

func init() {
	// 必须注册才能使用：
	prometheus.MustRegister(cpuTemp)
	prometheus.MustRegister(hdFailures)
}

func main() {
	//(测量指标)设置值
	cpuTemp.Set(65.3)
	//(累加指标)累加
	hdFailures.With(prometheus.Labels{"device": "/dev/sda"}).Inc()
	hdFailures.With(prometheus.Labels{"device": "/dev/sda"}).Inc()
	//设定路由1
	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(":8080", nil))
}
