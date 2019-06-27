package service

import (
	"context"
	"encoding/json"
	"fish/common/api/thrift/gen-go/rpc"
	"fish/common/tools"
	"fish/game/common"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strconv"
	"strings"
	"time"
)

type UserLockFishReq struct {
	UserId  UserId `json:"userId"`
	ChairId int    `json:"chairId"`
	FishId  FishId `json:"fishId"`
}

type LaserCatchReq struct {
	UserId  UserId `json:"userId"`
	ChairId int    `json:"chairId"`
	Fishes  string `json:"fishes"`
	Sign    string `json:"sign"`
}

type UserFireLaserReq struct {
	UserId     UserId  `json:"userId"`
	ChairId    int     `json:"chairId"`
	BulletKind int     `json:"bulletKind"`
	BulletId   int     `json:"bulletId"`
	Angle      float64 `json:"angle"`
	Sign       string  `json:"sign"`
	LockFishId FishId  `json:"lockFishId"`
}

func wsRequest(req []byte, client *Client) {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("wsRequest panic:%v ", r)
		}
	}()
	if req[0] == '4' && req[1] == '2' {
		reqJson := make([]string, 0)
		err := json.Unmarshal(req[2:], &reqJson)
		if err != nil {
			logs.Error("wsRequest json unmarshal err :%v", err)
			return
		}
		if client.Room == nil { //未登录
			logs.Info("未登录 login msg : %v", reqJson[0])
			if reqJson[0] == "login" {
				if len(reqJson) < 2 {
					return
				}
				//if reqByteData, ok := reqJson[1].([]byte); ok {
				reqData := make(map[string]string)
				if err := json.Unmarshal([]byte(reqJson[1]), &reqData); err != nil {
					roomIdStr := reqData["roomId"]
					if roomIdStr == "" { //客户端重连时roomId用的int类型。。。心累
						reqDataReconnect := make(map[string]int)
						if err := json.Unmarshal([]byte(reqJson[1]), &reqDataReconnect); err != nil {
							roomIdInt := reqDataReconnect["roomId"]
							roomIdStr = strconv.Itoa(roomIdInt)
						}
					}
					if roomIdInt, err := strconv.Atoi(roomIdStr); err == nil {
						roomId := RoomId(roomIdInt)
						RoomMgr.RoomLock.Lock()
						logs.Info("login get lock...")
						defer RoomMgr.RoomLock.Unlock()
						defer logs.Info("login set free lock...")
						if room, ok := RoomMgr.Rooms[roomId]; ok {
							//if room.Status == GameStatusWaitBegin {
							//	room.Status = GameStatusFree
							//	room.Utils.BuildFishTrace()
							//}
							logs.Debug("send succ")
							client.Room = room
							room.ClientReqChan <- &clientReqData{
								client,
								reqJson,
							}
						} else {
							logs.Error("room %v, not exists", roomId)
						}
					} else {
						logs.Error("roomId %v err : %v", roomIdStr, err)
					}
				}
				//}
			} else {
				logs.Error("invalid act %v", reqJson[0])
			}
		} else {
			//logs.Debug("send req to room [%d] succ 2", client.Room.RoomId)
			client.Room.ClientReqChan <- &clientReqData{
				client,
				reqJson,
			}
		}
	} else {
		logs.Error("invalid message %v", req)
	}
}

