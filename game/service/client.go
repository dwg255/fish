package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fish/common/api/thrift/gen-go/rpc"
	"fish/common/tools"
	"fish/game/common"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const (
	writeWait      = 1 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, //不验证origin
}

type Client struct {
	conn      *websocket.Conn
	UserInfo  *UserInfo
	Room      *room
	msgChan   chan []byte
	closeChan chan bool
}

type UserId int64

type UserInfo struct {
	UserId          UserId  `json:"userId"`
	Score           int     `json:"-"`
	Bill            int     `json:"-"` //账单
	ConversionScore float64 `json:"score"`
	Name            string  `json:"name"`
	Ready           bool    `json:"ready"`
	SeatIndex       int     `json:"seatIndex"`
	Vip             int     `json:"vip"`
	CannonKind      int     `json:"cannonKind"`
	Power           float64 `json:"power"`
	LockFishId      FishId  `json:"lockFishId"`
	Online          bool    `json:"online"`
	client          *Client `json:"-"`
	Ip              string  `json:"ip"`
}

type BulletId string
type Bullet struct {
	UserId     UserId   `json:"userId"`
	ChairId    int      `json:"chairId"`
	BulletKind int      `json:"bulletKind"`
	BulletId   BulletId `json:"bulletId"`
	Angle      float64  `json:"angle"`
	Sign       string   `json:"sign"`
	LockFishId FishId   `json:"lockFishId"`
}
type catchFishReq struct {
	BulletId BulletId `json:"bulletId"`
	FishId   FishId   `json:"fishId"`
}

func (c *Client) sendMsg(msg []byte) {
	if c.UserInfo != nil {
		//logs.Debug("user [%v] send msg %v", c.UserInfo.UserId, string(msg))
	}
	c.msgChan <- msg //为什么此处不担心发送数据到一个已关闭的chan？  因为channel没有手动关闭而是交给gc处理  :)
}

func (c *Client) sendToClient(data []interface{}) {
	if dataByte, err := json.Marshal(data); err != nil {
		logs.Error("broadcast [%v] json marshal err :%v ", data, err)
	} else {
		dataByte = append([]byte{'4', '2'}, dataByte...)
		c.sendMsg(dataByte)
	}
}

func (c *Client) sendToOthers(data []interface{}) {
	if dataByte, err := json.Marshal(data); err != nil {
		logs.Error("broadcast [%v] json marshal err :%v ", data, err)
	} else {
		dataByte = append([]byte{'4', '2'}, dataByte...)
		for _, userInfo := range c.Room.Users {
			if userInfo.UserId == c.UserInfo.UserId || userInfo.client == nil {
				continue
			}
			userInfo.client.sendMsg(dataByte)
		}
	}
}

func (c *Client) writePump() {
	PingTicker := time.NewTicker(pingPeriod)
	defer func() {
		PingTicker.Stop()
		//close(c.closeChan)
		RoomMgr.RoomLock.Lock()
		defer RoomMgr.RoomLock.Unlock()
		if c.Room != nil { //客户端在房间内
			if _, ok := RoomMgr.RoomsInfo[c.Room.RoomId]; ok { //房间没销毁
				c.Room.ClientReqChan <- &clientReqData{c, []string{"client_exit"}}
			}
		}

		if c.UserInfo != nil {
			logs.Info("user %v write goroutine exit...", c.UserInfo.UserId)
		}
	}()
	for {
		select {
		case msg := <-c.msgChan:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				logs.Error("sendMsg SetWriteDeadline err, %v", err)
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				logs.Error("sendMsg NextWriter err, %v", err)
				return
			}
			if _, err = w.Write(msg); err != nil {
				logs.Error("sendMsg Write err, %v", err)
			}
			if err = w.Close(); err != nil {
				_ = c.conn.Close()
			}
		case <-PingTicker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logs.Debug("PingTicker write message err : %v", err)
				return
			}
		case <-c.closeChan:
			if err := c.conn.Close(); err != nil {
				if c.UserInfo != nil {
					logs.Info("user %v client conn close err : %v", err)
				}
			}
			return
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
		//RoomMgr.RoomLock.Lock()
		//defer RoomMgr.RoomLock.Unlock()
		//if c.Room != nil { //客户端在房间内
		//	if _, ok := RoomMgr.RoomsInfo[c.Room.RoomId]; ok { //房间没销毁
		//		c.Room.ClientReqChan <- &clientReqData{c, []string{"client_exit"}}
		//	}
		//}
		if c.UserInfo != nil {
			logs.Info("user %v read goroutine exit...", c.UserInfo.UserId)
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				if c.UserInfo != nil { //client.userInfo == nil && client.Room != nil 是因为他没有先登录房间
					logs.Error("websocket userId [%v] UserInfo [%d] unexpected close error: %v", c.UserInfo.UserId, &c.UserInfo, err)
					////todo test
					//if c.Room != nil {
					//	logs.Debug("users count %v", len(c.Room.Users))
					//	for userId, userInfo := range c.Room.Users {
					//		logs.Debug("user id %v, name %v", userId, userInfo.Name)
					//	}
					//}
				} else {
					logs.Error("websocket unexpected close error: %v", err)
				}
			}
			return
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		if err != nil {
			if c.UserInfo != nil {
				logs.Error("message unMarsha1 err, user_id[%d] err:%v", c.UserInfo.UserId, err)
			} else {
				logs.Error("message unMarsha1 err:%v", err)
			}
		} else {
			wsRequest(message, c)
		}
	}
}

func ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logs.Error("upgrader err:%v", err)
		return
	}
	client := &Client{conn: conn, msgChan: make(chan []byte, 100), closeChan: make(chan bool, 1), UserInfo: &UserInfo{}}
	logs.Debug("new client...")

	go client.readPump()
	go client.writePump()

	if msg, err := json.Marshal(map[string]interface{}{
		"pingInterval": 25000,
		"pingTimeout":  5000,
		"sid":          "1JDAUJeOA81mbc00AAAA",
		"upgrades":     make([]int, 0),
	}); err != nil {
		logs.Error("new client send msg err : %v", err)
	} else {
		client.sendMsg(append([]byte{'0'}, msg...))
		client.sendMsg(append([]byte{'4', '0'}, ))
	}
}

func (c *Client) Fire(bullet *Bullet) {
	if bullet.BulletKind == 22 { //激光炮
		// todo 激光炮
		c.UserInfo.Power = 0
		c.sendToOthers([]interface{}{
			"user_fire_Reply",
			bullet,
		})
		return
	}
	c.Room.AliveBullet[bullet.BulletId] = bullet
	c.sendToOthers([]interface{}{"user_fire_Reply", bullet})

	c.UserInfo.Score -= c.Room.Conf.BaseScore * GetBulletMulti(bullet.BulletKind)
	c.UserInfo.Bill -= c.Room.Conf.BaseScore * GetBulletMulti(bullet.BulletKind)
	addPower, _ := strconv.ParseFloat(fmt.Sprintf("%.5f", float64(GetBulletMulti(bullet.BulletKind))/1000), 64)
	if c.UserInfo.Power < 1 {
		c.UserInfo.Power += addPower
	}
}

