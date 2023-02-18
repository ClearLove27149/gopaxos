package gopaxos

import (
	"log"
	"net/rpc"
)

type MsgArgs struct {
	Number int
	Value  interface{}
	From   int
	To     int
}

type MsgReplay struct {
	Ok     bool
	Number int
	Value  interface{}
}

func call(srv string, name string, args interface{}, replay interface{}) bool {
	c, err := rpc.Dial("tcp", srv)
	if err != nil {
		return false
	}
	defer c.Close()
	log.SetFlags(log.Ldate | log.Ltime)
	log.Print(" Call ", srv, " ", name, " ", args)
	err = c.Call(name, args, replay)
	if err == nil {
		return true
	}
	return false
}
