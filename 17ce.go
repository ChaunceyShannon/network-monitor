package main

import "encoding/json"

type httpLogData17ce struct {
	Rt    int    `json:"rt"`
	Error string `json:"error"`
	Txnid int    `json:"txnid"`
	Type  string `json:"type"`
	Data  struct {
		TaskID      string  `json:"TaskId"`
		ErrMsg      string  `json:"ErrMsg"`
		NodeID      int     `json:"NodeID"`
		HTTPCode    int     `json:"HttpCode"`
		TotalTime   float64 `json:"TotalTime"`
		FileSize    int     `json:"FileSize"`
		RealSize    int     `json:"RealSize"`
		TTFBTime    float64 `json:"TTFBTime"`
		DownTime    float64 `json:"DownTime"`
		ConnectTime float64 `json:"ConnectTime"`
		NsLookup    float64 `json:"NsLookup"`
		HTTPHead    string  `json:"HttpHead"`
		HTTPBodyMd5 string  `json:"HttpBodyMd5"`
		SrcIP       string  `json:"SrcIP"`
		NodeInfo    struct {
			IP     string `json:"ip"`
			Area   string `json:"area"`
			Isp    string `json:"isp"`
			ProID  string `json:"pro_id"`
			CityID string `json:"city_id"`
		} `json:"NodeInfo"`
		Srcip struct {
			Srcip     string `json:"srcip"`
			SrcipFrom string `json:"srcip_from"`
		} `json:"srcip"`
	} `json:"data"`
}

type l7ceStruct struct {
	Email    string
	ApiKey   string
	NodeType []string
	ISP      []string
	Province []string
	City     []string
}

type l7ceNodeInfo struct {
	ID      string // 节点ID
	AreaID  string // 区域ID
	ISPID   string // 运营商ID
	ProID   string // 省份ID
	CityID  string // 城市ID
	IP      string // IP地址
	Alive   bool   // 是否正常返回数据
	NodeNum int    // 上次测试这个地区使用了几个节点
}

func getNodes17ce(s string) (nodes []l7ceNodeInfo) {
	s1 := jsonLoads(s)
	s2 := toStringMap(s1["data"])
	var s3 map[string]interface{}
	for _, v := range s2 {
		s3 = toStringMap(v)
	}
	s4 := toStringMap(s3["nodes"])
	for nodeid, v := range s4 {
		lg.trace("Node ID:", nodeid)
		s5 := toStringMap(v)
		for _, v := range s5 {
			s6 := toStringMapString(v)
			nodes = append(nodes, l7ceNodeInfo{
				ID:      nodeid,
				AreaID:  s6["area"],
				ISPID:   s6["isp"],
				ProID:   s6["pro_id"],
				CityID:  s6["city_id"],
				IP:      s6["ip"],
				Alive:   false,
				NodeNum: 1,
			})
		}
	}
	return
}

func (c *l7ceStruct) init(apiAccount map[string]string, cfg map[string]string) {
	c.NodeType = strSplit(cfg["type"], ",")
	c.ISP = strSplit(cfg["isp"], ",")
	c.Province = strSplit(cfg["province"], ",")
	c.City = strSplit(cfg["city"], ",")
	c.Email = apiAccount["17ceEmail"]
	c.ApiKey = apiAccount["17ceApiKey"]
}

