package router

import (
	"fish/game/service"
	"net/http"
	"fish/game/controllers/invest"
)

func init()  {
	http.HandleFunc("/api/user/getUserInfo", invest.GetUserInfo)

	http.HandleFunc("/api/invest/register", invest.Register)
	http.HandleFunc("/api/invest/stake", invest.Stake)
	http.HandleFunc("/api/invest/rewardRecord",invest.RewardRecord)
	http.HandleFunc("/api/invest/recordList",invest.RecordList)
	http.HandleFunc("/api/invest/recordDetail",invest.RecordDetail)

	http.HandleFunc("/",service.ServeWs)
}
