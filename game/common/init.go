package common

import (
	"fmt"
)

const (
	CodeSuccess = 0
	ErrorAuthFailed = 10000
	ErrorParamInvalid = 10001
	ErrorUnknownError = 10002
	ErrorGoldNotEnough = 10003
	ErrorNotStakeTime = 10004
	ErrorRpcServerError = 10005
)

func init() {
	for i, investConf := range InvestConfArr {
		if investConf.Weight > 0 {
			WeightMap[investConf.Id] = investConf.Weight
			TotalWeight = TotalWeight + investConf.Weight
			if parentsArr, ok := investConf.InitParents(); !ok {
				msg := fmt.Sprintf("init investConf [%v] parents,err", investConf)
				panic(msg)
			} else {
				InvestConfArr[i].ParentArr = parentsArr
				for _, investConf := range parentsArr {		//使用:=声明作用域
					InvestConfArr[i].WinPosition = append(InvestConfArr[i].WinPosition, investConf.Id)
				}
			}
		}
		RateMap[investConf.Id] = investConf.Rate
	}
}

var (
	GameRecord = &Record{
		ResultChan:         make(chan *InvestConf),
		ContinuedWinRecord: make(map[*InvestConf]int), //连续赢的位置
		LastTimeWinRecord:  make(map[int]int), //距离上一次赢的局数
		EverydayResultRecord: &EveryDayRecord{
			Record: make(map[*InvestConf]int),
		},
	}
	TotalWeight int
	WeightMap   = make(map[int]int)
	RateMap     = make(map[int]float32)
	Timer       = map[int]int{
		0:  StatusPrepare,     //游戏间隔
		1:  StatusStartStake,  //开始押注
		2:  StatusSendStake,   //发送押注
		28: StatusEndStake,    //停止押注,下发当局结果
		33: StatusShowResult,  //播放抽奖动画，发下每人盈利
		43: StatusShowWinGold, //结算，并发送非金币
		45: StatusPrepare,     //游戏间隔
	}
	CanStakeChipConf = []*CanStakeChip{
		{
			GoldMin: 0,
			GoldMax: 100000,
			Chips: [3]int{100, 500, 1000},
		},
		{
			GoldMin: 100001,
			GoldMax: 1000000,
			Chips: [3]int{1000, 5000, 20000},
		},
		{
			GoldMin: 1000001,
			GoldMax: 10000000,
			Chips: [3]int{10000, 20000, 50000},
		},
		{
			GoldMin: 10000001,
			GoldMax: -1,
			Chips: [3]int{10000, 20000, 50000},
		},
	}
	InvestConfArr = []*InvestConf{
		{
			Id:       1,
			Name:     "女装",
			ParentId: 13,
			Rate:     10,
			Weight:   100,
			Icon:     "tubiao_01",
		},
		{
			Id:       2,
			Name:     "男装",
			ParentId: 13,
			Rate:     10,
			Weight:   100,
			Icon:     "tubiao_02",
		},
		{
			Id:       3,
			Name:     "童装",
			ParentId: 13,
			Rate:     10,
			Weight:   100,
			Icon:     "tubiao_03",
		},
		{
			Id:       4,
			Name:     "农机",
			ParentId: 14,
			Rate:     10,
			Weight:   100,
			Icon:     "tubiao_04",
		},
		{
			Id:       5,
			Name:     "农药",
			ParentId: 14,
			Rate:     10,
			Weight:   100,
			Icon:     "tubiao_05",
		},
		{
			Id:       6,
			Name:     "化肥",
			ParentId: 14,
			Rate:     10,
			Weight:   100,
			Icon:     "tubiao_06",
		},
		{
			Id:       7,
			Name:     "影院",
			ParentId: 15,
			Rate:     10,
			Weight:   100,
			Icon:     "tubiao_07",
		},
		{
			Id:       8,
			Name:     "KTV",
			ParentId: 15,
			Rate:     10,
			Weight:   100,
			Icon:     "tubiao_08",
		},
		{
			Id:       9,
			Name:     "网吧",
			ParentId: 15,
			Rate:     10,
			Weight:   100,
			Icon:     "tubiao_09",
		},
		{
			Id:       10,
			Name:     "中餐",
			ParentId: 16,
			Rate:     10,
			Weight:   100,
			Icon:     "tubiao_10",
		},
		{
			Id:       11,
			Name:     "西餐",
			ParentId: 16,
			Rate:     10,
			Weight:   100,
			Icon:     "tubiao_11",
		},
		{
			Id:       12,
			Name:     "酒吧",
			ParentId: 16,
			Rate:     10,
			Weight:   100,
			Icon:     "tubiao_12",
		},
		{
			Id:       13,
			Name:     "服装",
			ParentId: 17,
			Rate:     3.6,
			Weight:   0,
			Icon:     "tubiao_13",
		},
		{
			Id:       14,
			Name:     "制造",
			ParentId: 17,
			Rate:     3.6,
			Weight:   0,
			Icon:     "tubiao_14",
		},
		{
			Id:       15,
			Name:     "娱乐",
			ParentId: 18,
			Rate:     3.6,
			Weight:   0,
			Icon:     "tubiao_15",
		},
		{
			Id:       16,
			Name:     "餐饮",
			ParentId: 18,
			Rate:     3.6,
			Weight:   0,
			Icon:     "tubiao_16",
		},
		{
			Id:       17,
			Name:     "工业",
			ParentId: 0,
			Rate:     1.86,
			Weight:   0,
			Icon:     "tubiao_17",
		},
		{
			Id:       18,
			Name:     "服务",
			ParentId: 0,
			Rate:     1.86,
			Weight:   0,
			Icon:     "tubiao_18",
		},
		{
			Id:       19,
			Name:     "鸿运奖",
			ParentId: 0,
			Rate:     100,
			Weight:   10,
			Icon:     "tubiao_19",
		},
	}
)