func (c *Client) catchFish(fishId FishId, bulletId BulletId) {
	if bullet, ok := c.Room.AliveBullet[bulletId]; ok {
		if bullet.UserId == c.UserInfo.UserId {
			if fish, ok := c.Room.AliveFish[fishId]; ok {
				if IsHit(fish) {
					//logs.Debug("user %v,catchFish %v", c.UserInfo.UserId, fishId)
					killedFishes := []*Fish{fish}
					if fish.FishKind == FishKind30 { //全屏炸弹
						killedFishes = append(killedFishes, c.Room.getBombFish()...)
					} else if fish.FishKind >= FishKind23 && fish.FishKind <= FishKind26 { //一网打尽
						killedFishes = c.Room.getAllInOne(fish)
					} else if fish.FishKind >= FishKind31 && fish.FishKind <= FishKind33 {
						killedFishes = c.Room.getSameFish(fish)
					}
					//加钱
					addScore := 0
					for _, fish := range killedFishes {
						addScore += GetFishMulti(fish) * GetBulletMulti(bullet.BulletKind) * c.Room.Conf.BaseScore
					}
					//if addScore > c.Room.Conf.BaseScore*100 {
					//	addScore = c.Room.Conf.BaseScore * 100
					//}
					c.UserInfo.Score += addScore
					c.UserInfo.Bill += addScore //记账
					//todo %1的概率获取冰冻道具
					rand.Seed(time.Now().UnixNano())
					item := ""
					if rand.Intn(100) == 0 {
						item = "ice"
					}

					catchFishAddScore, _ := strconv.ParseFloat(fmt.Sprintf("%.5f", float64(addScore)/1000), 64)
					catchResult := []interface{}{"catch_fish_reply",
						map[string]interface{}{
							"userId":   c.UserInfo.UserId,
							"chairId":  bullet.ChairId,
							"bulletId": bullet.BulletId,
							"fishId":   strconv.Itoa(int(fish.FishId)),
							"addScore": catchFishAddScore,
							"item":     item,
						}}
					c.Room.broadcast(catchResult)
					for _, fish := range killedFishes {
						delete(c.Room.AliveFish, fish.FishId)
					}
				} else {
					//logs.Debug("hit fish failed...")
				}
			} else {
				logs.Debug("user [%v] catch fish fishId [%v] not in alive fish array...", c.UserInfo.UserId, fishId)
			}
		} else {
			logs.Debug("user [%v] catch fish bullet [%v] belong to user [%v] ...", c.UserInfo.UserId, bullet.BulletId, bullet.UserId)
		}
		delete(c.Room.AliveBullet, bulletId)
	} else {
		//客户端会多传命中，愚蠢的客户端
		//logs.Debug("user [%v] catch fish bullet [%v] not exists ...", c.UserInfo.UserId, bulletId)
	}
}

func (c *Client) laserCatchFish(data map[interface{}]interface{}) { //激光炮

}

func (c *Client) frozenScene(startTime time.Time) { //冰冻屏幕
	if c.Room.FrozenEndTime.Unix() > time.Now().Unix() {
		return
	}
	logs.Debug("frozenScene")
	c.Room.Status = GameStatusFrozen
	c.Room.Utils.StopBuildFish <- true
	c.Room.FrozenEndTime = startTime.Add(time.Second * 10)
	cutDown := c.Room.FrozenEndTime.Sub(time.Now())
	replyData := []interface{}{"user_frozen_reply", map[string]time.Duration{"cutDownTime": cutDown / 1e6}}
	c.sendToOthers(replyData)
	c.Room.frozenEndTimer = time.After(cutDown)
}

func (c *Client) exitRoom() {
	delete(c.Room.Users, c.UserInfo.UserId)
	//todo 持久化结算
}

func (c *Client) clearBill() {
	go func(userId UserId, bill int, roomId RoomId, power float64) {
		if client, closeTransportHandler, err := tools.GetRpcClient(common.GameConf.AccountHost, strconv.Itoa(common.GameConf.AccountPort)); err == nil {
			defer func() {
				if err := closeTransportHandler(); err != nil {
					logs.Error("close rpc err: %v", err)
				}
			}()
			if res, err := client.ModifyUserInfoById(context.Background(), "FISH_GAME_MODIFY", int32(userId), rpc.ModifyPropType_gems, int64(bill)); err == nil {
				if res.Code == rpc.ErrorCode_Success {
					logs.Debug("user [%v] clear bill success :in room [%v] fish game , add score : %v", userId, roomId, int64(bill))
				} else {
					logs.Debug("user [%v] clear bill failed :in room [%v] fish game , add score : %v,err code = %v", userId, roomId, int64(bill), res.Code)
				}
			} else {
				logs.Error("user [%v] clearBill [%v] err: %v", userId, bill, err)
			}
			if res, err := client.ModifyUserInfoById(context.Background(), "FISH_GAME_MODIFY", int32(userId), rpc.ModifyPropType_power, int64(power)); err == nil {
				if res.Code == rpc.ErrorCode_Success {
					logs.Debug("user [%v] clear power success :in room [%v] fish game , add power : %v", userId, roomId, int64(power))
				} else {
					logs.Debug("user [%v] clear power failed :in room [%v] fish game , add power : %v,err code = %v", userId, roomId, int64(power), res.Code)
				}
			} else {
				logs.Error("user [%v] clearBill [%v] err: %v", userId, bill, err)
			}
		}
	}(c.UserInfo.UserId, c.UserInfo.Bill, c.Room.RoomId, c.UserInfo.Power*1000)
}
