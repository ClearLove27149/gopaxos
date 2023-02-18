package gopaxos

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
)

type Learner struct {
	lis         net.Listener
	id          int
	AcceptedMsg map[int]MsgArgs // 记录接受者已接受的提案，（接受者id，请求消息）
}

// Learner
func newLearner(id int, acceptorIds []int) *Learner {
	learner := &Learner{
		id:          id,
		AcceptedMsg: make(map[int]MsgArgs),
	}
	for _, aid := range acceptorIds {
		learner.AcceptedMsg[aid] = MsgArgs{
			Number: 0,
			Value:  nil,
		}
	}
	learner.server()
	return learner
}
func (l *Learner) server() {
	rpcs := rpc.NewServer()
	rpcs.Register(l)
	addr := fmt.Sprintf("127.0.0.1:%d", l.id)
	lis, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatal("Listen error:", e)
	}
	l.lis = lis
	go func() {
		for {
			conn, err := l.lis.Accept()
			if err != nil {
				continue
			}
			go rpcs.ServeConn(conn)
		}
	}()
}
func (l *Learner) close() {
	l.lis.Close()
}
func (l *Learner) majority() int {
	return len(l.AcceptedMsg)/2 + 1
}

// process
/** 提案每接受一个提案就发给学习者，学习者需要统计哪些提案是超过半数接受者接受的，这个提案就是最后被批准的
**/
// Learn : 将接受者发来的提案加入map， 接受者接受提案后rpc调用
func (l *Learner) Learn(args *MsgArgs, replay *MsgReplay) error {
	a := l.AcceptedMsg[args.From]
	if a.Number < args.Number {
		l.AcceptedMsg[args.From] = *args
		replay.Ok = true
	} else {
		replay.Ok = false
	}
	return nil
}

// 批准提案
func (l *Learner) chosen() interface{} {
	acceptCounts := make(map[int]int)  // 保存接受者的同一个提案数目
	acceptMsg := make(map[int]MsgArgs) // 提案消息

	for _, accpted := range l.AcceptedMsg {
		if accpted.Number != 0 {
			acceptCounts[accpted.Number]++
			acceptMsg[accpted.Number] = accpted
		}
	}
	// 批准超过一半的接受者接受的提案
	for n, count := range acceptCounts {
		if count >= l.majority() {
			return acceptMsg[n].Value
		}
	}
	return nil
}
