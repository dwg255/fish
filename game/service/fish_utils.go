package service

import (
	"github.com/astaxie/beego/logs"
	"math/rand"
	"time"
)

type FishUtil struct {
	//ActiveFish []*Fish

	//Lock sync.Mutex
	CurrentFishId    FishId
	BuildFishChan    chan *Fish
	StopBuildFish    chan bool //暂停出鱼
	RestartBuildFish chan bool //重新开始出鱼
	Exit             chan bool //接收信号
}
type FishId int

//普通鱼
type Fish struct {
	FishKind        int       `json:"fishKind"`
	Trace           [][]int   `json:"trace"`
	Speed           int       `json:"speed"`
	FishId          FishId    `json:"fishId"`
	ActiveTime      time.Time `json:"-"`
	FrontActiveTime int64     `json:"activeTime"` //给客户端的时间
}

//组合鱼
type ArrayFish struct {
	FishKind  int    `json:"fishKind"`
	TraceKind int    `json:"traceKind"`
	FishId    FishId `json:"fishId"`
	Speed     int    `json:"speed"`
}

type FishArrayRet struct {
	FormationKind int            `json:"formationKind"`
	FishArray     [][]*ArrayFish `json:"fishArray"`
	EndTime       time.Time      `json:"-"`
	EndTimeStamp  int64          `json:"endTime"`
}

func (p *FishUtil) GenerateFishId() FishId {
	p.CurrentFishId++
	return p.CurrentFishId
}

func (p *FishUtil) BuildFishTrace() {
	var buildTrace = func() int {
		// 线路随机生成
		var traceId = 101
		//var traceRandom = Math.floor(Math.random() * 1000) + 1
		rand.Seed(time.Now().UnixNano())
		var traceKind = rand.Int()
		randNum := rand.Int()
		switch traceKind%3 + 1 {
		case 1: // 直线 201-217
			traceId = randNum%17 + 201
			break
		case 2: // 二阶曲线 1-10
			traceId = randNum%10 + 1
			break
		case 3: // 三阶曲线 101 -110
			traceId = randNum%10 + 101
			break
		}
		return traceId
	}

	c1 := time.NewTicker(time.Second * 2)
	c2 := time.NewTicker(time.Second*10 + time.Millisecond*100)
	c3 := time.NewTicker(time.Second*30 + time.Millisecond*200)
	c4 := time.NewTicker(time.Second * 61)
	go func() {
		defer func() {
			logs.Trace("exit utils")
		}()
		defer func() {
			c1.Stop()
			c2.Stop()
			c3.Stop()
			c4.Stop()
			close(p.BuildFishChan)
			close(p.StopBuildFish)
			close(p.RestartBuildFish)
		}()
		//logs.Debug("utils start running ...")
		var buildNormalFish = func() {
			rand.Seed(time.Now().UnixNano())
			traceKind := buildTrace()
			fishKind := rand.Intn(15) + 1
			traces := getPathMap(traceKind)

			//logs.Debug("add normal fish tick run")
			for i := 0; i < len(traces); i++ {
				fishId := p.GenerateFishId()
				p.AddFish(fishKind, traces[i], fishId)
			}
		}
		buildNormalFish()
		for {
			//logs.Error("for loop......")
			select {
			case <-c1.C: //随机生成鱼 1-15
				buildNormalFish()
			case <-c2.C: // 16-20
				//logs.Error("<-c2.C in")
				rand.Seed(time.Now().UnixNano())
				fishKind := rand.Intn(5) + 16
				fishId := p.GenerateFishId()
				traceKind := buildTrace()
				traces := getPathMap(traceKind)
				p.AddFish(fishKind, traces[0], fishId)
			case <-c3.C: // 21-34
				//logs.Error("<-c3.C in")
				rand.Seed(time.Now().UnixNano())
				fishKind := rand.Intn(14) + 21
				fishId := p.GenerateFishId()
				traceKind := buildTrace()
				traces := getPathMap(traceKind)
				p.AddFish(fishKind, traces[1], fishId)

			case <-c4.C: // 鱼王
				//logs.Error("<-c4.C in")
				fishKind := 35
				fishId := p.GenerateFishId()
				rand.Seed(time.Now().UnixNano())
				traceKind := rand.Intn(10) + 101
				traces := getPathMap(traceKind)
				p.AddFish(fishKind, traces[1], fishId)
			case <-p.StopBuildFish: //停止出鱼
				logs.Trace("build util StopBuildFish...")
				c1.Stop()
				c2.Stop()
				c3.Stop()
				c4.Stop()
				//logs.Debug("<-p.StopBuildFish")
			case <-p.RestartBuildFish:
				logs.Trace("build util RestartBuildFish...")
				c1 = time.NewTicker(time.Second * 2)
				c2 = time.NewTicker(time.Second*10 + time.Millisecond*100)
				c3 = time.NewTicker(time.Second*30 + time.Millisecond*200)
				c4 = time.NewTicker(time.Second * 61)
				//logs.Debug("<-p.RestartBuildFish")
				//return
			case <-p.Exit:
				//logs.Error("<-p.Exit")
				//退出关闭资源
				//close(p.StopBuildFish)
				//close(p.RestartBuildFish)
				close(p.Exit)
				//logs.Debug("<-p.Exit")
				return
			}
		}
	}()
}

