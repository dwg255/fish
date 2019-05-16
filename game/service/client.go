package service

import (
	"bytes"
	"net/http"
	"time"
	"github.com/gorilla/websocket"
	"github.com/astaxie/beego/logs"
	"encoding/json"
)

const (
	writeWait = 1 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMessageSize = 512

	clientStatusGuest = 0
	clientStatusLogin = 1
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

type UserInfo struct {
	UserId   UserId `json:"user_id"`
	Nickname string `json:"nickname"`
	Icon     string `json:"icon"`
	Gold     int    `json:"gold"`
}

type Client struct {
	Hub         *Hub
	conn        *websocket.Conn
	UserInfo    *UserInfo
	Status      int
	Room        *Room
}

func (c *Client) sendMsg(msg []byte)  {
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		c.Room.Lock.Lock()
		if _,ok := c.Room.RoomClients[c.UserInfo.UserId];ok {
			delete(c.Room.RoomClients,c.UserInfo.UserId)
		}
		c.Room.Lock.Unlock()

		c.conn.Close()
		c.Room.LoginOutChan <- c
	}
	w.Write(msg)
	if err := w.Close(); err != nil {
		c.Room.Lock.Lock()
		if _,ok := c.Room.RoomClients[c.UserInfo.UserId];ok {
			delete(c.Room.RoomClients,c.UserInfo.UserId)
		}
		c.Room.Lock.Unlock()

		c.conn.Close()
		c.Room.LoginOutChan <- c
	}
}

func (c *Client) readPump() {
	defer func() {
		if c.Status == clientStatusLogin {
			c.Room.Lock.Lock()
			if _,ok := c.Room.RoomClients[c.UserInfo.UserId];ok {
				delete(c.Room.RoomClients,c.UserInfo.UserId)
			}
			c.Room.Lock.Unlock()
			c.Room.LoginOutChan <- c
		}
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logs.Error("websocket user_id[%d] unexpected close error: %v", c.UserInfo.UserId, err)
			}
			return
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		logs.Debug("receive message from client:%s", message)
		data := make(map[string]interface{})
		err = json.Unmarshal(message, &data)
		if err != nil {
			logs.Error("message unmarsha1 err, user_id[%d] err:%v", c.UserInfo.UserId, err)
		} else {
			wsRequest(data, c)
		}
	}
}

func ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logs.Error("upgrader err:%v", err)
		return
	}
	client := &Client{Hub: HubMgr, conn: conn, UserInfo: &UserInfo{}}

	go client.readPump()
}
