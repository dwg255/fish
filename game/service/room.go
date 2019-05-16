package service

import (
	"sync"
	"time"
	"github.com/gorilla/websocket"
	"fmt"
	"game/invest/common"
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"game/invest/tools"
	"golang.org/x/net/context"
	"game/api/thrift/gen-go/rpc"
)

type UserPraise struct {
	UserId UserId `json:"user_id"`
	Num    int    `json:"num"`
}

type StakeInfo struct {
	Client    *Client
	Position  int
	StakeGold int
	SuccessCh chan bool
}

type Room struct {
	RoomId               RoomId
	RoomClients          map[UserId]*Client
	TeamRadio            chan *common.Message
	StakeInfoMap         map[*Client]map[int]int
	WaitSendStakeInfoMap map[*Client]map[int]int
	UserStakeChan        chan *StakeInfo
	LoginOutMap          map[UserId]*Client
	MemberCount          int

	LoginOutChan chan *Client

	PraiseInfo []*UserPraise
	Lock       sync.RWMutex
}

func (c *Client) NewRoom() (room *Room) {
	c.Hub.RoomIdInr = c.Hub.RoomIdInr + 1
	roomClients := map[UserId]*Client{c.UserInfo.UserId: c}
	room = &Room{
		RoomId:               c.Hub.RoomIdInr,
		RoomClients:          roomClients,
		MemberCount:          1,
		TeamRadio:            make(chan *common.Message, 100),
		PraiseInfo:           make([]*UserPraise, 0, 5),
		StakeInfoMap:         make(map[*Client]map[int]int),
		WaitSendStakeInfoMap: make(map[*Client]map[int]int),
		UserStakeChan:        make(chan *StakeInfo, 100),
		LoginOutChan:         make(chan *Client, 5),
		LoginOutMap:          make(map[UserId]*Client, 5),
	}
	c.Room = room
	c.Hub.Rooms[room] = 1
	c.Hub.UserToRoom[c.UserInfo.UserId] = room
	c.Status = clientStatusLogin
	go room.roomServer()
	logs.Debug("create new room ok!")
	return
}

func (p *Room) IntoRoom(c *Client) (succ bool, err error) {
	p.Lock.Lock()
	if len(p.RoomClients) > 4 {
		err = fmt.Errorf("room already full")
		return
	}
	if _, ok := p.RoomClients[c.UserInfo.UserId]; ok {
		err = fmt.Errorf("user already in room")
		return
	}
	p.RoomClients[c.UserInfo.UserId] = c
	p.MemberCount++ //人数 +1
	p.Lock.Unlock()

	c.Hub.Lock.Lock()
	c.Hub.UserToRoom[c.UserInfo.UserId] = p
	c.Hub.Rooms[p]++ //人数 +1
	c.Hub.Lock.Unlock()

	c.Room = p
	c.Status = clientStatusLogin
	succ = true
	return
}

