package main

import (
	"fish/game/common"
	"fmt"
	"github.com/astaxie/beego/config"
)

func initConf() (err error) {
	accountConf, err := config.NewConfig("ini", "./common/conf/account.conf")
	if err != nil {
		fmt.Println("new account config failed,err:", err)
		return
	}
	common.GameConf.AccountHost = accountConf.String("account_host")
	if common.GameConf.AccountHost == "" {
		return fmt.Errorf("conf err: account_host is null")
	}

	common.GameConf.AccountPort, err = accountConf.Int("account_port")
	if err != nil {
		return
	}

	hallConf, err := config.NewConfig("ini", "./common/conf/hall.conf")
	if err != nil {
		fmt.Println("new account config failed,err:", err)
		return
	}
	common.GameConf.HallHost = hallConf.String("hall_host")
	if common.GameConf.HallHost == "" {
		return fmt.Errorf("conf err: account_host is null")
	}

	common.GameConf.HallSecret = hallConf.String("hall_secret")
	if common.GameConf.HallSecret == "" {
		return fmt.Errorf("conf err: hall_secret is null")
	}

	common.GameConf.HallPort, err = hallConf.Int("hall_port")
	if err != nil {
		return
	}

	conf,err := config.NewConfig("ini","./common/conf/game.conf")
	if err != nil {
		fmt.Println("new config failed,err:",err)
		return
	}

	common.GameConf.GameHost = conf.String("game_host")
	if common.GameConf.GameHost == "" {
		return fmt.Errorf("conf err: game_host is null")
	}

	common.GameConf.GamePort,err = conf.Int("game_port")
	if err != nil {
		return
	}

	common.GameConf.LogPath = conf.String("log_path")
	if common.GameConf.LogPath == "" {
		return fmt.Errorf("conf err: log_path is null")
	}

	common.GameConf.LogLevel = conf.String("log_level")
	if common.GameConf.LogLevel == "" {
		return fmt.Errorf("conf err: log_level is null")
	}
	return
}