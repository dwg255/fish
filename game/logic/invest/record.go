package invest

import (
	"game/invest/common"
	"time"
)

func record() {
	for {
		date := time.Now().Format("2006-01-02")
		select {
		case gameResult := <-common.GameRecord.ResultChan:
			common.GameRecord.Lock.Lock()
			for i := 0; i < len(common.GameRecord.LasTTenTimesRank); i++ {
				if i == len(common.GameRecord.LasTTenTimesRank)-1 {
					common.GameRecord.LasTTenTimesRank[len(common.GameRecord.LasTTenTimesRank)-1] = &common.GameRank{
						Periods:    gameManage.TimesId,
						GameResult: gameResult,
					}
				} else if common.GameRecord.LasTTenTimesRank[i] == nil {
					gameManage.Lock.RLock()
					common.GameRecord.LasTTenTimesRank[i] = &common.GameRank{
						Periods:    gameManage.TimesId,
						GameResult: gameResult,
					}
					gameManage.Lock.RUnlock()
					break
				} else if common.GameRecord.LasTTenTimesRank[i+1] == nil {
					gameManage.Lock.RLock()
					common.GameRecord.LasTTenTimesRank[i+1] = &common.GameRank{
						Periods:    gameManage.TimesId,
						GameResult: gameResult,
					}
					gameManage.Lock.RUnlock()
					break
				} else {
					common.GameRecord.LasTTenTimesRank[i] = common.GameRecord.LasTTenTimesRank[i+1]
				}
			}

			lostPositionArr := make([]*common.InvestConf, 0)

			common.GameRecord.UnfortunatelyRecord = &common.UnfortunatelyPosition{}
			for _, investConf := range common.InvestConfArr {
				isWinPosition := false
				for _, winPosition := range gameResult.WinPosition {
					if investConf.Id == winPosition {
						isWinPosition = true
						break
					}
				}
				if !isWinPosition {
					//连续赢的记录清零
					lostPositionArr = append(lostPositionArr, investConf)
					common.GameRecord.ContinuedWinRecord[investConf] = 0

					//连续输的记录 +1
					if _, ok := common.GameRecord.LastTimeWinRecord[investConf.Id]; ok {
						common.GameRecord.LastTimeWinRecord[investConf.Id]++
					} else {
						common.GameRecord.LastTimeWinRecord[investConf.Id] = 1
					}

					//记录最不幸的位置
					if common.GameRecord.UnfortunatelyRecord.Count < common.GameRecord.LastTimeWinRecord[investConf.Id] {
						common.GameRecord.UnfortunatelyRecord.Count = common.GameRecord.LastTimeWinRecord[investConf.Id]
						common.GameRecord.UnfortunatelyRecord.Position = investConf.Id
					}
				} else {
					//连续赢的记录 +1
					if _, ok := common.GameRecord.ContinuedWinRecord[investConf]; ok {
						common.GameRecord.ContinuedWinRecord[investConf]++
					} else {
						common.GameRecord.ContinuedWinRecord[investConf] = 1
					}

					//连续输的记录 清零
					common.GameRecord.LastTimeWinRecord[investConf.Id] = 0

					if common.GameRecord.EverydayResultRecord.Date == date {
						if _, ok := common.GameRecord.EverydayResultRecord.Record[investConf]; ok {
							common.GameRecord.EverydayResultRecord.Record[investConf]++
						} else {
							common.GameRecord.EverydayResultRecord.Record[investConf] = 1
						}
					} else {
						common.GameRecord.EverydayResultRecord = &common.EveryDayRecord{
							Date:   date,
							Record: map[*common.InvestConf]int{investConf: 1},
						}
					}
				}
			}
			common.GameRecord.Lock.Unlock()
		}
	}
}
