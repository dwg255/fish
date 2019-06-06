package service

import (
	"encoding/json"
	"fish/common/tools"
	"fmt"
	"github.com/astaxie/beego/logs"
	"sync"
	"time"
)

type roomMgr struct {
	RoomLock   sync.Mutex
	RoomsInfo  map[RoomId]*RoomInfo //room暴露出去的信息和channel
	Rooms      map[RoomId]*room
	RoomIdChan <-chan int64
}

type RoomInfo struct {
	UserInfo    []UserId
	HttpReqChan chan *HttpReqData
	BaseScore   int //差点忘了请求的房间类型要一致才能进入
}

var (
	RoomMgr = &roomMgr{
		RoomLock:   sync.Mutex{},
		RoomsInfo:  make(map[RoomId]*RoomInfo),
		Rooms:      make(map[RoomId]*room), //只在room的协程里操作
		RoomIdChan: make(<-chan int64),
	}
)

const (
	GameStatusWaitBegin = iota
	GameStatusFree
	GameStatusPlay
	GameStatusFormation
	GameStatusFrozen
)

type RoomId int64

type room struct {
	RoomId        RoomId
	ActiveFish    []*Fish //待激活的鱼
	CreateTime    time.Time
	Users         map[UserId]*UserInfo
	Conf          *RoomConf
	FrozenEndTime time.Time
	//FormationEndTime  time.Time
	Status            int
	AliveFish         map[FishId]*Fish
	AliveBullet       map[BulletId]*Bullet
	Utils             *FishUtil
	fishArrayEndTimer <-chan time.Time
	frozenEndTimer    <-chan time.Time
	Exit              chan bool
	ClientReqChan     chan *clientReqData //todo 客户端的请求通过chan传递，省去加锁的写法.包括加入房间
	HttpReqChan       chan *HttpReqData
}

type clientReqData struct {
	client  *Client
	reqData []string
}

type HttpReqData struct {
	UserInfo UserInfo
	ErrChan  chan error
}

type RoomConf struct {
	BaseScore    int    `json:"gamebasescore"`
	MinHaveScore int    `json:"minhavescore"`
	MaxHaveScore int    `json:"maxhavescore"`
	TaxRatio     int    `json:"-"` //抽水 千分之
	Creator      UserId `json:"creator"`
}

func init() {
	if err := initGenerateUidTool(); err != nil {
		panic(err)
	}
}

func initGenerateUidTool() (err error) {
	if err, RoomMgr.RoomIdChan = tools.GenerateUid(1); err != nil {
		logs.Error("GenerateUid err: %v", err)
		return
	}
	return
}

func CreatePublicRoom(roomConf *RoomConf) (roomId RoomId) {
	//方法外获得锁
	roomId = RoomId(<-RoomMgr.RoomIdChan)
	RoomMgr.Rooms[roomId] = &room{
		RoomId:        roomId,
		ActiveFish:    make([]*Fish, 0),
		CreateTime:    time.Now(),
		Users:         make(map[UserId]*UserInfo, 4),
		Conf:          roomConf,
		FrozenEndTime: time.Time{},
		//FormationEndTime: time.Time{},
		Status:      GameStatusWaitBegin,
		AliveFish:   make(map[FishId]*Fish),
		AliveBullet: make(map[BulletId]*Bullet),
		Utils: &FishUtil{
			//ActiveFish: make([]*Fish, 0),
			//Lock:       sync.Mutex{},
			BuildFishChan:    make(chan *Fish, 10),
			StopBuildFish:    make(chan bool),    //暂停出鱼
			RestartBuildFish: make(chan bool),    //重新开始出鱼
			Exit:             make(chan bool, 1), //接收信号
		},
		fishArrayEndTimer: make(<-chan time.Time),
		frozenEndTimer:    make(<-chan time.Time),
		Exit:              make(chan bool, 1),
		ClientReqChan:     make(chan *clientReqData),
		HttpReqChan:       make(chan *HttpReqData),
	}
	RoomMgr.RoomsInfo[roomId] = &RoomInfo{
		UserInfo: make([]UserId, 0),
		//ClientReqChan: make(chan *clientReqData),
		HttpReqChan: RoomMgr.Rooms[roomId].HttpReqChan,
		BaseScore:   roomConf.BaseScore,
	}

	RoomMgr.Rooms[roomId].begin()
	return
}

func (room *room) EnterRoom(userInfo *UserInfo) (err error) {
	logs.Debug("user %d request enter room %v", userInfo.UserId, room.RoomId)

	userCount := len(room.Users)
	if userCount >= 4 {
		logs.Error("enterRoom err: room [%v] is full", room.RoomId)
		return
	}
	seatIndex := -1
out:
	for i := 0; i < 4; i++ {
		for _, roomUserInfo := range room.Users {
			if roomUserInfo.SeatIndex == i {
				continue out
			}
		}
		seatIndex = i
		break
	}

	if seatIndex == -1 {
		return fmt.Errorf("enterRoom  roomId [%v] failed", room.RoomId)
	}
	userInfo.SeatIndex = seatIndex
	room.Users[userInfo.UserId] = userInfo
	return
}