//todo 弱类型语言写的东西重构简直堪比火葬场
func handleUserRequest(clientReq *clientReqData) {
	reqJson := clientReq.reqData
	client := clientReq.client
	if len(reqJson) > 0 {
		act := reqJson[0]
		switch act {
		case "login":
			//logs.Debug("login")
			if len(reqJson) < 2 {
				return
			}
			reqData := make(map[string]interface{})
			if err := json.Unmarshal([]byte(reqJson[1]), &reqData); err == nil {
				token := reqData["sign"]
				if token, ok := token.(string); ok {
					logs.Debug("token %v", token)
					if rpcClient, closeTransportHandler, err := tools.GetRpcClient(common.GameConf.AccountHost, strconv.Itoa(common.GameConf.AccountPort)); err == nil {
						defer func() {
							if err := closeTransportHandler(); err != nil {
								logs.Error("close rpc err: %v", err)
							}
						}()
						if res, err := rpcClient.GetUserInfoByToken(context.Background(), token); err == nil {
							//logs.Debug("rpc res : %v", res.Code)
							if res.Code == rpc.ErrorCode_Success {
								userId := UserId(res.UserObj.UserId)
								for _, userInfo := range client.Room.Users {
									if userId == userInfo.UserId {
										userInfo.client = client
										userInfo.Online = true
										userInfo.Ip = "::1"
										client.UserInfo = userInfo
										logs.Debug("client userInfo get data...")
										seats := make([]interface{}, 0)
										cannonKindVip := map[int]int{0: 1, 1: 4, 2: 7, 3: 10, 4: 13, 5: 16, 6: 19}
										//todo check sign
										//score, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", float64(userInfo.Score)/1000), 64)
										userInfo.ConversionScore, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", float64(userInfo.Score)/1000), 64)
										for _, userInfo := range client.Room.Users {
											seats = append(seats, map[string]interface{}{
												"userId":    userInfo.UserId,
												"ip":        "",
												"score":     userInfo.ConversionScore,
												"name":      userInfo.Name,
												"vip":       userInfo.Vip,
												"online":    true,
												"ready":     userInfo.Ready,
												"seatIndex": userInfo.SeatIndex,

												// 正在使用哪种炮 todo 换为真实vip
												"cannonKind": cannonKindVip[0],
												// 能量值
												"power": 0,
											})
										}
										client.sendToClient([]interface{}{
											"login_result",
											map[string]interface{}{
												"errcode": 0,
												"errmsg":  "ok",
												"data": map[string]interface{}{
													"roomId":     strconv.Itoa(int(client.Room.RoomId)),
													"conf":       client.Room.Conf,
													"numofgames": 0,
													"seats":      seats,
												},
											},
										})
										client.sendToOthers([]interface{}{
											"new_user_comes_push",
											client.UserInfo,
										})
										client.sendToClient([]interface{}{
											"login_finished",
										})
										return
									}
								}
								//不用断开链接，客户端的问题导致需要保持很多无用链接。。。
								logs.Debug("user need enter room")
								client.closeChan <- true
								close(client.closeChan)
								return
							} else {
								logs.Error("account server rpc status: %v, err : %v", res.Code, err)
							}
						} else {
							logs.Debug("rpc GetUserInfoByToken err : %v", err)
						}
					} else {
						logs.Debug("get rpc [%v:%v] client err : %v", common.GameConf.AccountHost, common.GameConf.AccountPort, err)
					}
				}
			} else {
				logs.Error("json unmarshal err : %v", err)
			}
			client.Room = nil
		case "catch_fish":
			if len(reqJson) < 2 {
				return
			}
			//42["catch_fish","{\"userId\":101,\"chairId\":1,\"bulletId\":\"1_324965\",\"fishId\":\"10318923\",\"sign\":\"8bfef2b82dc7b97e4ad386ec40b83d2b\"}"]
			catchFishReq := catchFishReq{}
			if err := json.Unmarshal([]byte(reqJson[1]), &catchFishReq); err == nil {
				bulletId := catchFishReq.BulletId
				client.catchFish(catchFishReq.FishId, bulletId)
			} else {
				logs.Error("catch_fish req err: %v", err)
			}
		case "ready":
			if len(reqJson) < 2 {
				return
			}
			reqData := make(map[string]int)
			if err := json.Unmarshal([]byte(reqJson[1]), &reqData); err == nil {
				userId := UserId(reqData["userId"])
				client.Room.Users[userId].Ready = true
				if client.Room.Status == GameStatusWaitBegin {
					client.Room.Status = GameStatusFree
					//client.Room.begin()
					client.Room.Utils.BuildFishTrace()
				}
				client.UserInfo.Online = true
				roomUsers := make([]*UserInfo, 0)
				for i := 0; i < 4; i++ {
					seatHasPlayer := false
					for _, userInfo := range client.Room.Users {
						if userInfo.SeatIndex == i {
							userInfo.ConversionScore, err = strconv.ParseFloat(fmt.Sprintf("%.3f", float64(userInfo.Score)/1000), 64)
							if err != nil {
								logs.Error("ParseFloat [%v] err %v", userInfo.Score, err)
							}
							roomUsers = append(roomUsers, userInfo)
							seatHasPlayer = true
						}
					}
					if !seatHasPlayer {
						roomUsers = append(roomUsers, &UserInfo{
							SeatIndex: i,
						})
					}
				}
				client.sendToClient([]interface{}{
					"game_sync_push",
					map[string]interface{}{
						"roomBaseScore": client.Room.Conf.BaseScore,
						"seats":         roomUsers,
					},
				})
			} else {
				logs.Error("user req ready json unmarshal err : %v", err)
			}
		case "user_fire":
			if len(reqJson) < 2 {
				return
			}
			bullet := Bullet{}
			if err := json.Unmarshal([]byte(reqJson[1]), &bullet); err == nil {
				client.Fire(&bullet)
			} else {
				// todo 没办法 客户端bulletId 传的int :(
				userFireLaserReq := &UserFireLaserReq{}
				if err := json.Unmarshal([]byte(reqJson[1]), &userFireLaserReq); err == nil {
					bullet.UserId = userFireLaserReq.UserId
					bullet.ChairId = userFireLaserReq.ChairId
					bullet.BulletKind = userFireLaserReq.BulletKind
					bullet.BulletId = ""
					bullet.Angle = userFireLaserReq.Angle
					bullet.Sign = userFireLaserReq.Sign
					bullet.LockFishId = userFireLaserReq.LockFishId
					client.UserInfo.Power = 0
					client.sendToOthers([]interface{}{
						"user_fire_Reply",
						bullet,
					})
					return
				}
				logs.Error("user fire json err: %v", err)
			}
		case "laser_catch_fish":
			if len(reqJson) < 2 {
				return
			}
			laserCatchReq := LaserCatchReq{}
			if err := json.Unmarshal([]byte(reqJson[1]), &laserCatchReq); err == nil {
				fishIdStrArr := strings.Split(laserCatchReq.Fishes, "-")
				if len(fishIdStrArr) == 0 {
					logs.Debug("user [%v] laser_catch_fish catch zero fish...")
				}
				killedFishes := make([]string, 0)
				addScore := 0
				for _, fishStr := range fishIdStrArr {
					if fishIdInt, err := strconv.Atoi(fishStr); err == nil {
						fishId := FishId(fishIdInt)
						if fish, ok := client.Room.AliveFish[fishId]; ok {
							killedFishes = append(killedFishes, strconv.Itoa(int(fish.FishId)))
							//加钱
							addScore += GetFishMulti(fish) * GetBulletMulti(BulletKind["bullet_kind_laser"]) * client.Room.Conf.BaseScore
						} else {
							logs.Debug("user [%v] laser_catch_fish fishId [%v] not in alive fish array...", client.UserInfo.UserId, fishId)
						}
					} else {
						logs.Error("laser_catch_fish err : fishId [%v] err", fishStr)
					}
				}
				//if addScore > client.Room.Conf.BaseScore*200 { //最大200倍
				//	addScore = client.Room.Conf.BaseScore * 200
				//}
				client.UserInfo.Score += addScore
				client.UserInfo.Bill += addScore //记账
				catchFishAddScore, _ := strconv.ParseFloat(fmt.Sprintf("%.5f", float64(addScore)/1000), 64)
				client.Room.broadcast([]interface{}{
					"catch_fish_reply",
					map[string]interface{}{
						"userId":   laserCatchReq.UserId,
						"chairId":  laserCatchReq.ChairId,
						"fishId":   strings.Join(killedFishes, ","),
						"addScore": catchFishAddScore,
						"isLaser":  true,
					},
				})
			} else {
				logs.Error("laser_catch_fish err : %v", err)
			}
		case "user_lock_fish":
			if len(reqJson) < 2 {
				return
			}
			userLockFishReq := UserLockFishReq{}
			if err := json.Unmarshal([]byte(reqJson[1]), &userLockFishReq); err == nil {
				client.sendToOthers([]interface{}{
					"lock_fish_reply",
					userLockFishReq,
				})
			}
		case "user_frozen":
			if len(reqJson) < 2 {
				return
			}
			client.frozenScene(time.Now())
		case "user_change_cannon":
			if len(reqJson) < 2 {
				return
			}
			userChangeCannonReq := make(map[string]int)
			if err := json.Unmarshal([]byte(reqJson[1]), &userChangeCannonReq); err == nil {
				if userChangeCannonReq["cannonKind"] < 1 {
					return
				}
				if userChangeCannonReq["cannonKind"] == BulletKind["bullet_kind_laser"] {
					if client.UserInfo.Power < 1 {
						return
					}
				}
				client.UserInfo.CannonKind = userChangeCannonReq["cannonKind"]
				client.sendToOthers([]interface{}{
					"user_change_cannon_reply",
					userChangeCannonReq,
				})
			}
		case "exit":
			client.sendToOthers([]interface{}{
				"exit_notify_push",
				client.UserInfo.UserId,
			})
			jsonByte, err := json.Marshal([]string{"exit_result"})
			if err != nil {
				logs.Error("game ping json marshal err,%v", err)
				return
			}
			client.sendMsg(append([]byte{'4', '2'}, jsonByte...))
			client.sendMsg([]byte{'4', '1'})
			clientExit(client, false)

		case "dispress":
		case "disconnect":
		case "game_ping":
			jsonByte, err := json.Marshal([]string{"game_pong"})
			if err != nil {
				logs.Error("game ping json marshal err,%v", err)
				return
			}
			client.sendMsg(append([]byte{'4', '2'}, jsonByte...))
		case "client_exit":
			if client.UserInfo.Online {
				clientExit(client, true)
			}
		}
	}
}