func (p *FishUtil) AddFish(fishKind int, trace [][]int, fishId FishId) {

	var speed = 6
	if fishId >= 35 {
		speed = 3
	} else if fishId >= 30 {
		speed = 4
	} else if fishId >= 20 {
		speed = 5
	}
	p.BuildFishChan <- &Fish{
		FishKind:        fishKind,
		Trace:           trace,
		Speed:           speed,
		FishId:          fishId,
		ActiveTime:      time.Now(),
		FrontActiveTime: time.Now().UnixNano() / 1e6,
	}
}

//启动鱼阵
func BuildFishArray() (ret *FishArrayRet) {
	var fishId FishId
	var generateFishId = func() FishId {
		fishId++
		return 	fishId
	}
	var fishArray = make([][]*ArrayFish, 0)
	var duration = 0
	//直线鱼阵
	var buildFormationLine = func() {
		duration = 60
		fishArray = append(fishArray, make([]*ArrayFish, 0))
		fishArray = append(fishArray, make([]*ArrayFish, 0))
		var kind = 14
		for i := 0; i < 30; i++ {
			kind = i/3 + 10
			fishArray[0] = append(fishArray[0], &ArrayFish{
				FishKind:  kind,
				TraceKind: 0,
				FishId:    generateFishId(),
				Speed:     0,
			})
			fishArray[1] = append(fishArray[1], &ArrayFish{
				FishKind:  kind,
				TraceKind: 0,
				FishId:    generateFishId(),
				Speed:     0,
			})
		}
	}

	//环形鱼阵
	var buildCircleGroupFish = func() {
		duration = 60
		kind, fishNum := 1, 20
		for i := 0; i < 10; i++ {
			kind += 2
			fishArray = append(fishArray, make([]*ArrayFish, 0))
			if i > 20 {
				fishNum = 10
			}
			for j := 0; j < fishNum; j++ {
				fishArray[i] = append(fishArray[i], &ArrayFish{
					FishKind:  kind,
					TraceKind: 0,
					FishId:    generateFishId(),
					Speed:     0,
				})
			}
		}
	}

	// 两个螺旋形数组
	var buildSpiralGroupFish = func() {
		duration = 60
		fishArray = append(fishArray, make([]*ArrayFish, 0))
		fishArray = append(fishArray, make([]*ArrayFish, 0))
		kind := 1
		for i := 1; i <= 30; i++ {
			kind = ((i-1)/10 + 1) * 5
			fishArray[0] = append(fishArray[0],&ArrayFish{
				FishKind:  kind,
				TraceKind: 0,
				FishId:    generateFishId(),
				Speed:     0,
			})
			fishArray[1] = append(fishArray[1],&ArrayFish{
				FishKind:  kind,
				TraceKind: 0,
				FishId:    generateFishId(),
				Speed:     0,
			})
		}
	}
	ret = &FishArrayRet{}
	ret.FormationKind = rand.Intn(3) + 1
	//ret.FormationKind = 1
	switch ret.FormationKind {
	case 1:
		buildFormationLine()
	case 2:
		buildCircleGroupFish()
	case 3:
		buildSpiralGroupFish()
	}
	ret.FishArray = fishArray
	ret.EndTime = time.Now().Add(time.Second * time.Duration(duration))
	ret.EndTimeStamp = ret.EndTime.Unix() * 1e3
	return
}

//是否命中
func IsHit(f *Fish) bool {
	rand.Seed(time.Now().UnixNano())
	// todo 调整概率
	//return rand.Intn(GetFishMulti(f)) == 0
	return rand.Intn(GetFishMulti(f)*3/5) == 0
	//return true
}

func GetFishMulti(fish *Fish) int {
	if multi, ok := FishMulti[fish.FishKind]; ok {
		return multi
	} else {
		return 2
	}
}

// 根据id取得子弹的倍数
func GetBulletMulti(BulletKind int) int {
	if multi, ok := BulletMulti[BulletKind]; ok {
		return multi
	} else {
		return 1
	}
}
