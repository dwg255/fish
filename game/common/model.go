package common

type InvestBase struct {
	Id          int    `json:"id"`
	GameTimesId string `json:"game_times_id"`
	Periods     int    `json:"periods"`
	Pool        int    `json:"game_pool"`
	GameResult  int    `json:"game_result"`
	StakeDetail string `json:"stake_detail"`
	StartTime   string `json:"start_time"`
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
