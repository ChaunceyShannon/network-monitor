package main

type httpLog struct {
	NodeIP   string // 测速节点的IP
	Province string // 测速节点的省份
	City     string // 测速节点的城市
	ISP      string // 测速节点的运营商

	ServerLocation string // 服务器的位置
	ServerIP       string // 服务器的IP

	NsLookup  float64 // 域名解析时间
	Connect   float64 // 连接时间
	FirstByte float64 // 收到首字节时间， 从请求开始到响应开始之间所用的时间
	Download  float64 // 下载时间
	Total     float64 // 总时间
	Code      int     // HTTP 状态码
	Rate      int64   // 下载速度，每秒的字节数

	ErrMsg string // 为空则正常，否则出错
}

type thirdPartAPI interface {
	init(map[string]string, map[string]string)
	test(Item) []httpLog
}
