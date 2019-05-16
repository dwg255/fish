package invest

import (
	"time"
	"game/invest/tools"
	"game/invest/common"
	"github.com/astaxie/beego/logs"
	"encoding/json"
)

var (
	gameConf   *common.GameConf
	gameManage = &common.GameManage{
		Timer: -1,
	}
	sendInterval      = 1 //发送押注间隔，可调节，作为服务降级
	CenterCommandChan chan *common.Message
	gameInfo          *common.GameInfo
)

func InitInvest(ch chan *common.Message, conf *common.GameConf) (GameInfo *common.GameInfo, GameManage *common.GameManage) {
	GameManage = gameManage
	CenterCommandChan = ch
	gameConf = conf
	GameInfo = &common.GameInfo{
		StakeMap:   make(map[int]int),
		GameResult: common.InvestConfArr[0],
		Pool:       10000, //初始奖池
	}
	gameInfo = GameInfo
	return
}

//启动游戏
func RunLogic() {
	go record()
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				tick()
			}
		}
	}()
}

func tick() {
	gameManage.Lock.Lock()
	defer gameManage.Lock.Unlock()
	gameManage.Timer++
	gameManage.StakeCountdown = gameManage.StakeCountdown - 1
	gameStatus, ok := common.Timer[gameManage.Timer]
	if ok {
		gameManage.GameStatus = gameStatus
		switch gameStatus {
		case common.StatusPrepare:
			gameManage.Periods++
			gameManage.Timer = 0
			gameManage.SendTime = 0
			gameManage.TimesId = tools.CreateUid()

			gameInfo.Lock.Lock()
			gameInfo.Periods = gameManage.Periods
			gameInfo.StakeMap = make(map[int]int)
			gameInfo.Lock.Unlock()
		case common.StatusStartStake:
			gameManage.StakeCountdown = 27
		case common.StatusEndStake:
			defer func() {
				var needGold int
				gameInfo.Lock.Lock()
				gameInfo.GameResult, needGold = getResult()
				common.GameRecord.ResultChan <- gameInfo.GameResult
				gameInfo.Pool = gameInfo.Pool - needGold
				gameInfo.Lock.Unlock()

				message := &common.Message{
					Type:  common.StatusSendResult,
					Timer: gameManage.StakeCountdown,
				}
				CenterCommandChan <- message

				go func() {
					//todo 持久化消息
					gameInfo.Lock.RLock()
					//logs.Debug(" get gameInfo lock")
					pool := gameInfo.Pool
					periods := gameInfo.Periods
					gameResult := gameInfo.GameResult.Id
					stakeDetail, err := json.Marshal(gameInfo.StakeMap)
					if err != nil {
						logs.Error("json marsha1 total stake detail [%v] err:%v", gameInfo.StakeMap, err)
						stakeDetail = []byte{}
					}
					gameInfo.Lock.RUnlock()

					gameManage.Lock.RLock()
					gameTimesId := gameManage.TimesId
					gameManage.Lock.RUnlock()

					conn := gameConf.RedisConf.RedisPool.Get()
					defer conn.Close()
					data := common.InvestBase{
						GameTimesId: gameTimesId,
						Periods:     periods,
						Pool:        pool,
						GameResult:  gameResult,
						StakeDetail: string(stakeDetail),
						StartTime:   time.Now().Format("2006-01-02 15:04:05"),
					}
					dataStr, err := json.Marshal(data)
					if err != nil {
						logs.Error("json marsha1 invest base err:%v", err)
						return
					}
					_, err = conn.Do("lPush", gameConf.RedisKey.RedisKeyInvestBase, string(dataStr))
					//logs.Debug("lPush [%v] total user stake msg [%v]", gameConf.RedisKey.RedisKeyInvestBase, string(dataStr))
					if err != nil {
						if err != nil {
							logs.Error("lPush total user stake msg [%v] failed, err:%v", gameConf.RedisKey.RedisKeyInvestBase, err)
						}
					}
				}()
			}()
		}
		message := &common.Message{
			Type:  gameStatus,
			Timer: gameManage.StakeCountdown,
		}
		CenterCommandChan <- message
	}
	if gameManage.GameStatus == common.StatusSendStake { //处理发送押注
		gameManage.SendTime++
		if gameManage.SendTime%sendInterval == 0 {
			message := &common.Message{
				Type:  gameManage.GameStatus,
				Timer: gameManage.StakeCountdown,
			}
			CenterCommandChan <- message
		}
	}
}
