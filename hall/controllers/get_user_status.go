package controllers

import (
	"context"
	"encoding/json"
	"fish/common/api/thrift/gen-go/rpc"
	"fish/common/tools"
	"fish/hall/common"
	"github.com/astaxie/beego/logs"
	"net/http"
	"strconv"
)

func GetUserStatus(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("GetUserInfo panic:%v ", r)
		}
	}()
	//logs.Debug("new request url:[%s]",r.URL)
	account := r.FormValue("account")
	if len(account) == 0 {
		return
	}
	token := r.FormValue("sign")
	if len(token) == 0 {
		return
	}
	ret := map[string]interface{}{
		"errcode": 1,
		"errmsg":  "failed",
	}
	if client, closeTransportHandler, err := tools.GetRpcClient(common.HallConf.AccountHost, strconv.Itoa(common.HallConf.AccountPort)); err == nil {
		defer func() {
			if err := closeTransportHandler(); err != nil {
				logs.Error("close rpc err: %v", err)
			}
		}()
		if res, err := client.GetUserInfoByToken(context.Background(), token); err == nil {
			if res.Code == rpc.ErrorCode_Success {
				ret = map[string]interface{}{
					"errcode": 0,
					"errmsg":  "ok",
					"gems":    res.UserObj.Gems,
				}
			}
		} else {
			logs.Error("call rpc GetUserStatus err: %v", err)
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
