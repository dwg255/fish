package service

import (
	"github.com/astaxie/beego/logs"
	"game/invest/tools"
	"game/api/thrift/gen-go/rpc"
	"context"
)

func wsRequest(data map[string]interface{}, client *Client) {
	defer func() {
		if r:= recover(); r != nil {
			logs.Error("GetUserInfo panic:%v ",r)
		}
	}()
	if req, ok := data["act"]; ok {
		if act, ok := req.(string); ok {
			switch act {
			case "login_server_invest":
				logs.Debug("user request login")
				if token, ok := data["token"]; ok {
					if tokenStr, ok := token.(string); ok {
						rpcClient, closeTransport, err := tools.GetRpcClient()
						if err != nil {
							logs.Debug("get rpc client err:%v", err)
						}
						defer closeTransport()

						resp, err := rpcClient.GetUserInfoByken(context.Background(), tokenStr)
						if err != nil || resp.Code != rpc.ErrorCode_Success {
							logs.Debug("check user token [%v] failed err:%v, resp:%v", token, err, resp)
						} else {
							client.UserInfo.UserId = UserId(resp.UserObj.UserId)
							client.UserInfo.Nickname = resp.UserObj.NickName
							client.UserInfo.Icon = resp.UserObj.AvatarAuto
							client.UserInfo.Gold = int(resp.UserObj.Gold)
							client.Status = clientStatusGuest
							HubMgr.LoginChan <- client
						}
					}
				}
			case "praise_invest":
				if req, ok := data["num"]; ok {
					logs.Debug("receive num %v invalid", req)
					if numFloat, ok := req.(float64); ok {
						logs.Debug("receive str num %v invalid", req)
						num := int(numFloat)
						client.Room.Lock.Lock()
						defer client.Room.Lock.Unlock()
						for _, praiseItem := range client.Room.PraiseInfo {
							if praiseItem.UserId == client.UserInfo.UserId {
								praiseItem.Num = praiseItem.Num + num
								logs.Debug("room praise %v", client.Room.PraiseInfo)
								return
							}
						}
						client.Room.PraiseInfo = append(client.Room.PraiseInfo, &UserPraise{
							UserId: client.UserInfo.UserId,
							Num:    num,
						})
						logs.Debug("room praise %v", client.Room.PraiseInfo)
					}
				}
			}
		}
	}
}
