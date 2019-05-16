package invest

import (
	"net/http"
	"github.com/astaxie/beego/logs"
	"time"
	"encoding/json"
	"game/invest/tools"
	"game/invest/common"
	"game/api/thrift/gen-go/rpc"
	"context"
	"game/invest/service"
)

func RecordList(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("RewardList panic:%v ", r)
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

		var stakeList []common.InvestUserStake

		err = service.GameConf.MysqlConf.Pool.Select(&stakeList, "select * from game_invest_user_stake where user_id=?", resp.UserObj.UserId)
		if err != nil {
			logs.Error("select user [%d] stake err:%v", resp.UserObj.UserId, err)
		}
		userStakeList := []map[string]interface{}{}
		for _, stakeItem := range stakeList {
			userStakeList = append(userStakeList, map[string]interface{}{
				"record_id":     stakeItem.Periods,
				"get_gold":      stakeItem.WinGold,
				"game_times_id": stakeItem.GameTimesId,
				"stake_time":    stakeItem.StakeTime,
				"stake_gold":    stakeItem.UserAllStake,
			})
		}
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
			"record_list":  userStakeList,
		}
	}
}
