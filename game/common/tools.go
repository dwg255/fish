package common

import "github.com/astaxie/beego/logs"

//优化，保存父级，不用每次都计算
func (p *InvestConf) GetParents() (parentArr []*InvestConf) {
	return p.ParentArr
}

func (p *InvestConf) InitParents() (parentArr []*InvestConf ,ok bool)  {
	parentArr = append(parentArr, p)
	for parentInvestConf, hasFather := p.GetFather(); hasFather; {
		parentArr = append(parentArr, parentInvestConf)
		parentInvestConf, hasFather = parentInvestConf.GetFather()
	}
	p.ParentArr = parentArr
	if len(p.ParentArr) > 0 {
		ok = true
	}
	return
}

func (p *InvestConf) GetFather() (parentInvestConf *InvestConf, hasFather bool) {
	if p.ParentId == 0 {
		return
	}
	for _, investConf := range InvestConfArr {
		if investConf.Id == p.ParentId {
			parentInvestConf, hasFather = investConf, true
			return
		}
	}
	return
}

func (p *InvestConf) GetNeedGold(StakeMap map[int]int) (needGold int) {
	winPositionArr := p.GetParents()
	for _,investConf := range winPositionArr{
		stakeGold,ok := StakeMap[investConf.Id]
		if !ok {
			logs.Error("user total stake in [%d] is nil",investConf.Id)
			continue
		}
		needGold = needGold + int(float32(stakeGold) * investConf.Rate)
	}
	return
}