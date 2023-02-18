package gopaxos

import "testing"

func start(acceptorIds []int, learnerIds []int) ([]*Acceptor, []*Learner) {
	acceptors := make([]*Acceptor, 0)
	for _, aid := range acceptorIds {
		a := newAcceptor(aid, learnerIds)
		acceptors = append(acceptors, a)
	}
	learners := make([]*Learner, 0)
	for _, lid := range learnerIds {
		l := newLearner(lid, acceptorIds)
		learners = append(learners, l)
	}
	return acceptors, learners
}

func end(acceptors []*Acceptor, learners []*Learner) {
	for _, ac := range acceptors {
		ac.close()
	}
	for _, le := range learners {
		le.close()
	}
}

// testing
func TestSingleProposer(t *testing.T) {
	acceptorIds := []int{8001, 8002, 8003}
	learnerIds := []int{9001}
	acceptors, learners := start(acceptorIds, learnerIds)

	defer end(acceptors, learners)
	// 提议者
	p := &Proposer{
		id:        1,
		acceptors: acceptorIds,
	}
	pValue := "hello, world"
	value := p.propose(pValue)
	if value != pValue {
		t.Errorf("value=%s, expect value=%s", value, pValue)
	}
	learnValue := learners[0].chosen()
	if learnValue != pValue {
		t.Errorf("learnValue=%s, expect value=%s", learnValue, pValue)
	}

}
