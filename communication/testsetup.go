package communication

import (
  "errors"
  "fmt"

  "github.com/FLAGlab/DCoPN/petrinet"
)

type myTestPeerNode struct {
  address string
	rftNode *RaftNode
  shouldFail bool
}

func (cpeer myTestPeerNode) SendMessage(pMsg petriMessage) error {
  if cpeer.shouldFail {
    return errors.New("Test error")
  }
  cpeer.rftNode.pMsg <- pMsg
	return nil
}

type connectionsMap struct {
  nodes map[string]*myTestPeerNode
}

func (cm *connectionsMap) removePeer(addr string) {
  cm.nodes[addr].rftNode.close()
  delete(cm.nodes, addr)
}

type myTestCommunicationNode struct {
	connections *connectionsMap
  self *myTestPeerNode
}

func (cn *myTestCommunicationNode) CountPeers() int {
	return len(cn.connections.nodes) - 1
}

func (cn *myTestCommunicationNode) Broadcast(pMsg petriMessage) []error {
  var errs []error
  for _, peer := range cn.connections.nodes {
    if peer.address != cn.ExternalAddress() {
      errs = append(errs, peer.SendMessage(pMsg))
    }
  }
	return errs
}

func (cn *myTestCommunicationNode) ExternalAddress() string {
	return cn.self.address
}

func (cn *myTestCommunicationNode) Dial(address string) (PeerNode, error) {
	if address == cn.self.address {
    return nil, fmt.Errorf("Couldn't find %v (actually it's self)", address)
  }
  peer, ok := cn.connections.nodes[address]
  if ok {
    return peer, nil
  }
	return nil, errors.New("Address doesn't exist")
}

func simpleTestPetriNet(id int, ctx string) *petrinet.PetriNet {
  pn := petrinet.Init(id, ctx)
  // p1 -1-> t1 -2-> p2 -2-> t2 -3-> p3
  // p1 : inital = 5
  pn.AddPlace(1, 5, "")
  pn.AddPlace(2, 0, "")
  pn.AddPlace(3, 0, "")
  pn.AddTransition(1, 0)
  pn.AddTransition(2, 0)
  pn.AddInArc(1, 1, 1)
  pn.AddInArc(2, 2, 2)
  pn.AddOutArc(1, 2, 2)
  pn.AddOutArc(2, 3, 3)
  return pn
}

func experiment1TestPetriNet(id int, ctx string) *petrinet.PetriNet {
  p := petrinet.Init(id, ctx)

  p.AddPlace(1,1,"")
  p.AddPlace(2,0,"")
  p.AddPlace(3,0,"")
  p.AddTransition(1, 1)
  p.AddTransition(3, 1)
  p.AddTransition(2, 0)
  p.AddTransition(4, 0)
  p.AddRemoteTransition(2)
  p.AddRemoteInhibitorArc(2, 2, 1, ctx)
  p.AddInArc(1, 2, 1)
  p.AddInArc(3, 4, 1)
  p.AddInArc(2, 4, 1)
  p.AddOutArc(1, 1, 1)
  p.AddOutArc(2, 2, 1)
  p.AddOutArc(3, 3, 1)
  return p
}

func endConnectionsMap(m *connectionsMap) {
  for _, peer := range m.nodes {
    peer.rftNode.close()
  }
}

func startListening(m *connectionsMap, exclude map[string]bool) {
  for addr, peer := range m.nodes {
    if !exclude[addr] {
      go peer.rftNode.Listen()
    }
  }
}

func setUpTestPetriNodes(pnets []*petrinet.PetriNet, leaderId int) (*connectionsMap, *myTestPeerNode) {
  connections := connectionsMap{make(map[string]*myTestPeerNode)}
  var leaderPeer *myTestPeerNode
  for _, pnet := range pnets {
    addr := fmt.Sprintf("addr_%v", pnet.ID)
    myTestComm := &myTestCommunicationNode{connections: &connections}
    testPeer := &myTestPeerNode{addr, nil, false}
    myTestComm.self = testPeer
    pnNode := InitPetriNode(myTestComm, pnet)
    rn := InitRaftNode(pnNode, leaderId == pnet.ID)
    testPeer.rftNode = rn
    connections.nodes[addr] = testPeer
    if leaderId == pnet.ID {
      leaderPeer = testPeer
    }
  }
  return &connections, leaderPeer
}

func (cm *connectionsMap) setUniversalPetriNet(universal *petrinet.PetriNet) {
  for _, node := range cm.nodes {
    node.rftNode.pNode.SetUniversalPetriNet(universal)
  }
}
