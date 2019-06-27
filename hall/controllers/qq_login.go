package controllers

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"net/http"
)
var (
	appId = 101673379
	AppKey = "c18b1b56f2f88ef423bfeadbad9a816c"
	redirectUri = "http://fish.blzz.shop/qq/message"
)
func QQLogin(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("QQLogin panic:%v ", r)
		}
	}()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	//qqLoginUrl := fmt.Sprintf("https://graph.qq.com/oauth2.0/authorize?response_type=code&client_id=%d&redirect_uri=%s&state=%d&display=%s",appId,redirectUri,1,"mobile")
	qqLoginUrl := fmt.Sprintf("https://graph.qq.com/oauth2.0/authorize?response_type=code&client_id=%d&redirect_uri=%s&state=%d&display=%s",appId,redirectUri,1,"pc")
	http.Redirect(w,r,qqLoginUrl,302)
}
