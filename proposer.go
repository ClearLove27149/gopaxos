package gopaxos

import (
	"fmt"
)

type Proposer struct {
	id        int   // 服务器id
	round     int   // 提议者已知的最大轮次
	number    int   // 提案编号（轮次，服务器id）
	acceptors []int // 接受者id列表
}

func (p *Proposer) proposalNumber() int {
	return p.round<<16 | p.id
}
func (p *Proposer) majority() int {
	return len(p.acceptors)/2 + 1
}

func (p *Proposer) propose(v interface{}) interface{} {
	p.round++
	p.number = p.proposalNumber()
	// 第一阶段，广播从客户端接受的提案
	prepareCount := 0
	maxNumber := 0

	for _, aid := range p.acceptors {
		args := MsgArgs{
			Number: p.number,
			From:   p.id,
			To:     aid,
		}
		replay := new(MsgReplay)
		err := call(fmt.Sprintf("127.0.0.1:%d", aid), "Acceptor.Prepare", args, replay)
		if !err {
			continue
		}

		if replay.Ok {
			prepareCount++
			if replay.Number > maxNumber {
				maxNumber = replay.Number
				v = replay.Value
			}
		}
		if prepareCount == p.majority() {
			break
		}

	}
	// 第二阶段，发送超过半数接受者承诺的提案
	acceptCount := 0
	if prepareCount >= p.majority() {
		for _, aid := range p.acceptors {
			args := MsgArgs{
				Number: p.number,
				From:   p.id,
				To:     aid,
				Value:  v,
			}
			replay := new(MsgReplay)
			ok := call(fmt.Sprintf("127.0.0.1:%d", aid), "Acceptor.Accept", args, replay)
			if !ok {
				continue
			}
			if replay.Ok {
				acceptCount++
			}
		}
	}
	if acceptCount >= p.majority() {
		// 选择提案的值
		return v
	}
	return nil
}
