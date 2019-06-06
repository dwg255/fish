package controllers

import (
	"crypto/md5"
	"fmt"
	"github.com/astaxie/beego/logs"
	"net/http"
	"strconv"
	"time"
)

func EnterPublicRoom(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()
	minLoad := 0
	gameUrl := ""
	for serverUrl,load := range serverInfo{
		if minLoad == 0 {
			 minLoad = load
			 gameUrl = serverUrl
		}else{
			if load <= minLoad {
				minLoad = load
				gameUrl = serverUrl
			}
		}
	}
	if gameUrl == "" {
		logs.Error("no game server running ...")
		return
	}
	target := "http://" + gameUrl + r.URL.Path
	//target := "http://127.0.0.1:9001"  + r.URL.Path
	timeStamp := time.Now().Unix()
	data := []byte("t" + strconv.Itoa(int(timeStamp)))
	token := fmt.Sprintf("%x", md5.Sum(data))
	if len(r.URL.RawQuery) > 0 {
		target += "?" + r.URL.RawQuery + "&token=" + token + "&t=" + strconv.Itoa(int(timeStamp))
	} else {
		target += "?token=" + token + "&t=" + strconv.Itoa(int(timeStamp))
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}
