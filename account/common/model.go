package common

type FishUser struct {
	UserId      int    `json:"userid" dB:"UsERID"`
	Account     string `json:"account" DB:"account"`
	Name        string `json:"name" DB:"name"`
	Sex         int    `json:"sex" dB:"sex"`
	HeadImg     string `json:"headimg" DB:"headimg"`
	Lv          int    `json:"lv" DB:"LV"`
	Exp         int    `json:"exP" Db:"exP"`
	Coins       int    `json:"coins" db:"coins"`
	Vip         int    `json:"vip" db:"vip"`
	Money       int    `json:"money" DB:"money"`
	Gems        int    `json:"gems" db:"gems"`
	RoomId      string `json:"roomid" DB:"roomid"`
	History     string `json:"history" dB:"history"`
	Power       int    `json:"power" DB:"power"`
	ReNameCount int    `json:"renamecount" db:"renamecount"`
	ReHeadCount int    `json:"reheadcount" db:"reheadcount"`
	PropId      int    `json:"propid" db:"propid"`
}

type InvestUserStake struct {
	Id           int    `json:"id" db:"id"`
	GameTimesId  string `json:"game_times_id" db:"game_times_id"`
	Periods      int    `json:"record_id" db:"periods"`
	RoomId       int    `json:"room_id" db:"room_id"`
	RoomType     int    `json:"room_type" db:"room_type"`
	UserId       int    `json:"user_id" db:"user_id"`
	Nickname     string `json:"nickname" db:"nickname"`
	UserAllStake int    `json:"stake_gold" db:"user_all_stake"`
	WinGold      int    `json:"get_gold" db:"get_gold"`
	StakeDetail  string `json:"stake_detail" db:"stake_detail"`
	GameResult   int    `json:"game_result" db:"game_result"`
	Pool         int    `json:"game_pool" db:"game_pool"`
	StakeTime    string `json:"stake_time" db:"last_stake_time"`
}
