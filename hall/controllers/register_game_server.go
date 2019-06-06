package controllers

import (
	"crypto/md5"
	"fish/hall/common"
	"fmt"
	"github.com/astaxie/beego/logs"
	"net/http"
	"strconv"
	"sync"
)

var lock = sync.Mutex{}
var serverInfo = make(map[string]int)
//todo 优化的空间：可以加个注销服务接口。心跳加入时间，长时间未发送心跳的游戏服务器由大厅发起询问或暂时挂起。懒得做 :(

func RegisterGameServer(w http.ResponseWriter, r *http.Request) {
	ret := "failed"
	defer func() {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if _, err := w.Write([]byte(ret)); err != nil {
			logs.Error("CreateRoom err: %v", err)
		}
	}()
	gameHost := r.FormValue("gameHost")
	if len(gameHost) == 0 {
		logs.Error("load game server failed,err : invalid param gameHost %v", gameHost)
		return
	}
	gamePort := r.FormValue("gamePort")
	if len(gamePort) == 0 {
		logs.Error("load game server failed,err : invalid param gamePort %v", gamePort)
		return
	}
	loadStr := r.FormValue("load")
	if len(loadStr) == 0 {
		logs.Error("load game server failed,err : invalid param load %v ", loadStr)
		return
	}
	t := r.FormValue("t")
	if len(t) == 0 {
		logs.Error("load game server failed,err : invalid param t %v ", t)
		return
	}
	sign := r.FormValue("sign")
	if len(sign) == 0 {
		logs.Error("load game server failed,err : invalid param sign %v ", sign)
		return
	}
	if fmt.Sprintf("%x", md5.Sum([]byte(common.HallConf.HallSecret+t))) != sign {
		logs.Error("load game server failed,check sign failed  ")
		return
	}
	if loadInt, err := strconv.Atoi(loadStr); err != nil {
		logs.Error("load game server [%v:%v],err : invalid param load [%v]", gameHost, gamePort, loadStr)
		return
	} else {
		ret = "success"
		serverUrl := gameHost + ":" + gamePort
		lock.Lock()
		defer lock.Unlock()
		serverInfo[serverUrl] = loadInt
	}
}
