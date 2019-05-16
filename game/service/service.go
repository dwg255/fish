package service

import (
	"github.com/astaxie/beego/logs"
	"game/invest/logic/invest"
	"game/invest/common"
)

var (
	GameConf *common.GameConf
	HubMgr   *Hub
)

func InitService(serviceConf *common.GameConf) {
	GameConf = serviceConf
	var err error
	HubMgr, err = NewHub()
	if err != nil {
		logs.Error("new newHub err :%v", err)
		panic("new newHub err ")
	}
	go HubMgr.run()
	go HubMgr.broadcast()
	HubMgr.GameInfo, HubMgr.GameManage = invest.InitInvest(HubMgr.CenterCommandChan, GameConf)
	go invest.RunLogic()
	logs.Debug("init service succ")
}

