package invest

import (
	"net/http"
	"github.com/astaxie/beego/logs"
	"encoding/json"
	"time"
	"game/invest/common"
	"game/api/thrift/gen-go/rpc"
	"game/invest/tools"
	"golang.org/x/net/context"
)

func RewardRecord(w http.ResponseWriter, r *http.Request) {
	defer func() {
		//if r := recover(); r != nil {
		//	logs.Error("RewardRecord panic:%v ", r)
		//}
	}()
	//logs.Debug("new request url:[%s]",r.URL)
	ret := make(map[string]interface{})
	data := make([]map[string]interface{}, 0)

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
		result := map[string]interface{}{
			"identity_obj": identityObj,
		}
		reqType := r.FormValue("type")

		common.GameRecord.Lock.RLock()
		defer common.GameRecord.Lock.RUnlock()
		switch reqType {
		case "1":
			//logs.Debug("%v",common.GameRecord.LasTTenTimesRank)
			//return
			for i := len(common.GameRecord.LasTTenTimesRank) - 1; i >= 0; i-- {
				rank := common.GameRecord.LasTTenTimesRank[i]
				item := make(map[string]interface{})
				if rank != nil {
					if len(rank.GameResult.ParentArr) > 1 {
						//填坑， :-(
						for k, itemInvestConf := range rank.GameResult.ParentArr {
							key := "third_item"
							if k == 0 {
								key = "third_item"
							} else if k == 1 {
								key = "second_item"
							} else if k == 2 {
								key = "first_item"
							}
							item[key] = map[string]interface{}{
								"item_id": itemInvestConf.Id,
								"name":    itemInvestConf.Name,
								"icon":    itemInvestConf.Icon,
							}
						}
					} else {
						item["first_item"] = map[string]interface{}{}
						item["second_item"] = map[string]interface{}{}
						item["third_item"] = map[string]interface{}{
							"item_id": rank.GameResult.Id,
							"name":    rank.GameResult.Name,
							"icon":    rank.GameResult.Icon,
						}
					}
					data = append(data, item)
				}
			}
		case "2":
			for _, investConf := range common.InvestConfArr {
				item := make(map[string]interface{})
				isSet := false
				for investConfId, count := range common.GameRecord.LastTimeWinRecord {
					if investConf.Id == investConfId {
						item = map[string]interface{}{
							"item_id": investConfId,
							"name":    investConf.Name,
							"icon":    investConf.Icon,
							"count":   count,
						}
						isSet = true
						break
					}
				}
				if !isSet {
					item = map[string]interface{}{
						"item_id": investConf.Id,
						"name":    investConf.Name,
						"icon":    investConf.Icon,
						"count":   -1,
					}
				}
				data = append(data, item)
			}
		case "3":
			for _,investConf := range common.InvestConfArr{
				var item  map[string]interface{}
				if _,ok := common.GameRecord.EverydayResultRecord.Record[investConf];ok {
					item = map[string]interface{}{
						"item_id" : investConf.Id,
						"name" : investConf.Name,
						"icon" : investConf.Icon,
						"count" : common.GameRecord.EverydayResultRecord.Record[investConf],
					}
				} else {
					item = map[string]interface{}{
						"item_id" : investConf.Id,
						"name" : investConf.Name,
						"icon" : investConf.Icon,
						"count" : 0,
					}
				}
				data = append(data,item)
			}
		case "4":
			for _,investConf := range common.InvestConfArr{
				var item  map[string]interface{}
				if _,ok := common.GameRecord.ContinuedWinRecord[investConf];ok {
					item = map[string]interface{}{
						"item_id" : investConf.Id,
						"name" : investConf.Name,
						"icon" : investConf.Icon,
						"count" : common.GameRecord.ContinuedWinRecord[investConf],
					}
				} else {
					item = map[string]interface{}{
						"item_id" : investConf.Id,
						"name" : investConf.Name,
						"icon" : investConf.Icon,
						"count" : 0,
					}
				}
				data = append(data,item)
			}
		default:
			ret["code"] = common.ErrorParamInvalid
			return
		}
		result["data"] = data
		ret["result"] = result
		ret["msg"] = "请求成功"
	}
}
