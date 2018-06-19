package main

import (
	"math/rand"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

//先定义结构体，这是一个集群的指标采集器，每个集群都有自己的Zone,代表集群的名称。另外两个是保存的采集的指标
type ClusterManager struct {
	Zone         string
	OOMCountDesc *prometheus.Desc
	RAMUsageDesc *prometheus.Desc
	// ... many more fields
}

//实现一个采集工作,
//放到了ReallyExpensiveAssessmentOfTheSystemState函数中实现，
//每次执行的时候，返回一个按照主机名作为键采集到的数据，
//两个返回值分别代表了OOM错误计数，和RAM使用指标信息。
func (c *ClusterManager) ReallyExpensiveAssessmentOfTheSystemState() (
	oomCountByHost map[string]int, ramUsageByHost map[string]float64,
) {
	oomCountByHost = map[string]int{
		"foo.example.org": int(rand.Int31n(1000)),
		"bar.example.org": int(rand.Int31n(1000)),
	}
	ramUsageByHost = map[string]float64{
		"foo.example.org": rand.Float64() * 100,
		"bar.example.org": rand.Float64() * 100,
	}
	return
}

//实现Describe接口，传递指标描述符到channel
func (c *ClusterManager) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.OOMCountDesc
	ch <- c.RAMUsageDesc
}

//Collect函数将执行抓取函数并返回数据，
//返回的数据传递到channel中，
//并且传递的同时绑定原先的指标描述符。以及指标的类型（一个Counter和一个Guage）
func (c *ClusterManager) Collect(ch chan<- prometheus.Metric) {
	oomCountByHost, ramUsageByHost := c.ReallyExpensiveAssessmentOfTheSystemState()
	for host, oomCount := range oomCountByHost {
		ch <- prometheus.MustNewConstMetric(
			c.OOMCountDesc,
			prometheus.CounterValue,
			float64(oomCount),
			host,
		)
	}
	for host, ramUsage := range ramUsageByHost {
		ch <- prometheus.MustNewConstMetric(
			c.RAMUsageDesc,
			prometheus.GaugeValue,
			ramUsage,
			host,
		)
	}
}

//创建结构体及对应的指标信息,
//NewDesc参数第一个为指标的名称，
//第二个为帮助信息，
//显示在指标的上面作为注释，
//第三个是定义的label名称数组，
//第四个是定义的Labels
func NewClusterManager(zone string) *ClusterManager {
	return &ClusterManager{
		Zone: zone,
		OOMCountDesc: prometheus.NewDesc(
			"clustermanager_oom_crashes_total",
			"Number of OOM crashes.",
			[]string{"host"},
			prometheus.Labels{"zone": zone},
		),
		RAMUsageDesc: prometheus.NewDesc(
			"clustermanager_ram_usage_bytes",
			"RAM usage as reported to the cluster manager.",
			[]string{"host"},
			prometheus.Labels{"zone": zone},
		),
	}
}

func ExampleCollector() {
	//创建收集器
	workerDB := NewClusterManager("db")
	workerCA := NewClusterManager("ca")

	// 自定义收集器
	reg := prometheus.NewPedanticRegistry()
	// 实现收集器
	reg.MustRegister(workerDB)
	reg.MustRegister(workerCA)
	//自定义收集+默认收集
	gatherers := prometheus.Gatherers{
		prometheus.DefaultGatherer,
		reg,
	}
	// 定义http.handle
	h := promhttp.HandlerFor(gatherers,
		promhttp.HandlerOpts{
			ErrorLog:      log.NewErrorLogger(),
			ErrorHandling: promhttp.ContinueOnError,
		})
	//路由
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
	log.Infoln("Start server at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Errorf("Error occur when start server %v", err)
		os.Exit(1)
	}
}
