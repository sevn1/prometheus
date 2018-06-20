package main

import (
	"bytes"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

//自定义采集器
type statusCollect struct {
	reqDesc       *prometheus.CounterVec
	respsizeDesc  *prometheus.Desc
	respsizevalue int64
}

func (s *statusCollect) ReqAdd(code, method string) {
	s.reqDesc.WithLabelValues(code, method).Inc()
}

func (s *statusCollect) ReqSizeAdd(size int64) {
	atomic.AddInt64(&s.respsizevalue, size)
}

//实现Describe接口，传递指标描述符到channel
func (s *statusCollect) Describe(ch chan<- *prometheus.Desc) {
	ch <- s.respsizeDesc
	s.reqDesc.Describe(ch)
}

//Collect函数将执行抓取函数并返回数据，
func (s *statusCollect) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(s.respsizeDesc, prometheus.CounterValue, float64(s.respsizevalue))
	s.reqDesc.Collect(ch)
}

//声明采集器
func NewStatusCollect() *statusCollect {
	opts := prometheus.CounterOpts{Namespace: "Test", Subsystem: "http", Name: "request", Help: "requst count"}
	return &statusCollect{
		reqDesc:      prometheus.NewCounterVec(opts, []string{"code", "method"}),
		respsizeDesc: prometheus.NewDesc("Namespace_http_respsize_count", "http respsize count", nil, nil),
	}
}

func main() {
	//定义采集器
	status := NewStatusCollect()
	regist := prometheus.NewRegistry()
	regist.MustRegister(status)

	http.HandleFunc("/metric", func(w http.ResponseWriter, r *http.Request) {
		status.ReqAdd("200", strings.ToLower(r.Method))
		//实现收集器
		entry, err := regist.Gather()
		if err != nil {
			http.Error(w, "An error has occurred during metrics collection:\n\n"+err.Error(), http.StatusInternalServerError)
			return
		}

		buf := bytes.NewBuffer(nil)
		contentType := expfmt.Negotiate(r.Header)
		enc := expfmt.NewEncoder(buf, contentType)

		for _, met := range entry {
			if err := enc.Encode(met); err != nil {
				http.Error(w, "An error has occurred during metrics encoding:\n\n"+err.Error(), http.StatusInternalServerError)
				// return
			}
		}

		if buf.Len() == 0 {
			http.Error(w, "No metrics encoded, last error:\n\n"+err.Error(), http.StatusInternalServerError)
			return
		}
		status.ReqSizeAdd(int64(buf.Len()))
		w.Write(buf.Bytes())
	})

	http.ListenAndServe(":8081", nil)
}
