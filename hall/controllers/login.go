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
	"strings"
)

func Login(w http.ResponseWriter, r *http.Request) {
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
		"errcode": 0,
		"errmsg":  "ok",
	}
	if client, closeTransportHandler, err := tools.GetRpcClient(common.HallConf.AccountHost, strconv.Itoa(common.HallConf.AccountPort)); err == nil {
		defer func() {
			if err := closeTransportHandler(); err != nil {
				logs.Error("close rpc err: %v", err)
			}
		}()
		ip := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0])
		if res, err := client.GetUserInfoByToken(context.Background(), token); err == nil {
			if res.Code == rpc.ErrorCode_Success {
				ret = map[string]interface{}{
					"errcode":     0,
					"errmsg":      "ok",
					"account":     res.UserObj.NickName,
					"userid":      res.UserObj.UserId,
					"name":        res.UserObj.NickName,
					"headimg":     res.UserObj.HeadImg,
					"lv":          res.UserObj.Lv,
					"exp":         res.UserObj.Exp,
					"coins":       res.UserObj.Gems,
					"vip":         res.UserObj.Vip,
					"money":       res.UserObj.Gems,
					"gems":        res.UserObj.Gems,
					"ip":          ip,
					"sex":         res.UserObj.Sex,
					"RenameCount": res.UserObj.ReNameCount,
					"ReHeadCount": res.UserObj.ReHeadCount,
					"item": map[string]int64{
						"ice": res.UserObj.Ice,
					},
				}
			}
		} else {
			logs.Error("call rpc Login err: %v", err)
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