func (p *Room) roomServer() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		logs.Debug("room [%d] logout ", p.RoomId)
		ticker.Stop()
		close(p.TeamRadio)
		close(p.UserStakeChan)
		close(p.LoginOutChan)

		HubMgr.Lock.Lock()
		if _, ok := HubMgr.Rooms[p]; ok {
			delete(HubMgr.Rooms, p)
		} else {
			logs.Error("empty room [%d] logout from hub failed", p.RoomId)
		}
		HubMgr.Lock.Unlock()
	}()
	for {
		select {
		case message, ok := <-p.TeamRadio:
			if !ok {
				for _, Client := range p.RoomClients {
					Client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				}
				logs.Debug("room [%d] team radio close", p.RoomId)
				return
			}
			data := &Message{
				NewStake: make(map[string]map[string]int),
			}
			//logs.Debug("game status %v",message.Type)
			switch message.Type {
			case common.StatusPrepare:
				//处理离线
				p.Lock.Lock()
				roomUserChange := false
				for userId := range p.LoginOutMap {
					roomUserChange = true
					HubMgr.Lock.Lock()
					if _, ok := HubMgr.UserToRoom[userId]; ok {
						delete(HubMgr.UserToRoom, userId)
					}
					if _, ok := HubMgr.LoginOutMap[userId]; ok {
						delete(HubMgr.LoginOutMap, userId)
					}
					HubMgr.Rooms[p]--
					HubMgr.Lock.Unlock()
					p.MemberCount--
					if _, ok := p.RoomClients[userId]; ok { //冗余操作
						delete(p.RoomClients, userId)
					}
				}
				p.Lock.Unlock()

				p.Lock.RLock()
				if p.MemberCount == 0 {
					logs.Debug("room [%d] member count = 0 ,room exit", p.RoomId)
					p.Lock.RUnlock()
					return
				}
				if roomUserChange {
					//通知其他人
					userInRoom := make([]map[string]interface{}, 0)
					for _, roomClient := range p.RoomClients {
						userInRoom = append(userInRoom, map[string]interface{}{
							"user_id":  roomClient.UserInfo.UserId,
							"nickname": roomClient.UserInfo.Nickname,
							"icon":     roomClient.UserInfo.Icon,
							"gold":     roomClient.UserInfo.Gold,
						})
					}
					var noticeUserInRoom = make(map[string]interface{})
					noticeUserInRoom["act"] = "room_user_info"
					noticeUserInRoom["user_in_room"] = userInRoom
					p.sendToRoomMembers(noticeUserInRoom)
				}
				p.Lock.RUnlock()

				data.Act = "change_state"
				data.State = "init"
				//初始化数据
				p.Lock.Lock()
				p.PraiseInfo = make([]*UserPraise, 0, 5)
				p.LoginOutMap = make(map[UserId]*Client)
				p.StakeInfoMap = make(map[*Client]map[int]int)
				p.WaitSendStakeInfoMap = make(map[*Client]map[int]int)
				p.Lock.Unlock()

			case common.StatusStartStake:
				data.Act = "change_state"
				data.State = "staking"
				data.Timer = 27
				HubMgr.GameManage.Lock.RLock()
				data.Periods = HubMgr.GameManage.Periods
				HubMgr.GameManage.Lock.RUnlock()
			case common.StatusSendStake:
				//发送加油
				p.Lock.RLock()
				if len(p.PraiseInfo) != 0 {
					praiseMsg := map[string]interface{}{
						"act":         "praise_info",
						"praise_info": p.PraiseInfo,
						"timer":       HubMgr.GameManage.StakeCountdown,
					}
					p.sendToRoomMembers(praiseMsg)
					p.PraiseInfo = make([]*UserPraise, 0, 5)
				}
				p.Lock.RUnlock()

				//logs.Debug("WaitSendStakeInfoMap:%v",p.WaitSendStakeInfoMap)
				//logs.Debug("StakeInfoMap:%v",p.StakeInfoMap)

				data.Act = "stake_info"
				HubMgr.GameManage.Lock.RLock()
				data.Timer = HubMgr.GameManage.StakeCountdown
				HubMgr.GameManage.Lock.RUnlock()
				HubMgr.GameInfo.Lock.RLock()
				data.Pool = HubMgr.GameInfo.Pool
				HubMgr.GameInfo.Lock.RUnlock()

				p.Lock.RLock()
				for client, stakeMap := range p.WaitSendStakeInfoMap {
					key := fmt.Sprintf("user_%d", client.UserInfo.UserId)
					var userNewStake = make(map[string]int)
					for position, gold := range stakeMap {
						field := fmt.Sprintf("position_%d", position)
						userNewStake[field] = gold
					}
					data.NewStake[key] = userNewStake
				}
				stake := make(map[int]int)
				for _, stakeMap := range p.StakeInfoMap {
					for position, gold := range stakeMap {
						if stakeGold, ok := stake[position]; ok {
							stake[position] = stakeGold + gold
						} else {
							stake[position] = gold
						}
					}
				}
				p.Lock.RUnlock()

				for position, gold := range stake {
					data.StakeInfo = append(data.StakeInfo, map[string]int{"position": position, "stake_gold": gold})
				}
				//清空代发消息
				p.Lock.Lock()
				p.WaitSendStakeInfoMap = make(map[*Client]map[int]int)
				p.Lock.Unlock()

				if len(data.NewStake) == 0 {
					continue
				}

			case common.StatusEndStake:
				data.Act = "change_state"
				data.State = "endstake"
			case common.StatusSendResult:
				//这儿是之前留的坑，两个不同的字段用了同一个name,故改用map
				HubMgr.GameInfo.Lock.RLock()
				gameResult := map[string]interface{}{
					"act":          "push_result",
					"win_result":   HubMgr.GameInfo.GameResult.Id,
					"win_position": HubMgr.GameInfo.GameResult.WinPosition,
				}
				HubMgr.GameInfo.Lock.RUnlock()
				p.Lock.RLock()
				p.sendToRoomMembers(gameResult)
				p.Lock.RUnlock()
				//continue
			case common.StatusShowResult:
				data.Act = "change_state"
				data.State = "show_result"
			case common.StatusShowWinGold:
				data.Act = "change_state"
				data.State = "show_win_result"
			}
			if message.Type != common.StatusSendResult {
				p.Lock.RLock()
				p.sendToRoomMembers(data)
				p.Lock.RUnlock()
			} else { //结算
				data = &Message{
					Act: "send_room_win_result",
				}
				p.Lock.RLock()
				HubMgr.GameInfo.Lock.RLock()

				for userId, client := range p.RoomClients {
					userWInResult := &UserWinResult{
						UserId: userId,
					}
					for _, positionId := range HubMgr.GameInfo.GameResult.WinPosition {
						if stakeGold, ok := p.StakeInfoMap[client][positionId]; ok {
							if rate, ok := common.RateMap[positionId]; ok {
								userWInResult.WinGold = userWInResult.WinGold + int(float32(stakeGold)*rate)
							} else {
								logs.Error("read position [%d] rate failed ", positionId)
							}
						}
					}
					if userWInResult.WinGold > 0 {
						data.UserWinResult = append(data.UserWinResult, userWInResult)
					}
				}
				//处理离线用户盈利
				for userId, client := range p.LoginOutMap {
					userWInResult := &UserWinResult{
						UserId: userId,
					}
					for _, positionId := range HubMgr.GameInfo.GameResult.WinPosition {
						if stakeGold, ok := p.StakeInfoMap[client][positionId]; ok {
							if rate, ok := common.RateMap[positionId]; ok {
								userWInResult.WinGold = userWInResult.WinGold + int(float32(stakeGold)*rate)
							} else {
								logs.Error("read position [%d] rate failed ", positionId)
							}
						}
					}
					if userWInResult.WinGold > 0 {
						data.UserWinResult = append(data.UserWinResult, userWInResult)
					}
				}
				p.Lock.RUnlock()
				HubMgr.GameInfo.Lock.RUnlock()

				if len(data.UserWinResult) > 0 {
					p.Lock.RLock()
					p.sendToRoomMembers(data)
					p.Lock.RUnlock()
					go p.addWinGold(data.UserWinResult)
				}
				//log.Debug("send_room_win_result:%v", data.UserWinResult)
			}
		case <-ticker.C: //心跳
			for _, Client := range p.RoomClients {
				Client.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := Client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					Client.conn.Close()
					p.LoginOutChan <- Client
				}
			}
		case cli := <-p.LoginOutChan:
			p.Lock.Lock()
			if _, ok := p.LoginOutMap[cli.UserInfo.UserId]; ok {
				p.Lock.Unlock()
				continue
			}
			p.LoginOutMap[cli.UserInfo.UserId] = cli
			if _, ok := p.RoomClients[cli.UserInfo.UserId]; ok {
				delete(p.RoomClients, cli.UserInfo.UserId)
			}
			p.Lock.Unlock()

			HubMgr.Lock.Lock()
			HubMgr.LoginOutMap[cli.UserInfo.UserId] = cli
			delete(HubMgr.UserToRoom, cli.UserInfo.UserId)
			HubMgr.Lock.Unlock()

		case stakeInfo := <-p.UserStakeChan:
			//logs.Debug("user [%d] stake gold : %d,position:%d", stakeInfo.Client.UserInfo.UserId, stakeInfo.StakeGold, stakeInfo.Position)
			HubMgr.GameManage.Lock.RLock()
			if HubMgr.GameManage.GameStatus == common.StatusStartStake || HubMgr.GameManage.GameStatus == common.StatusSendStake {
				if stakeInfoMap, ok := p.StakeInfoMap[stakeInfo.Client]; ok {
					if positionStakeMap, ok := stakeInfoMap[stakeInfo.Position]; ok {
						p.StakeInfoMap[stakeInfo.Client][stakeInfo.Position] = positionStakeMap + stakeInfo.StakeGold
					} else {
						p.StakeInfoMap[stakeInfo.Client][stakeInfo.Position] = stakeInfo.StakeGold
					}
				} else {
					p.StakeInfoMap[stakeInfo.Client] = map[int]int{stakeInfo.Position: stakeInfo.StakeGold}
				}

				if stakeInfoMap, ok := p.WaitSendStakeInfoMap[stakeInfo.Client]; ok {
					if positionStakeMap, ok := stakeInfoMap[stakeInfo.Position]; ok {
						p.WaitSendStakeInfoMap[stakeInfo.Client][stakeInfo.Position] = positionStakeMap + stakeInfo.StakeGold
					} else {
						p.WaitSendStakeInfoMap[stakeInfo.Client][stakeInfo.Position] = stakeInfo.StakeGold
					}
				} else {
					p.WaitSendStakeInfoMap[stakeInfo.Client] = map[int]int{stakeInfo.Position: stakeInfo.StakeGold}
				}
				HubMgr.GameInfo.Lock.Lock()
				if totalStake, ok := HubMgr.GameInfo.StakeMap[stakeInfo.Position]; ok {
					HubMgr.GameInfo.StakeMap[stakeInfo.Position] = totalStake + stakeInfo.StakeGold
				} else {
					HubMgr.GameInfo.StakeMap[stakeInfo.Position] = stakeInfo.StakeGold
				}
				HubMgr.GameInfo.Pool = HubMgr.GameInfo.Pool + stakeInfo.StakeGold*(100-GameConf.PumpingRate)/100
				HubMgr.GameInfo.Lock.Unlock()
				HubMgr.GameManage.Lock.RUnlock()
				stakeInfo.SuccessCh <- true

			} else {
				stakeInfo.SuccessCh <- false
				HubMgr.GameManage.Lock.RUnlock()
			}
		}
	}
}

