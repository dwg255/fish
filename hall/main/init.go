package main

import (
	"encoding/json"
	"fish/hall/common"
	"github.com/astaxie/beego/logs"
)

func conversionLogLevel(logLevel string) int {
	switch logLevel {
	case "debug":
		return logs.LevelDebug
	case "warn":
		return logs.LevelWarn
	case "info":
		return logs.LevelInfo
	case "trace":
		return logs.LevelTrace
	}
	return logs.LevelDebug
}

func initLogger() (err error) {
	config := make(map[string]interface{})
	config["filename"] = common.HallConf.LogPath
	config["level"] = conversionLogLevel(common.HallConf.LogLevel)

	configStr, err := json.Marshal(config)
	if err != nil {
		return
	}
	err = logs.SetLogger(logs.AdapterFile, string(configStr))
	return
}

func initSec() (err error) {
	err = initLogger()
	if err != nil {
		return
	}
	return
}
