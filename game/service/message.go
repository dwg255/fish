package service

type Message struct {
	Act                     string                    `json:"act"`
	State                   string                    `json:"state"`
	//WinResult               int                       `json:"win_result"`
	WinPosition             []int                     `json:"win_position"`
	Pool                    int                       `json:"pool"`
	MaxContinueLostPosition int                       `json:"max_continue_lost_position"`
	MaxContinueLostCount    int                       `json:"max_continue_lost_count"`
	Timer                   int                       `json:"timer"`
	Periods                 int                       `json:"periods"`
	NewStake                map[string]map[string]int `json:"new_stake"`
	StakeInfo               []map[string]int          `json:"stake_info"`
	UserWinResult           []*UserWinResult          `json:"win_result"`
}

type UserWinResult struct {
	UserId  UserId `json:"user_id"`
	WinGold int `json:"win_gold"`
}