func clientExit(client *Client, closeClient bool) {
	logs.Debug("user %v exit close client: %v ...", client.UserInfo.UserId, closeClient)
	if client.UserInfo.Bill != 0 {
		client.clearBill()
	}
	RoomMgr.RoomLock.Lock()
	logs.Info("clientExit get lock...")
	defer RoomMgr.RoomLock.Unlock()
	defer logs.Info("clientExit set free lock...")
	client.UserInfo.Online = false
	roomUserIdArr := make([]UserId, 0)
	if roomInfo, ok := RoomMgr.RoomsInfo[client.Room.RoomId]; ok {
		for _, roomUserId := range roomInfo.UserInfo {
			if roomUserId != client.UserInfo.UserId {
				roomUserIdArr = append(roomUserIdArr, roomUserId)
			}
		}
		roomInfo.UserInfo = roomUserIdArr
		delete(client.Room.Users, client.UserInfo.UserId)
		if closeClient {
			client.closeChan <- true
			close(client.closeChan) //关闭channel不影响取出关闭前传送的数据，继续取将得到零值  :-)
		}
		if len(client.Room.Users) == 0 { //房间无人，消除房间
			delete(RoomMgr.RoomsInfo, client.Room.RoomId)
			delete(RoomMgr.Rooms, client.Room.RoomId)
			logs.Debug("room %v is empty now ...", client.Room.RoomId)
			client.Room.Exit <- true
			logs.Debug("send exit sign succ ...")
		}
		//close(client.msgChan)
	} else {
		logs.Error("exit client not in room...")
	}
}
