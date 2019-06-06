package main

import (
	"fish/hall/common"
	_ "fish/hall/router"
	"flag"
	"fmt"
	"github.com/astaxie/beego/logs"
	"net/http"
)

func main() {
	err := initConf()
	if err != nil {
		logs.Error("init conf err: %v", err)
		return
	}

	err = initSec()
	if err != nil {
		logs.Error("init sec err: %v", err)
		return
	}

	var addr = flag.String("addr", fmt.Sprintf(":%d", common.HallConf.HallPort), "http service address")
	logs.Debug("hall server listen port %v",*addr)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		logs.Error("ListenAndServe err: %v", err)
	}
}
