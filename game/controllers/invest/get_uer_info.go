package invest

import (
	"net/http"
	"github.com/astaxie/beego/logs"
	"game/invest/tools"
	"golang.org/x/net/context"
	"encoding/json"
	"game/api/thrift/gen-go/rpc"
	"time"
	"game/invest/common"
)

func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r:= recover(); r != nil {
			logs.Error("GetUserInfo panic:%v ",r)
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
		w.Write(data)
	}()

	token := r.FormValue("token")
	if len(token) == 0 {
		token = r.PostFormValue("token")
	}
	rpcClient, closeTransport, err := tools.GetRpcClient()
	if err != nil {
		logs.Debug("get rpc client err:%v", err)
		ret["code"] = common.ErrorRpcServerError
		return
	}
	defer closeTransport()

	resp, err := rpcClient.GetUserInfoByken(context.Background(), token)
	//logs.Debug("user info %v",resp)
	if err != nil || resp.Code != rpc.ErrorCode_Success {
		logs.Debug("check user token [%v] failed err:%v, resp:%v", token, err, resp)
		ret["code"] = common.ErrorAuthFailed
	} else {
		ret["code"] = int(resp.Code)
		identityObj := map[string]interface{}{
			"userId":     resp.UserObj.UserId,
			"nickName":   resp.UserObj.NickName,
			"avatarAuto": resp.UserObj.AvatarAuto,
			"sex":        1,
			"gold":       resp.UserObj.Gold,
			"goldenBean": resp.UserObj.Gold,
			"diamond":    100000,
			"level":      12,
			"vipLevel":   1,
		}
		ret["result"] = map[string]interface{}{
			"identity_obj": identityObj,
		}
	}
}
