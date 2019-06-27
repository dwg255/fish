package main

import (
	"crypto/md5"
	"fish/game/common"
	_ "fish/game/router"
	"fish/game/service"
	"flag"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func main() {
	err := service.LoadTraceFile("./common/conf/traces.json")
	if err != nil {
		logs.Error("LoadTraceFile err: %v", err)
		return
	}
	fmt.Println("run")
	err = initConf()
	if err != nil {
		logs.Error("init conf err: %v", err)
		return
	}
	logs.Debug("init config succ")

	err = initSec()
	if err != nil {
		logs.Error("init sec err: %v", err)
		return
	}
	logs.Debug("initSec succ")

	var addr = flag.String("addr", fmt.Sprintf(":%d", common.GameConf.GamePort), "http service address")

	logs.Debug("game server listen port %v", *addr)

	go func() {
		heartBeatTicker := time.NewTicker(time.Second * 2)
		for {
			select {
			case <-heartBeatTicker.C:
				params := url.Values{}

				Url, err := url.Parse(fmt.Sprintf("http://%s:%d/register_game_server", common.GameConf.HallHost, common.GameConf.HallPort))
				if err != nil {
					panic(err.Error())
				}
				params.Set("gameHost", common.GameConf.GameHost)
				params.Set("gamePort", strconv.Itoa(common.GameConf.GamePort))
				t := time.Now().Unix()
				params.Set("t", strconv.Itoa(int(t)))
				sign := fmt.Sprintf("%x", md5.Sum([]byte(common.GameConf.HallSecret+strconv.Itoa(int(t)))))
				params.Set("sign", sign)

				service.RoomMgr.RoomLock.Lock()
				params.Set("load", strconv.Itoa(len(service.RoomMgr.RoomsInfo))) //只传房间数，因为游戏服务器总是优先填满不满员的房间

				service.RoomMgr.RoomLock.Unlock()

				Url.RawQuery = params.Encode()
				urlPath := Url.String()
				resp, err := http.Get(urlPath)
				s, err := ioutil.ReadAll(resp.Body)
				if string(s) != "success" {
					logs.Info("register game server response:", string(s))
				}
				_ = resp.Body.Close()
			}
		}
	}()
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		logs.Error("ListenAndServe err: %v", err)
	}
}
