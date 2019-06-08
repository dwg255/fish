package controllers

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fish/common/api/thrift/gen-go/rpc"
	"fish/common/tools"
	"fish/game/common"
	"fish/game/service"
	"fmt"
	"github.com/astaxie/beego/logs"
	"net/http"
	"strconv"
	"time"
)

func EnterPublicRoom(w http.ResponseWriter, r *http.Request) {
	logs.Debug("new request EnterPublicRoom")

	defer func() {
		if r := recover(); r != nil {
			logs.Error("EnterPublicRoom panic:%v ", r)
		}
	}()
	account := r.FormValue("account")
	if len(account) == 0 {
		return
	}
	baseParam := r.FormValue("baseParam")
	if len(baseParam) == 0 {
		return
	}
	baseScore, err := strconv.Atoi(baseParam)
	if err != nil {
		logs.Debug("request enterPublicRoom err invalid baseParam %v", baseParam)
		return
	}
	sign := r.FormValue("sign")
	if len(sign) == 0 {
		return
	}
	token := r.FormValue("token")
	if len(token) == 0 {
		return
	}
	t := r.FormValue("t")
	if len(token) == 0 {
		return
	}
	ret := map[string]interface{}{
		"errcode": -1,
		"errmsg":  "enter room failed.",
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
	if client, closeTransportHandler, err := tools.GetRpcClient(common.GameConf.AccountHost, strconv.Itoa(common.GameConf.AccountPort)); err == nil {
		defer func() {
			if err := closeTransportHandler(); err != nil {
				logs.Error("close rpc err: %v", err)
			}
		}()
		if res, err := client.GetUserInfoByToken(context.Background(), sign); err == nil {
			if res.Code == rpc.ErrorCode_Success {
				userId := service.UserId(res.UserObj.UserId)
				if token == fmt.Sprintf("%x", md5.Sum([]byte("t"+t))) {
					// todo lock 🔒
					var roomId service.RoomId
					service.RoomMgr.RoomLock.Lock()
					logs.Info("EnterPublicRoom get lock...")
					defer service.RoomMgr.RoomLock.Unlock()
					defer logs.Info("EnterPublicRoom set free lock...")
					for RoomId, RoomInfo := range service.RoomMgr.RoomsInfo {
						for _, roomUserId := range RoomInfo.UserInfo {
							if userId == roomUserId {
								ret = map[string]interface{}{
									"errcode": 0,
									"errmsg":  "ok",
									"ip":      common.GameConf.GameHost,
									"port":    common.GameConf.GamePort,
									"roomId":  RoomId,
									"sign":    sign,
									"time":    time.Now().Unix() * 1000,
									"token":   token,
								}
								return
							}
						}
						if roomId == 0 && len(RoomInfo.UserInfo) < 4 && RoomInfo.BaseScore == baseScore {
							roomId = RoomId
						}
					}
					if roomId == 0 { //房间全满
						roomId = service.CreatePublicRoom(&service.RoomConf{
							BaseScore:    baseScore,
							MinHaveScore: service.MinHaveScore,
							MaxHaveScore: service.MaxHaveScore,
							TaxRatio:     service.TaxRatio,
							Creator:      userId,
						})
					}
					cannonKindVip := map[int]int{
						0: 1,
						1: 4,
						2: 7,
						3: 10,
						4: 13,
						5: 16,
						6: 19,
					}

					if roomInfo, ok := service.RoomMgr.RoomsInfo[roomId]; ok {
						resChan := make(chan error)
						roomInfo.HttpReqChan <- &service.HttpReqData{
							UserInfo: service.UserInfo{
								UserId:     userId,
								Score:      int(res.UserObj.Gems),
								Name:       res.UserObj.NickName,
								Ready:      false,
								SeatIndex:  0,
								Vip:        int(res.UserObj.Vip),
								CannonKind: cannonKindVip[int(res.UserObj.Vip)],
								Power:      float64(res.UserObj.Power) / 1000,
								LockFishId: 0,
							}, ErrChan: resChan,
						}
						timeOut := time.After(time.Second)
						select {
						case <-timeOut:
							return
						case err := <-resChan:
							if err != nil {
								logs.Error("EnterPublicRoom enter room [%d] err: %v", roomId, err)
							} else {
								exists := false
								for _, roomUserId := range service.RoomMgr.RoomsInfo[roomId].UserInfo {
									if roomUserId == userId {
										exists = true
									}
								}
								if !exists {
									roomInfo.UserInfo = append(service.RoomMgr.RoomsInfo[roomId].UserInfo, userId)
								}
								ret = map[string]interface{}{
									"errcode": 0,
									"errmsg":  "ok",
									"ip":      common.GameConf.GameHost,
									"port":    common.GameConf.GamePort,
									"roomId":  strconv.Itoa(int(roomId)),
									"sign":    sign,
									"time":    time.Now().Unix() * 1000,
									"token":   token,
									"mark":    1,
								}
								return
							}
						}
					}
				} else {
					logs.Error("EnterPublicRoom check token err")
				}
			}
		} else {
			logs.Error("get UserInfo by token err: %v", err)
		}
	}
}
