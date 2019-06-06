package controllers

import (
	"context"
	"encoding/json"
	"fish/common/tools"
	"fish/hall/common"
	"github.com/astaxie/beego/logs"
	"net/http"
	"strconv"
)

func GetMessage(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("GetUserInfo panic:%v ", r)
		}
	}()
	messageType := r.FormValue("type")
	if len(messageType) == 0 {
		return
	}
	//logs.Debug("new request url:[%s]",r.URL)
	ret := map[string]interface{}{
		"errcode": 1,
		"errmsg":  "get message failed",
	}
	//logs.Debug("get rpc client %v:%v", common.HallConf.AccountHost, common.HallConf.AccountPort)
	if client, closeTransportHandler, err := tools.GetRpcClient(common.HallConf.AccountHost, strconv.Itoa(common.HallConf.AccountPort)); err == nil {
		defer func() {
			if err := closeTransportHandler(); err != nil {
				logs.Error("close rpc err: %v", err)
			}
		}()
		if res, err := client.GetMessage(context.Background(), messageType); err == nil {
			ret = map[string]interface{}{
				"errcode": 0,
				"errmsg":  "ok",
				"msg":     res,
				"version": common.HallConf.Version,
			}
		} else {
			logs.Error("call rpc GetMessage err: %v", err)
		}
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
