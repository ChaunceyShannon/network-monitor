package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Header struct {
	Key   string
	Value string
}

type Item struct {
	Method string   // http method
	Schema string   // http or https
	Domain string   // Domain name
	Path   string   // URL path
	Data   string   // Post data
	Header []Header // HTTP headers to add
}

var apis = map[string]thirdPartAPI{
	"17ce": &l7ceStruct{},
}

func loadItem(path string) map[string]Item {
	yfile, err := ioutil.ReadFile(path)
	panicerr(err)

	items := make(map[string]Item)

	err = yaml.Unmarshal(yfile, &items)
	panicerr(err)

	return items
}

func main() {
	// 把整个ini配置文件映射为map
	icfg := argparser("").parseArgs().cfg
	args := make(map[string]map[string]string)
	for _, section := range icfg.SectionStrings() {
		args[section] = make(map[string]string)
		for _, k := range icfg.Section(section).KeyStrings() {
			args[section][k] = icfg.Section(section).Key(k).Value()
		}
	}

	for name, cfg := range args {
		if name == "api" {
			continue
		}
		if cfg["enable"] == "true" {
			lg.trace("配置\"" + name + "\"已启用")
			lg.trace("调用的API是:", cfg["api"])
			run(apis[cfg["api"]], name, args["api"], cfg, loadItem(cfg["item"]))
		}
	}
}
