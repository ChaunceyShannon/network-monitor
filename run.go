package main

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metricsStruct struct {
	NsLookup  *prometheus.GaugeVec
	Connect   *prometheus.GaugeVec
	FirstByte *prometheus.GaugeVec
	Download  *prometheus.GaugeVec
	Total     *prometheus.GaugeVec
	Rate      *prometheus.GaugeVec

	Error *prometheus.CounterVec
}

var metrics metricsStruct

func startMetricServer() {
	metrics = metricsStruct{}

	metrics.NsLookup = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nslookup_time",
		Help: "DNS查询花费的时间",
	}, []string{"cfg", "name", "province", "city", "isp", "server_location"})
	prometheus.MustRegister(metrics.NsLookup)

	metrics.Connect = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connect_time",
		Help: "TCP连接花费的时间",
	}, []string{"cfg", "name", "province", "city", "isp", "server_location"})
	prometheus.MustRegister(metrics.Connect)

	metrics.FirstByte = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "first_byte_time",
		Help: "从请求开始到收到首字节所花费的时间",
	}, []string{"cfg", "name", "province", "city", "isp", "server_location"})
	prometheus.MustRegister(metrics.FirstByte)

	metrics.Download = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "download_time",
		Help: "下载数据花费的时间",
	}, []string{"cfg", "name", "province", "city", "isp", "server_location"})
	prometheus.MustRegister(metrics.Download)

	metrics.Total = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "total_time",
		Help: "总共花费的时间",
	}, []string{"cfg", "name", "province", "city", "isp", "server_location"})
	prometheus.MustRegister(metrics.Total)

	metrics.Rate = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rate_time",
		Help: "下载的速度，单位是byte/s",
	}, []string{"cfg", "name", "province", "city", "isp", "server_location"})
	prometheus.MustRegister(metrics.Rate)

	metrics.Error = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "error_count",
		Help: "监控节点回报错误的次数",
	}, []string{"cfg", "message"})
	prometheus.MustRegister(metrics.Error)

	ginEng := gin.New()

	ginEng.GET("/metrics", gin.WrapH(promhttp.Handler()))
	go ginEng.Run("0.0.0.0:8080")
}

func run(tpa thirdPartAPI, cfgName string, apiAccount map[string]string, cfg map[string]string, items map[string]Item) {
	startMetricServer()

	metricsMap := make(map[string][]prometheus.Labels)

	tpa.init(apiAccount, cfg)
	for {
		for name, item := range items {
			lg.trace("开始测试条目:", name)

			httpLogResult := tpa.test(item)
			// 清理之前的这个name的条目
			for _, label := range metricsMap[name] {
				metrics.NsLookup.Delete(label)
				metrics.Connect.Delete(label)
				metrics.FirstByte.Delete(label)
				metrics.Download.Delete(label)
				metrics.Total.Delete(label)
				metrics.Rate.Delete(label)
			}
			metricsMap[name] = []prometheus.Labels{}

			// 更新到metrics
			for _, tl := range httpLogResult {
				lg.trace(tl)

				if tl.ErrMsg != "" {
					metrics.Error.With(prometheus.Labels{"cfg": cfgName, "message": tl.ErrMsg})
					continue
				}

				label := prometheus.Labels{
					"cfg":             cfgName,
					"name":            name,
					"province":        tl.Province,
					"city":            tl.City,
					"isp":             tl.ISP,
					"server_location": tl.ServerLocation,
					// "server_ip":       tl.ServerIP,
				}
				metricsMap[name] = append(metricsMap[name], label)

				metrics.NsLookup.With(label).Set(tl.NsLookup)
				metrics.Connect.With(label).Set(tl.Connect)
				metrics.FirstByte.With(label).Set(tl.FirstByte)
				metrics.Download.With(label).Set(tl.Download)
				metrics.Total.With(label).Set(tl.Total)
				metrics.Rate.With(label).Set(float64(tl.Rate))
			}
		}

		lg.trace("休眠" + str(cfg["sleep"]) + "秒")
		sleep(toInt(cfg["sleep"]))
	}
}
