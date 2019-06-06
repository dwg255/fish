package service

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
)

/*
 // 座位号
 -------------
 0   1   2
 7               3
 6   5   4
 -------------
*/

var (
	SwitchSceneTimer = 60 * 5 // 5分钟切一次场景

	//鱼阵场景
	SceneKind1 = 0
	SceneKind2 = 1
	SceneKind3 = 2
	SceneKind4 = 3
	SceneKind5 = 4
	SceneKind6 = 5
	SceneKind7 = 6
	SceneKind8 = 7

	//鱼的种类
	FishKind1  = 1
	FishKind2  = 2
	FishKind3  = 3
	FishKind4  = 4
	FishKind5  = 5
	FishKind6  = 6
	FishKind7  = 7
	FishKind8  = 8
	FishKind9  = 9
	FishKind10 = 10
	FishKind11 = 11
	FishKind12 = 12
	FishKind13 = 13
	FishKind14 = 14
	FishKind15 = 15
	FishKind16 = 16
	FishKind17 = 17
	FishKind18 = 18
	FishKind19 = 19
	FishKind20 = 20
	FishKind21 = 21
	FishKind22 = 22
	FishKind23 = 23 // 一网打尽
	FishKind24 = 24 // 一网打尽
	FishKind25 = 25 // 一网打尽
	FishKind26 = 26 // 一网打尽
	FishKind27 = 27
	FishKind28 = 28
	FishKind29 = 29
	FishKind30 = 30 // 全屏炸弹
	FishKind31 = 31 // 同类炸弹
	FishKind32 = 32 // 同类炸弹
	FishKind33 = 33 // 同类炸弹
	FishKind34 = 34
	FishKind35 = 35

	FishMulti = map[int]int{
		1:  2,
		2:  2,
		3:  3,
		4:  4,
		5:  5,
		6:  5,
		7:  6,
		8:  7,
		9:  8,
		10: 9,
		11: 10,
		12: 11,
		13: 12,
		14: 18,
		15: 25,
		16: 30,
		17: 35,
		18: 40,
		19: 45,
		20: 50,
		21: 80,
		22: 100,
		23: 45, //45-150, // 一网打尽
		24: 45, //45-150, // 一网打尽
		25: 45, //45-150, // 一网打尽
		26: 45, //45-150, // 一网打尽
		27: 50,
		28: 60,
		29: 70,
		30: 100, // 全屏炸弹
		31: 110, // 同类炸弹
		32: 110, // 同类炸弹
		33: 110, // 同类炸弹
		34: 120,
		35: 200,
	}

	BulletKind = map[string]int{
		"bullet_kind_normal_1": 0,
		"bullet_kind_normal_2": 1,
		"bullet_kind_normal_3": 2,
		"bullet_kind_vip1_1":   3,
		"bullet_kind_vip1_2":   4,
		"bullet_kind_vip1_3":   5,
		"bullet_kind_vip2_1":   6,
		"bullet_kind_vip2_2":   7,
		"bullet_kind_vip2_3":   8,
		"bullet_kind_vip3_1":   9,
		"bullet_kind_vip3_2":   10,
		"bullet_kind_vip3_3":   11,
		"bullet_kind_vip4_1":   12,
		"bullet_kind_vip4_2":   13,
		"bullet_kind_vip4_3":   14,
		"bullet_kind_vip5_1":   15,
		"bullet_kind_vip5_2":   16,
		"bullet_kind_vip5_3":   17,
		"bullet_kind_vip6_1":   19,
		"bullet_kind_vip6_2":   20,
		"bullet_kind_vip6_3":   21,
		"bullet_kind_laser":    22,
	}

	BulletMulti = map[int]int{
		1:  1,
		2:  2,
		3:  3,
		4:  1,
		5:  3,
		6:  5,
		7:  1,
		8:  3,
		9:  5,
		10: 1,
		11: 3,
		12: 5,
		13: 1,
		14: 3,
		15: 5,
		16: 1,
		17: 3,
		18: 5,
		19: 1,
		20: 3,
		21: 5,
		22: 1, // 激光炮
	}
)

const (
	//SUB_S_GAME_CONFIG             = "SUB_S_GAME_CONFIG"
	//SUB_S_FISH_TRACE              = "SUB_S_FISH_TRACE"
	//SUB_S_EXCHANGE_FISHSCORE      = "SUB_S_FISH_TRACE"
	//SUB_S_USER_FIRE               = "SUB_S_FISH_TRACE"
	//SUB_S_CATCH_FISH              = "SUB_S_FISH_TRACE"
	//SUB_S_BULLET_ION_TIMEOUT      = "SUB_S_BULLET_ION_TIMEOUT"
	//SUB_S_LOCK_TIMEOUT            = "SUB_S_LOCK_TIMEOUT"
	//SUB_S_CATCH_SWEEP_FISH        = "SUB_S_CATCH_SWEEP_FISH"
	//SUB_S_CATCH_SWEEP_FISH_RESULT = "SUB_S_CATCH_SWEEP_FISH_RESULT"
	//SUB_S_HIT_FISH_LK             = "SUB_S_HIT_FISH_LK"
	//SUB_S_SWITCH_SCENE            = "SUB_S_SWITCH_SCENE"
	//SUB_S_STOCK_OPERATE_RESULT    = 111 //库存操作
	//SUB_S_SCENE_END               = 112 //场景结束
	//SUB_S_CATCH_FISHRESULT        = 113 //捉鱼结果
	//SUB_S_SETTLE_FISHSCORE        = 114 //解决鱼分数
	//SUB_S_SWIM_SCENE              = 115 //游泳场景
	//SUB_S_SPECIAL_PRICE1          = 116 //特价1
	//SUB_S_ADD_PRICE1_SCORE        = 117 //添加价格分数
	//SUB_S_END_SPECIAL1            = 118 //结束特别
	//SUB_S_SPECIAL_PRICE2          = 119
	//SUB_S_UPDATE_POS              = 120
	//SUB_S_END_SPECIAL2            = 121
	//SUB_S_SPECIAL_PRICE3          = 122
	//SUB_S_END_SPECIAL3            = 123
	//SUB_S_LOCK_FISH               = 124 //锁定鱼
	//SUB_S_BLACK_LIST              = 125 //黑名单
	//SUB_S_WHITE_LIST              = 126 //白名单
	//SUB_S_BIGFISH_LIST            = 127 //大鱼名单
	//SUB_S_LINE_TRACE              = 128 //线追踪
	//SUB_S_SHOAL_TRACE             = 129 //浅追踪

	//基础分值，底分
	GameBaseScore = 1
	//最小携带金币
	MinHaveScore = 1
	//最大携带金币
	MaxHaveScore = 100
	//抽水比例，千分比，5代表千分之5
	TaxRatio = 5
)

var pathMap = make(map[string][][][]int)

func LoadTraceFile(path string) (err error) {
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("file %v not exists", path)
	}
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var jsonStrByte []byte
	for {
		buf := make([]byte, 1024)
		readNum, err := file.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		for i := 0; i < readNum; i++ {
			jsonStrByte = append(jsonStrByte, buf[i])
		}
		if 0 == readNum {
			break
		}
	}
	err = json.Unmarshal(jsonStrByte, &pathMap)
	if err != nil {
		fmt.Printf("json unmarsha1 err:%v \n", err)
		return
	} else {
		fmt.Println("success")
	}
	return
}

func getPathMap(id int) [][][]int {
	return pathMap[strconv.Itoa(id)]
}
