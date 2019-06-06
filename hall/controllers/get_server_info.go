package controllers

import (
	"encoding/json"
	"fish/hall/common"
	"github.com/astaxie/beego/logs"
	"net/http"
	"strconv"
)

func GetServerInfo(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("GetUserInfo panic:%v ", r)
		}
	}()
	//logs.Debug("new request url:[%s]",r.URL)
	ret := map[string]interface{}{
		"appweb":  "please wait",
		"hall":    common.HallConf.HallHost + ":" + strconv.Itoa(common.HallConf.HallPort),
		"version": common.HallConf.Version,
	}
	defer func() {
		data, err := json.Marshal(ret)
		if err != nil {
			logs.Error("json marsha1 failed err:%v", err)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if _, err := w.Write(data); err != nil {
			logs.Error("CreateRoom err: %v", err)
		}
	}()
}