func (c *l7ceStruct) test(item Item) (httpLogArr []httpLog) {
	// 准备url
	user := c.Email
	api_pwd := c.ApiKey
	ut := str(now())
	code := md5sum(base64Encode((md5sum(api_pwd)[4:23] + user + ut)))
	websocketURL := "wss://wsapi.17ce.com:8001/socket/?ut=" + ut + "&code=" + code + "&user=" + user

	// 准备请求的数据
	requestJson := map[string]interface{}{
		"txnid":             now(),
		"nodetype":          nodeTypeArray2idArray17ce(c.NodeType),
		"num":               1,
		"Url":               item.Schema + "://" + item.Domain + item.Path,
		"TestType":          "HTTP",
		"TimeOut":           8,
		"Request":           item.Method,
		"NoCache":           true,
		"Cookie":            "",
		"Trace":             false,
		"UserAgent":         "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:90.0) Gecko/20100101 Firefox/90.0",
		"FollowLocation":    2,
		"GetMD5":            true,
		"GetResponseHeader": true,
		"MaxDown":           1048576,
		"AutoDecompress":    true,
		"type":              1,
		"isps":              ispArray2idArray17ce(c.ISP),
		"pro_ids":           provinceArray2idArray17ce(c.Province),
		"areas":             []int{1},
	}

	if len(cityArray2idArray17ce(c.City)) != 0 {
		requestJson["city_ids"] = cityArray2idArray17ce(c.City)
	}

	for _, header := range item.Header {
		if header.Key == "Referer" {
			requestJson["Referer"] = header.Value
		} else if header.Key == "User-Agent" {
			requestJson["UserAgent"] = header.Value
		} else if header.Key == "Cookie" {
			requestJson["Cookie"] = header.Value
		} else if header.Key == "Host" {
			requestJson["Host"] = header.Value
		}
	}

	lg.trace("连接17ce的API服务器")
	websocket := getWebSocket(websocketURL)
	defer websocket.close()

	lg.trace("发送请求：", requestJson)
	websocket.send(jsonDumps(requestJson))

	// 保存当次有参加检测的Node的id
	// 跟17ce的技术沟通之后，说是要保存Node的信息，如果没有返回则再次针对这个省市运营商发起一个测试
	var nodes []l7ceNodeInfo

	lg.trace("开始收取测试日志")
	for {
		msg := websocket.recv(300)
		lg.trace(msg)
		jm := getXPathJson(msg)
		statusCode := toInt(jm.first("//rt").text())
		if statusCode != 1 {
			panicerr(id2error17ce(statusCode))
		}

		if !jm.exists("//type") {
			continue
		}

		dataType := jm.first("//type").text()
		if dataType == "TaskEnd" {
			break
		}

		if dataType == "TaskAccept" {
			nodes = getNodes17ce(msg)
		}

		if dataType == "NewData" {
			var newDataLog httpLogData17ce
			err := json.Unmarshal([]byte(msg), &newDataLog)
			panicerr(err)

			for k := range nodes {
				if toInt(nodes[k].ID) == newDataLog.Data.NodeID {
					nodes[k].Alive = true
				}
			}
			httpLogArr = append(httpLogArr, httpLog{
				NodeIP:   newDataLog.Data.NodeInfo.IP,
				Province: id2province17ce(toInt(newDataLog.Data.NodeInfo.ProID)),
				City:     id2city17ce(toInt(newDataLog.Data.NodeInfo.CityID)),
				ISP:      id2isp17ce(toInt(newDataLog.Data.NodeInfo.Isp)),

				ServerLocation: newDataLog.Data.Srcip.SrcipFrom,
				ServerIP:       newDataLog.Data.Srcip.Srcip,

				NsLookup:  newDataLog.Data.NsLookup,
				Connect:   newDataLog.Data.ConnectTime,
				FirstByte: newDataLog.Data.TTFBTime,
				Download:  newDataLog.Data.DownTime,
				Total:     newDataLog.Data.TotalTime,
				Code:      newDataLog.Data.HTTPCode,
				Rate:      int64(toFloat64(newDataLog.Data.RealSize) / newDataLog.Data.DownTime),

				ErrMsg: newDataLog.Data.ErrMsg,
			})
		}
	}

	// 一轮的检测过后，应该会有一些节点是掉线的，没有返回数据，需要再次发起查询请求
	for {
		// 如果还有掉线的，就继续，否则break
		if func() bool {
			for _, n := range nodes {
				if !n.Alive {
					lg.trace("有一些节点是掉线的")
					return false
				}
			}
			return true
		}() {
			break
		}

		// 针对每一个掉线的，重查
		for k := range nodes {
			if !nodes[k].Alive {
				lg.trace("重查一次:", id2isp17ce(toInt(nodes[k].ISPID)), id2province17ce(toInt(nodes[k].ProID)), id2city17ce(toInt(nodes[k].CityID)))
				requestJson["txnid"] = now()
				requestJson["isps"] = []int{toInt(nodes[k].ISPID)}
				requestJson["pro_ids"] = []int{toInt(nodes[k].ProID)}
				requestJson["city_ids"] = []int{toInt(nodes[k].CityID)}

				// 由于设置为1会总是没反悔，说是要设置多几个，合理推测那节点顺序在一定时间内不会改变
				// 那估计可以累加，加到有1条返回为止，最后消耗的是1个积分。
				nodes[k].NodeNum += 1
				requestJson["num"] = nodes[k].NodeNum
				lg.trace("设置节点个数为:", nodes[k].NodeNum)

				lg.trace("发送请求：", requestJson)
				websocket.send(jsonDumps(requestJson))

				for {
					msg := websocket.recv(300)
					lg.trace(msg)
					jm := getXPathJson(msg)
					statusCode := toInt(jm.first("//rt").text())
					if statusCode != 1 {
						panicerr(id2error17ce(statusCode))
					}

					if !jm.exists("//type") {
						continue
					}

					dataType := jm.first("//type").text()
					if dataType == "TaskEnd" {
						break
					}

					if dataType == "NewData" {
						var newDataLog httpLogData17ce
						err := json.Unmarshal([]byte(msg), &newDataLog)
						panicerr(err)

						nodes[k].Alive = true

						httpLogArr = append(httpLogArr, httpLog{
							NodeIP:   newDataLog.Data.NodeInfo.IP,
							Province: id2province17ce(toInt(newDataLog.Data.NodeInfo.ProID)),
							City:     id2city17ce(toInt(newDataLog.Data.NodeInfo.CityID)),
							ISP:      id2isp17ce(toInt(newDataLog.Data.NodeInfo.Isp)),

							ServerLocation: newDataLog.Data.Srcip.SrcipFrom,
							ServerIP:       newDataLog.Data.Srcip.Srcip,

							NsLookup:  newDataLog.Data.NsLookup,
							Connect:   newDataLog.Data.ConnectTime,
							FirstByte: newDataLog.Data.TTFBTime,
							Download:  newDataLog.Data.DownTime,
							Total:     newDataLog.Data.TotalTime,
							Code:      newDataLog.Data.HTTPCode,
							Rate:      int64(toFloat64(newDataLog.Data.RealSize) / newDataLog.Data.DownTime),

							ErrMsg: newDataLog.Data.ErrMsg,
						})
					}
				}
			}
		}
	}

	return
}
