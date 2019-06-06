package tools

import (
	"errors"
	"sync"
	"time"
)

const (
	nodeBits  uint8 = 10
	stepBits  uint8 = 12
	nodeMax   int64 = -1 ^ (-1 << nodeBits)
	stepMax   int64 = -1 ^ (-1 << stepBits)
	timeShift uint8 = nodeBits + stepBits
	nodeShift uint8 = stepBits
)

var (
	Epoch int64 = 1546272000000
)

type Node struct {
	mu           sync.Mutex
	timestamp    int64
	node         int64
	step         int64
	GenerateChan chan int64
	stop         chan bool
}

func GenerateUid(nodeNum int64) (err error, generateChan <-chan int64) {
	var node *Node
	if node, err = NewNode(nodeNum);err == nil {
		generateChan = node.GenerateChan
		go node.run()
		return
	}

	return
}

func (n *Node) run() {
	for {
		n.GenerateChan <- n.Generate()
	}
}

func NewNode(nodeNum int64) (*Node, error) {
	if nodeNum < 0 || nodeNum > nodeMax {
		return nil, errors.New("Node number must be between 0 and 1023")
	}
	return &Node{
		timestamp:    0,
		node:         nodeNum,
		step:         0,
		GenerateChan: make(chan int64),
		stop:         make(chan bool),
	}, nil
}

func (n *Node) Generate() int64 {
	n.mu.Lock()
	defer n.mu.Unlock()
	now := time.Now().UnixNano() / 1e6
	if n.timestamp == now {
		n.step++
		if n.step > stepMax {
			for now <= n.timestamp {
				now = time.Now().UnixNano() / 1e6
			}
			n.step = 0
		}
	} else {
		n.timestamp = now
		n.step = 0
	}
	return (now-Epoch)<<timeShift | n.node<<nodeShift | n.step
}
