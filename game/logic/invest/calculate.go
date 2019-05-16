package invest

import (
	"math/rand"
	"game/invest/common"
	"github.com/astaxie/beego/logs"
	"time"
)

func getResult() (resultInvestConf *common.InvestConf,needGold int) {
	//gameInfo.Lock.Lock()
	//defer gameInfo.Lock.Unlock()
	rand.Seed(time.Now().UnixNano())
	randInt := rand.Intn(common.TotalWeight)
	var weightAdd int
	for _,investConf := range common.InvestConfArr{
		weightAdd = weightAdd + investConf.Weight
		if investConf.Weight == 0 {
			continue
		}
		if randInt <= weightAdd {
			resultInvestConf = investConf
			break
		}
	}
	winPositionArr := resultInvestConf.GetParents()
	//logs.Debug("total stake map:%v",gameInfo.StakeMap)
	for _,investConf := range winPositionArr{
		stakeGold,ok := gameInfo.StakeMap[investConf.Id]
		if !ok {
			//logs.Error("user total stake in [%d] is nil",investConf.Id)
			continue
		}
		needGold = needGold + int(float32(stakeGold) * investConf.Rate)
	}
	if needGold <= gameInfo.Pool{
		return
	}
	logs.Debug("gold in pool not enough ,pool gold %d ;need: %d",gameInfo.Pool,needGold)

	var legalResultMap = make(map[*common.InvestConf]int)
	needMinGold := needGold
	for _,investConf := range common.InvestConfArr{
		if investConf.Weight != 0 {
			winPositionArr := investConf.GetParents()
			itemNeedGold := 0
			for _,itemInvestConf := range winPositionArr{
				stakeGold,ok := gameInfo.StakeMap[itemInvestConf.Id]
				if !ok {
					//logs.Debug("user total stake in [%d] is nil",investConf.Id)
					continue
				}
				itemNeedGold = itemNeedGold + int(float32(stakeGold) * investConf.Rate)
			}
			if itemNeedGold <= gameInfo.Pool{
				needGold = itemNeedGold
				resultInvestConf = investConf
				if investConf.Id != 19 {
					legalResultMap[investConf] = itemNeedGold
				}
				//return
			}
			if needMinGold > itemNeedGold {
				needMinGold = itemNeedGold
				resultInvestConf = investConf
			}
		}
	}
	if len(legalResultMap) > 0 {	//合法结果不止一个时，随机取一个
		for resultInvestConf,needMinGold = range legalResultMap{
			break
		}
	}
	return
}

