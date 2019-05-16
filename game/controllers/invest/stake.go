package invest

import (
	"net/http"
	"github.com/astaxie/beego/logs"
	"strconv"
	"game/invest/tools"
	"game/invest/service"
	"game/invest/common"
	"golang.org/x/net/context"
	"encoding/json"
	"game/api/thrift/gen-go/rpc"
	"time"
)

func Stake(w http.ResponseWriter, r *http.Request) {
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
	stakeItemIdStr := r.FormValue("stake_item_id")
	stakeItemId, err := strconv.Atoi(stakeItemIdStr)
	if err != nil {
		ret["code"] = common.ErrorParamInvalid
		logs.Error("request invest stake param stake_item_id [%s] err:", stakeItemIdStr, err)
		return
	}

	var positionCheck bool
	for _, investConf := range common.InvestConfArr {
		if investConf.Id == stakeItemId {
			positionCheck = true
			break
		}
	}
	if !positionCheck {
		ret["code"] = common.ErrorParamInvalid
		logs.Error("user [] stake in err position :%v", stakeItemId)
		return
	}

	stakeGoldenBeanStr := r.FormValue("stake_golden_bean")
	stakeGoldenBean, err := strconv.Atoi(stakeGoldenBeanStr)
	if err != nil || stakeGoldenBean <= 0 {
		ret["code"] = common.ErrorParamInvalid
		logs.Error("request invest stake param stakeGoldenBean [%s] err:", stakeGoldenBeanStr, err)
		return
	}

	token := r.FormValue("token")
	rpcClient, closeTransport, err := tools.GetRpcClient()
	if err != nil {
		logs.Debug("get rpc client err:%v", err)
		ret["code"] = common.ErrorRpcServerError
	}
	defer closeTransport()

	resp, err := rpcClient.GetUserInfoByken(context.Background(), token)
	if err != nil || resp.Code != rpc.ErrorCode_Success {
		ret["code"] = common.ErrorAuthFailed
		logs.Debug("check user token [%v] failed", token)
	} else {
		userId := resp.UserObj.UserId
		c, ok := service.HubMgr.UserToRoom[service.UserId(userId)]
		if !ok {
			ret["code"] = common.ErrorUnknownError
			logs.Error("user [%d] request invest stake ,but not in room ", userId)
			return
		}
		var canStakeGoldList [3]int
		for _, v := range common.CanStakeChipConf {
			if v.GoldMax == -1 && resp.UserObj.Gold >= v.GoldMin {
				canStakeGoldList = v.Chips
			} else if resp.UserObj.Gold >= v.GoldMin && resp.UserObj.Gold <= v.GoldMax {
				canStakeGoldList = v.Chips
			}
		}
		if len(canStakeGoldList) != 3 {
			ret["code"] = common.ErrorUnknownError
			return
		}
		var stakeGoldCheck bool
		for _, stakeGold := range canStakeGoldList {
			if stakeGold == stakeGoldenBean {
				stakeGoldCheck = true
			}
		}
		if !stakeGoldCheck {
			//不开启验证
			//ret["code"] = common.ErrorParamInvalid
			//return
		}
		if client, ok := c.RoomClients[service.UserId(userId)]; ok {
			if service.HubMgr.GameManage.GameStatus == common.StatusStartStake || service.HubMgr.GameManage.GameStatus == common.StatusSendStake {
				resp, err := rpcClient.ModifyGoldById(context.Background(), "invest_stake", int32(userId), -1*int64(stakeGoldenBean))
				if err != nil || resp.Code != rpc.ErrorCode_Success {
					ret["code"] = common.ErrorGoldNotEnough
					return
				} else {
					successCh := make(chan bool, 1)
					defer close(successCh)
					c.UserStakeChan <- &service.StakeInfo{
						Client:    client,
						Position:  stakeItemId,
						StakeGold: stakeGoldenBean,
						SuccessCh: successCh,
					}
					select {
					case stakeSuccess := <-successCh:
						if stakeSuccess {
							ret["code"] = common.CodeSuccess
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
							ret["msg"] = "押注成功"
						} else {
							ret["code"] = common.ErrorNotStakeTime
							rpcClient.ModifyGoldById(context.Background(), "invest_stake", int32(userId), int64(stakeGoldenBean))
						}
					}
				}
			}
		}
	}
}