func (p *Room) sendToRoomMembers(data interface{}) {
	//logs.Debug("room send to members message %v", data)
	msg, err := json.Marshal(data)
	if err != nil {
		logs.Error("send room win result data marsha1 err:%v", err)
	}
	for _, Client := range p.RoomClients {
		Client.sendMsg(msg)
	}
}

func (p *Room) addWinGold(userWinResult []*UserWinResult) {
	HubMgr.GameInfo.Lock.RLock()
	periods := HubMgr.GameInfo.Periods
	pool := HubMgr.GameInfo.Pool
	gameResult := HubMgr.GameInfo.GameResult.Id
	HubMgr.GameInfo.Lock.RUnlock()
	HubMgr.GameManage.Lock.RLock()
	gameTimesId := HubMgr.GameManage.TimesId
	HubMgr.GameManage.Lock.RUnlock()
	stakeTime := time.Now().Format("2006-01-02 15:04:05")
	client, closeTransportHandler, err := tools.GetRpcClient()
	defer closeTransportHandler()
	if err != nil {
		logs.Error("get rpc client err:%v,add user win gold failed :%v", err, userWinResult)
		return
	}

	for _, winResult := range userWinResult {
		resp, err := client.ModifyGoldById(context.Background(), "invest win gold", int32(winResult.UserId), int64(winResult.WinGold))
		if err != nil || resp.Code != rpc.ErrorCode_Success {
			logs.Error("user [%d] win gold [%v] add failed err:%v", winResult.UserId, winResult.WinGold, err)
		}
	}

	//持久化
	conn := GameConf.RedisConf.RedisPool.Get()
	defer conn.Close()
	//todo 完善
	for client, stakeGoldMap := range p.StakeInfoMap {
		var userAllStake, winGold int
		for _, stakeGold := range stakeGoldMap {
			userAllStake = userAllStake + stakeGold
		}
		for _, userWin := range userWinResult {
			if userWin.UserId == client.UserInfo.UserId {
				winGold = winGold + userWin.WinGold
			}
		}
		stakeDetail, err := json.Marshal(stakeGoldMap)
		if err != nil {
			logs.Error("json marsha1 user stake detail [%v] err:%v", stakeGoldMap, err)
			stakeDetail = []byte{}
		}
		data := common.InvestUserStake{
			GameTimesId:  gameTimesId,
			Periods:      periods,
			RoomId:       int(p.RoomId),
			RoomType:     0,
			UserId:       int(client.UserInfo.UserId),
			Nickname:     client.UserInfo.Nickname,
			UserAllStake: userAllStake,
			WinGold:      winGold,
			StakeDetail:  string(stakeDetail),
			GameResult:   gameResult,
			Pool:         pool,
			StakeTime:    stakeTime,
		}

		dataStr, err := json.Marshal(data)
		if err != nil {
			logs.Error("json marsha1 user stake msg err:%v", err)
			return
		}
		//logs.Debug("lPush [%v] total user stake msg [%v]", GameConf.RedisKey.RedisKeyUserStake,string(dataStr))
		_,err = conn.Do("lPush", GameConf.RedisKey.RedisKeyUserStake, string(dataStr))
		if err != nil {
			logs.Error("lPush user stake msg [%v] err:%v",GameConf.RedisKey.RedisKeyUserStake,err)
		}
	}
}

