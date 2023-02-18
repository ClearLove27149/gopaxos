package gopaxos

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
)

type Acceptor struct {
	lis            net.Listener // 监听
	id             int
	minProposal    int         // 接受者承诺的提案编号
	acceptedNumber int         // 接受者接受的提案编号
	acceptedValue  interface{} // 接受者已接受的提案值

	learners []int
}

// Acceptor
func newAcceptor(id int, learners []int) *Acceptor {
	acceptor := &Acceptor{
		id:       id,
		learners: learners,
	}
	acceptor.server()
	return acceptor
}

func (a *Acceptor) server() {
	rpcs := rpc.NewServer()
	rpcs.Register(a)
	addr := fmt.Sprintf(":%d", a.id)
	l, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatal("Listen error:", e)
	}
	a.lis = l
	go func() {
		for {
			conn, err := a.lis.Accept()
			if err != nil {
				continue
			}
			go rpcs.ServeConn(conn)
		}
	}()
}
func (a *Acceptor) close() {
	a.lis.Close()
}

// process
// 1、承诺提案
// 接受者处理提议者第一阶段的请求，判断提议者的提案编号是否比已承诺的天编号大，若是则承诺接受这个提案
// 并且如果已经有了接受过的提案，将其编号和值返回
func (a *Acceptor) Prepare(args *MsgArgs, replay *MsgReplay) error {
	if args.Number > a.minProposal {
		a.minProposal = args.Number
		replay.Number = a.acceptedNumber
		replay.Value = a.acceptedValue
		replay.Ok = true
	} else {
		replay.Ok = false
	}
	return nil
}

// 2、接受提案
// 接受者处理提议者的二阶段请求，提议者把超过半数接受者承诺的提案发过来，接受者判断这个提案是否不小于承诺的提案
// 若是，则接受这个提案，并发给学习者学习
func (a *Acceptor) Accept(args *MsgArgs, replay *MsgReplay) error {
	if args.Number >= a.minProposal {
		a.minProposal = args.Number
		a.acceptedNumber = args.Number
		a.acceptedValue = args.Value
		replay.Ok = true
		// 转发给学习者学习
		for _, lid := range a.learners {
			// 协程
			go func(learner int) {
				addr := fmt.Sprintf("127.0.0.1:%d", learner)
				args.From = a.id
				args.To = learner
				resp := new(MsgReplay)
				ok := call(addr, "Learner.Learn", args, resp)
				if !ok {
					return
				}
			}(lid)
		}
	} else {
		replay.Ok = false
	}
	return nil
}
