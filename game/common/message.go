package common

import "sync"

type Message struct {
	Type  int
	Timer int
}

type GameInfo struct {
	StakeMap   map[int]int
	Lock       sync.RWMutex
	GameResult *InvestConf
	Pool       int //奖池
	Periods    int
}

type Record struct {
	ResultChan           chan *InvestConf
	LasTTenTimesRank     [10]*GameRank
	ContinuedWinRecord   map[*InvestConf]int //连续赢的位置
	LastTimeWinRecord    map[int]int //距离上一次赢的局数
	UnfortunatelyRecord  *UnfortunatelyPosition
	EverydayResultRecord *EveryDayRecord
	Lock                 sync.RWMutex
}

type GameRank struct {
	Periods    string      `json:"periods"`
	GameResult *InvestConf `json:"game_result"`
}

//最不幸的位置和连续不出记录
type UnfortunatelyPosition struct {
	Position int
	Count    int
}

type EveryDayRecord struct {
	Date   string
	Record map[*InvestConf]int
}