//调用处外层 已对 hubMgr 加锁 (读锁)
func (p *Room) UserLoginResponse(client *Client) {
	//协程内发送，避免读锁长时间不能释放
	var sendLoginResult = func() {}
	var sendOtherMembers = func() {}
	var sendEndStake = func() {}
	var sendPushResult = func() {}
	var sendRoomWinResult = func() {}
	var sendShowResult = func() {}
	defer func() {
		go func() {
			sendLoginResult()
			sendOtherMembers()
			sendEndStake()
			sendPushResult()
			sendRoomWinResult()
			sendShowResult()
		}()
	}()

	p.Lock.RLock()
	defer p.Lock.RUnlock()

	var data = make(map[string]interface{})
	data["act"] = "login_result"
	data["user_obj"] = map[string]interface{}{
		"userId":     client.UserInfo.UserId,
		"nickName":   client.UserInfo.Nickname,
		"avatarAuto": client.UserInfo.Icon,
		"sex":        1,
		"gold":       client.UserInfo.Gold,
		"goldenBean": client.UserInfo.Gold,
		"diamond":    100000,
		"level":      12,
		"vipLevel":   1,
	}

	if stakeInfoMap, ok := p.StakeInfoMap[client]; ok {
		selfStake := make([]map[string]int, 0)
		for positionId, stakeGold := range stakeInfoMap {
			stakeItem := map[string]int{"position": positionId, "stake_gold": stakeGold}
			selfStake = append(selfStake, stakeItem)
		}
		data["self_stake"] = selfStake
	}

	stakeInfo := make(map[int]int, 0)
	for _, stakeInfoItem := range p.StakeInfoMap {
		for position, stakeGold := range stakeInfoItem {
			if _, ok := stakeInfo[position]; ok {
				stakeInfo[position] = stakeInfo[position] + stakeGold
			} else {
				stakeInfo[position] = stakeGold
			}
		}
	}

	stakeInfoMap := make([]map[string]int, 0)
	for position, stakeGold := range stakeInfo {
		stakeInfoMapItem := map[string]int{
			"position":   position,
			"stake_gold": stakeGold,
		}
		stakeInfoMap = append(stakeInfoMap, stakeInfoMapItem)
	}
	data["stake_info"] = stakeInfoMap

	userInRoom := make([]map[string]interface{}, 0)
	for _, roomClient := range p.RoomClients {
		userInRoom = append(userInRoom, map[string]interface{}{
			"user_id":  roomClient.UserInfo.UserId,
			"nickname": roomClient.UserInfo.Nickname,
			"icon":     roomClient.UserInfo.Icon,
			"gold":     roomClient.UserInfo.Gold,
		})
	}
	for _, roomClient := range p.LoginOutMap {
			userInRoom = append(userInRoom, map[string]interface{}{
			"user_id":  roomClient.UserInfo.UserId,
			"nickname": roomClient.UserInfo.Nickname,
			"icon":     roomClient.UserInfo.Icon,
			"gold":     roomClient.UserInfo.Gold,
		})
	}
	data["user_in_room"] = userInRoom

	HubMgr.GameInfo.Lock.RLock()
	defer HubMgr.GameInfo.Lock.RUnlock()
	data["last_win_result"] = HubMgr.GameInfo.GameResult.Id
	data["pool"] = HubMgr.GameInfo.Pool
	data["periods"] = HubMgr.GameInfo.Periods

	data["chips_config"] = common.CanStakeChipConf
	HubMgr.GameManage.Lock.RLock()
	defer HubMgr.GameManage.Lock.RUnlock()
	data["timer"] = HubMgr.GameManage.StakeCountdown

	loginResultMsg, err := json.Marshal(data)
	if err != nil {
		logs.Error("send room win result data marsha1 err:%v", err)
		return
	}
	sendLoginResult = func() {
		client.sendMsg(loginResultMsg)
	}

	//通知其他人
	var NoticeOthers = make(map[string]interface{})
	NoticeOthers["act"] = "room_user_info"
	NoticeOthers["user_in_room"] = data["user_in_room"]
	sendOtherMembersMsg, err := json.Marshal(NoticeOthers)
	if err != nil {
		logs.Error("send room win result data marsha1 err:%v", err)
		return
	}
	sendOtherMembers = func() {
		p.Lock.RLock()
		defer p.Lock.RUnlock()
		for _, roomClient := range p.RoomClients {
			if roomClient != client {
				roomClient.sendMsg(sendOtherMembersMsg)
			}
		}
	}

	//流畅性处理
	if HubMgr.GameManage.GameStatus == common.StatusEndStake || HubMgr.GameManage.GameStatus == common.StatusShowResult {
		if HubMgr.GameManage.GameStatus == common.StatusEndStake {
			data := &Message{
				Act:   "change_state",
				State: "endstake",
			}
			endStakeMsg, err := json.Marshal(data)
			if err != nil {
				logs.Error("send room win result data marsha1 err:%v", err)
				return
			}
			sendEndStake = func() {
				client.sendMsg(endStakeMsg)
			}
		}

		gameResult := map[string]interface{}{
			"act":          "push_result",
			"win_result":   HubMgr.GameInfo.GameResult.Id,
			"win_position": HubMgr.GameInfo.GameResult.WinPosition,
		}
		pushResultMsg, err := json.Marshal(gameResult)
		if err != nil {
			logs.Error("send room win result data marsha1 err:%v", err)
			return
		}
		sendPushResult = func() {
			client.sendMsg(pushResultMsg)
		}

		data := &Message{
			Act:      "send_room_win_result",
			NewStake: make(map[string]map[string]int),
		}
		for userId, client := range p.RoomClients {
			userWInResult := &UserWinResult{
				UserId: userId,
			}
			for _, positionId := range HubMgr.GameInfo.GameResult.WinPosition {
				if stakeGold, ok := p.StakeInfoMap[client][positionId]; ok {
					if rate, ok := common.RateMap[positionId]; ok {
						userWInResult.WinGold = userWInResult.WinGold + int(float32(stakeGold)*rate)
					} else {
						logs.Error("read position [%d] rate failed ", positionId)
					}
				}
			}
			if userWInResult.WinGold > 0 {
				data.UserWinResult = append(data.UserWinResult, userWInResult)
			}
		}
		if len(data.UserWinResult) > 0 {
			roomWinResultMsg, err := json.Marshal(data)
			if err != nil {
				logs.Error("send room win result data marsha1 err:%v", err)
				return
			}
			sendRoomWinResult = func() {
				client.sendMsg(roomWinResultMsg)
			}
		}

		if HubMgr.GameManage.GameStatus == common.StatusShowResult {
			data := &Message{
				Act:   "change_state",
				State: "show_result",
			}
			showResultMsg, err := json.Marshal(data)
			if err != nil {
				logs.Error("send room win result data marsha1 err:%v", err)
				return
			}
			sendShowResult = func() {
				client.sendMsg(showResultMsg)
			}
		}
	}
}