func (room *room) begin() {
	logs.Debug("room %d begin", room.RoomId)
	buildNormalFishTicker := time.NewTicker(time.Second * 1)     //普通鱼每秒刷一次
	buildGroupFishTicker := time.NewTicker(time.Second * 5 * 60) //鱼群
	flushTimeOutFishTicker := time.NewTicker(time.Second * 5)    //清理过期鱼

	go func() {
		defer func() {
			logs.Trace("room %v exit...", room.RoomId)
			buildNormalFishTicker.Stop()
			buildGroupFishTicker.Stop()
			flushTimeOutFishTicker.Stop()
			room.Utils.Exit <- true
			go func() { //启动协程取数据，防止utils阻塞在出鱼阶段导致无法退出 :)
				for range room.Utils.BuildFishChan {
				}
			}()
			close(room.Exit)
			close(room.HttpReqChan)
			close(room.ClientReqChan)

			RoomMgr.RoomLock.Lock()
			logs.Info("exit room goroutine get lock...")
			defer RoomMgr.RoomLock.Unlock()
			defer logs.Info("exit room goroutine set free lock...")

			delete(RoomMgr.Rooms, room.RoomId)
		}()
		//defer room.Wg.Done()
		for {
			select {
			case <-buildNormalFishTicker.C:
				room.flushFish()
			case <-buildGroupFishTicker.C:
				if room.Status != GameStatusFree {
					continue
				}
				room.Utils.StopBuildFish <- true
				room.AliveFish = make(map[FishId]*Fish) //清理鱼
				room.buildFormation()
			case <-flushTimeOutFishTicker.C:
				now := time.Now()
				AliveFishCheck := make(map[FishId]*Fish)
				for _, fish := range room.AliveFish {
					if now.Sub(fish.ActiveTime) < 60*2*time.Second {
						AliveFishCheck[fish.FishId] = fish
					}
				}
				room.AliveFish = AliveFishCheck
			case fish := <-room.Utils.BuildFishChan:
				room.ActiveFish = append(room.ActiveFish, fish)
			case clientReq := <-room.ClientReqChan:
				//logs.Debug("room [%d] receive client message %v", room.RoomId, clientReq.reqData)
				handleUserRequest(clientReq)
			case httpReq := <-room.HttpReqChan:
				httpReq.ErrChan <- room.EnterRoom(&httpReq.UserInfo)
				close(httpReq.ErrChan)
			case <-room.Exit:
				return
			case <-room.fishArrayEndTimer:
				room.Status = GameStatusFree
				room.Utils.RestartBuildFish <- true
			case <-room.frozenEndTimer:
				room.Status = GameStatusFree
				room.Utils.RestartBuildFish <- true
			}
		}
	}()
}

func (room *room) flushFish() {
	if room.Status != GameStatusFree {
		return
	}
	newFish := make([]*Fish, 0)
	for _, fish := range room.ActiveFish {
		if _, ok := room.AliveFish[fish.FishId]; ok {
			continue
		}
		//if len(room.AliveFish) < 30 {
		room.AliveFish[fish.FishId] = fish
		newFish = append(newFish, fish)
		//} else {
		//	break
		//}
	}
	room.ActiveFish = make([]*Fish, 0)
	if len(newFish) > 0 {
		room.broadcast([]interface{}{"build_fish_reply", newFish})
	}
}

func (room *room) buildFormation() {
	if room.Status != GameStatusFree {
		room.Utils.RestartBuildFish <- true
		return
	}
	room.Status = GameStatusFormation
	fishArrayData := BuildFishArray()
	activeTime := time.Now()
	for _, fishArray := range fishArrayData.FishArray {
		for _, arrayFish := range fishArray {
			room.AliveFish[arrayFish.FishId] = &Fish{
				FishId:     arrayFish.FishId,
				FishKind:   arrayFish.FishKind,
				Speed:      0,
				ActiveTime: activeTime,
			}
		}
	}
	room.fishArrayEndTimer = time.After(fishArrayData.EndTime.Sub(time.Now()))
	room.broadcast([]interface{}{
		"build_fishArray_reply",
		fishArrayData,
	})
}

func (room *room) getBombFish() (killedFishes []*Fish) {
	for _, fish := range room.AliveFish {
		if len(killedFishes) == 20 {
			return
		}
		if fish.FishKind < FishKind11 {
			killedFishes = append(killedFishes, fish)
		}
	}
	return
}

//一网打尽
func (room *room) getAllInOne(oneFish *Fish) (killedFishes []*Fish) {
	for _, fish := range room.AliveFish {
		if fish.FishKind >= FishKind23 && fish.FishKind <= FishKind26 {
			killedFishes = append(killedFishes, fish)
		}
	}
	return
}

//同类炸弹
func (room *room) getSameFish(oneFish *Fish) (killedFishes []*Fish) {
	switch oneFish.FishKind {
	case FishKind31:
		for _, fish := range room.AliveFish {
			if fish.FishKind == FishKind31 || fish.FishKind == FishKind12 {
				killedFishes = append(killedFishes, fish)
			}
		}
	case FishKind32:
		for _, fish := range room.AliveFish {
			if fish.FishKind == FishKind32 || fish.FishKind == FishKind1 {
				killedFishes = append(killedFishes, fish)
			}
		}
	case FishKind33:
		for _, fish := range room.AliveFish {
			if fish.FishKind == FishKind33 || fish.FishKind == FishKind7 {
				killedFishes = append(killedFishes, fish)
			}
		}
	}
	return
}

func (room *room) broadcast(data []interface{}) {
	if dataByte, err := json.Marshal(data); err != nil {
		logs.Error("broadcast [%v] json marshal err :%v ", data, err)
	} else {
		dataByte = append([]byte{'4', '2'}, dataByte...)
		for _, userInfo := range room.Users {
			if userInfo.client != nil {
				userInfo.client.sendMsg(dataByte)
			}
		}
	}
}
