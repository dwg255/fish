package controllers

import (
	"context"
	"encoding/json"
	"fish/common/api/thrift/gen-go/rpc"
	"fish/common/tools"
	"fish/hall/common"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type qqUserInfo struct {
	Ret                int `json:"ret"`
	Msg                string `json:"msg"`
	IsLost            int `json:"is_lost"`
	Nickname           string `json:"nickname"`
	Gender             string `json:"gender"`
	Province           string `json:"province"`
	City               string `json:"city"`
	Year               string `json:"year"`
	Constellation      string `json:"constellation"`
	FigureUrl          string `json:"figureurl"`
	FigureUrl1        string `json:"figureurl_1"`
	FigureUrl2        string `json:"figureurl_2"`
	FigureUrlQQ1     string `json:"figureurl_qq_1"`
	FigureUrlQQ2     string `json:"figureurl_qq_2"`
	FigureUrlQQ       string `json:"figureurl_qq"`
	FigureUrlType     string `json:"figureurl_type"`
	IsYellowVip      string `json:"is_yellow_vip"`
	Vip                string `json:"vip"`
	YellowVipLevel   string `json:"yellow_vip_level"`
	Level              string `json:"level"`
	IsYellowYearVip string `json:"is_yellow_year_vip"`
}

func QQCallback(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("QQCallback panic:%v ", r)
		}
	}()
	var sign string
	defer func() {
		var script = fmt.Sprintf(`<script>localStorage.setItem("sign","%s");location.href="/"</script>`,sign)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if _, err := w.Write([]byte(script)); err != nil {
			logs.Error("QQCallback err: %v", err)
		}
	}()
	r.ParseForm()
	code := r.FormValue("code")
	if len(code) == 0 {
		logs.Error("QQCallback :code is null")
	}
	getTokenUrl := fmt.Sprintf("https://graph.qq.com/oauth2.0/token?grant_type=authorization_code&client_id=%d&client_secret=%s&code=%s&redirect_uri=%s", appId, AppKey, code, redirectUri)
	resp, err := http.Get(getTokenUrl)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	m, _ := url.ParseQuery(string(body))
	if len(m["access_token"]) > 0 {
		accessToken := m["access_token"][0]

		getOpenIdUrl := fmt.Sprintf("https://graph.qq.com/oauth2.0/me?access_token=%s", accessToken)
		resp, err := http.Get(getOpenIdUrl)
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		callback := string(body)
		//callback( {"client_id":"101673379","openid":"EC0F4A930140B2581EDC71A08D824985"} );
		start := strings.Index(callback, "{")
		end := strings.Index(callback, "}")
		retMapStr := callback[start : end+1]
		retMap := make(map[string]string)
		if err := json.Unmarshal([]byte(retMapStr), &retMap); err != nil {
			panic(err)
		}
		if len(retMap["openid"]) != 0 {
			getUserInfoUrl := fmt.Sprintf("https://graph.qq.com/user/get_user_info?access_token=%s&oauth_consumer_key=%d&openid=%s", accessToken, appId, retMap["openid"])
			resp, err := http.Get(getUserInfoUrl)
			if err != nil {
				panic(err)
			}

			defer resp.Body.Close()
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			qqUserInfo := &qqUserInfo{}
			if err := json.Unmarshal([]byte(body), qqUserInfo); err != nil {
				panic(err)
			}

			fmt.Println(string(body))
			fmt.Println(qqUserInfo)
			fmt.Println(qqUserInfo.Nickname)
			if client, closeTransportHandler, err := tools.GetRpcClient(common.HallConf.AccountHost, strconv.Itoa(common.HallConf.AccountPort)); err == nil {
				defer func() {
					if err := closeTransportHandler(); err != nil {
						logs.Error("close rpc err: %v", err)
					}
				}()
				var sex int8
				if qqUserInfo.Gender != "男" {
					sex = 1
				}
				if resp, err := client.CreateQQUser(context.Background(), &rpc.UserInfo{
					UserName: qqUserInfo.Nickname,
					NickName: qqUserInfo.Nickname,
					Sex:      sex,
					HeadImg:  "1",
					Lv:       int32(rand.Intn(7)),
					Exp:      0,
					Vip:      int8(rand.Intn(7)),
					Gems:     10000,
					Ice:      10,
					QqInfo: &rpc.QqInfo{
						OpenId:        retMap["openid"],
						FigureUrl:     qqUserInfo.FigureUrlQQ1,
						Province:      qqUserInfo.Province,
						City:          qqUserInfo.City,
						TotalSpending: 0,
					},
				}); err == nil {
					sign = resp.UserObj.Token
				}
			}
		}
	}
}
