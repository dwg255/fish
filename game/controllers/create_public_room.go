package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"net/http"
	"time"
)

func CreatePublicRoom(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("CreatePublicRoom panic:%v ", r)
		}
	}()
	//logs.Debug("new request url:[%s]",r.URL)
	ret := make(map[string]interface{})
	ret["time"] = time.Now().Format("2006-01-02 15:04:05")
	defer func() {
		data, err := json.Marshal(ret)
		if err != nil {
			logs.Error("json marsha1 failed err:%v", err)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if _, err := w.Write(data); err != nil {
			logs.Error("CreateRoom err: %v",err)
		}
	}()
}
