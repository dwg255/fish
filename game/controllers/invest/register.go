package invest

import (
	"net/http"
	"github.com/astaxie/beego/logs"
	"game/invest/tools"
	"golang.org/x/net/context"
	"game/api/thrift/gen-go/rpc"
	"time"
	"math/rand"
	"fmt"
)

func Register(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r:= recover(); r != nil {
			logs.Error("Register panic:%v ",r)
		}
	}()
	time.Now().UnixNano()
	nickname := r.FormValue("nickname")
	if len(nickname) == 0 {
		nickname = r.PostFormValue("nickname")
	}
	host := r.FormValue("host")
	if len(host) == 0 {
		host = r.PostFormValue("host")
	}
	rpcClient, closeTransport, err := tools.GetRpcClient()
	if err != nil {
		logs.Debug("get rpc client err:%v", err)
		return
	}
	defer closeTransport()

	rand.Seed(time.Now().UnixNano())
	pic := rand.Intn(20)
	picName := fmt.Sprintf("http://invest.blzz.shop/login/header_img/%d.jpg",pic)
	resp, err := rpcClient.CreateNewUser(context.Background(), nickname,picName,100000)
	//logs.Debug("user info %v",resp)
	if err != nil || resp.Code != rpc.ErrorCode_Success {
		logs.Debug("create user failed err:%v, resp:%v", err, resp)
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		url := "http://guestinvest.blzz.shop?token=" + resp.UserObj.Token + "&host=" + host
		//url := "http://invest.blzz.shop?token=" + resp.UserObj.Token
		//url := "http://invest.com/index.html?token=" + resp.UserObj.Token
		http.Redirect(w, r, url, http.StatusFound)
	}
}
