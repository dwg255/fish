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

func RecordDetail(w http.ResponseWriter, r *http.Request) {
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
	gameTimesId := r.FormValue("game_times_id")
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

		err = service.GameConf.MysqlConf.Pool.Select(&stakeList, "select * from game_invest_user_stake where user_id=? and game_times_id=?", resp.UserObj.UserId,gameTimesId)
		if err != nil {
			logs.Error("select user [%d] stake err:%v", resp.UserObj.UserId, err)
			return
		}
		if len(stakeList) != 1 {
			logs.Error("get user [%d],game_tiems_id[%v] record detail api err:%v",resp.UserObj.UserId,gameTimesId,err)
			ret["code"] = common.ErrorParamInvalid
			return
		}
		var detail = make([]map[string]interface{},0)
		var winItemData = make(map[string]interface{})
		var gameResult *common.InvestConf
		for _,investConf := range common.InvestConfArr{
			if investConf.Id == stakeList[0].GameResult {
				gameResult = investConf
				break
			}
		}
		stakeDetail := make(map[int]int)
		err = json.Unmarshal([]byte(stakeList[0].StakeDetail),&stakeDetail)
		if err != nil {
			logs.Error("game_times_id [%s] stake detail [%v],json unmarsha1 failed,err:%v",gameTimesId,stakeList[0].StakeDetail,err)
			ret["code"] = common.ErrorParamInvalid
			return
		}
		for position,gold := range stakeDetail{
			isWinPosition := false
			for _,investConf := range gameResult.ParentArr{
				if investConf.Id == position {
					detail = append(detail,map[string]interface{}{
						"item_id" : investConf.Id,
						"name" : investConf.Name,
						"stake_gold" : gold,
						"get_gold" : int(float32(gold) * investConf.Rate),
					})
					isWinPosition = true
				}
			}
			if !isWinPosition {
				for _,investConf := range common.InvestConfArr {
					if investConf.Id == position {
						detail = append(detail, map[string]interface{}{
							"item_id":    investConf.Id,
							"name":       investConf.Name,
							"stake_gold": gold,
							"get_gold":   0,
						})
						break
					}
				}
			}
		}

		if gameResult == nil {
			logs.Error("game_times_id [%s] game result [%d],not invalid",gameTimesId,stakeList[0].GameResult)
			ret["code"] = common.ErrorParamInvalid
			return
		}
		if len(gameResult.ParentArr) > 1 {
			//填坑， :-(
			for k, itemInvestConf := range gameResult.ParentArr {
				key := "third_item"
				if k == 0 {
					key = "third_item"
				} else if k == 1 {
					key = "second_item"
				} else if k == 2 {
					key = "first_item"
				}
				winItemData[key] = map[string]interface{}{
					"item_id": itemInvestConf.Id,
					"name":    itemInvestConf.Name,
					"icon":    itemInvestConf.Icon,
				}
			}
		} else {
			winItemData["first_item"] = map[string]interface{}{}
			winItemData["second_item"] = map[string]interface{}{}
			winItemData["third_item"] = map[string]interface{}{
				"item_id": gameResult.Id,
				"name":    gameResult.Name,
				"icon":    gameResult.Icon,
			}
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
			"record_id":stakeList[0].Periods,
			"win_item_data":winItemData,
			"detail":  detail,
		}
	}
}
