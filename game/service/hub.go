package service

import (
	"sync"
	"github.com/astaxie/beego/logs"
	"game/invest/common"
)

type RoomId int
type UserId int

type Hub struct {
	Rooms             map[*Room]int			//房间人数映射
	UserToRoom        map[UserId]*Room
	CenterCommandChan chan *common.Message //center发送的游戏状态

	LoginChan    chan *Client //登录房间
	RoomIdInr    RoomId       //获取自增room_id
	GameInfo     *common.GameInfo
	GameManage   *common.GameManage //游戏状态
	LoginOutMap  map[UserId]*Client
	Lock         sync.RWMutex
}

func NewHub() (hub *Hub, err error) {
	hub = &Hub{
		Rooms:             make(map[*Room]int),
		UserToRoom:        make(map[UserId]*Room),
		CenterCommandChan: make(chan *common.Message, 10),
		LoginChan:         make(chan *Client, 1000),
		LoginOutMap:       make(map[UserId]*Client, 1000),
	}
	return
}

func (h *Hub) run() {
	for {
		select {
		case cli := <-h.LoginChan:
			h.Lock.Lock()
			//判断是否重复登录
			if _,ok := h.UserToRoom[cli.UserInfo.UserId];ok {
				h.Lock.Unlock()
				continue
			}
			//判断是否重连用户
			if prevCli, ok := h.LoginOutMap[cli.UserInfo.UserId]; ok {
				room := prevCli.Room
				cli.Hub = HubMgr
				cli.Room = room
				cli.Status = clientStatusLogin

				h.UserToRoom[cli.UserInfo.UserId] = room
				h.Lock.Unlock()
				room.Lock.Lock()
				if previousCli, ok := room.LoginOutMap[cli.UserInfo.UserId]; ok {
					delete(room.LoginOutMap, cli.UserInfo.UserId)
					room.RoomClients[cli.UserInfo.UserId] = cli
					if _, ok := room.StakeInfoMap[previousCli]; ok {
						room.StakeInfoMap[cli] = room.StakeInfoMap[previousCli]
						delete(room.StakeInfoMap, previousCli)
					}
					if _, ok := room.WaitSendStakeInfoMap[previousCli]; ok {
						room.WaitSendStakeInfoMap[cli] = room.WaitSendStakeInfoMap[previousCli]
						delete(room.WaitSendStakeInfoMap, previousCli)
					}
				} else {
					logs.Error("user [%d] in hub,but not in room [%d]", cli.UserInfo.UserId, room.RoomId)
				}
				delete(h.LoginOutMap, cli.UserInfo.UserId)
				room.Lock.Unlock()

				h.Lock.RLock()
				room.UserLoginResponse(cli)
				h.Lock.RUnlock()

				continue
			}
			h.Lock.Unlock()

			//加入游戏
			h.Lock.RLock()
			success := false
			for room, count := range h.Rooms {
				if count < 5 {

					h.Lock.RUnlock()	//释放读锁，方法内加写锁
					ok, err := room.IntoRoom(cli)
					h.Lock.RLock()

					if err != nil {
						logs.Error("user [%d] into room failed,err:%v", cli.UserInfo.UserId, err)
						continue
					}
					if ok {
						logs.Debug("user into room succ")
						success = true
						//todo 通知房间用户
						room.UserLoginResponse(cli)
						break
					}
				}
			}
			if !success {
				room := cli.NewRoom()
				room.UserLoginResponse(cli)
			}
			h.Lock.RUnlock()
		}
	}
}

func (h *Hub) broadcast() {
	for {
		select {
		case message := <-h.CenterCommandChan:
			//logs.Debug("game status %v",message.Type)
			for room := range h.Rooms {
				room.TeamRadio <- message
			}
			if message.Type == common.StatusPrepare {
				h.Lock.Lock()
				for _, cli := range h.LoginOutMap {
					if _, ok := h.UserToRoom[cli.UserInfo.UserId]; ok {
						delete(h.UserToRoom, cli.UserInfo.UserId)
					}
				}
				h.LoginOutMap = make(map[UserId]*Client)
				h.Lock.Unlock()
			}
		}
	}
}
