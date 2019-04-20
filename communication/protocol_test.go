package communication

import (
  "fmt"
  "testing"

  "github.com/FLAGlab/DCoPN/petrinet"
)

const (
  stepErrMsg = "Should be at %v(%v) but was %v"
  placeErrMsg = "Expected place %v from %v to have %v marks, but had %v"
  priorityErrMsg = "Expected priority %v but was %v"
)

type pnTransitionPicker struct {
  transitionIDToFire int
  addrToFire string
}

func (tp *pnTransitionPicker) pick(options map[string][]*petrinet.Transition) (*petrinet.Transition, string) {
  trs := options[tp.addrToFire]
  for _, tr := range trs {
    if tr.ID == tp.transitionIDToFire {
      return tr, tp.addrToFire
    }
  }
  fmt.Printf("Transition with address %v and id %v wasnt on the options %v", tp.addrToFire, tp.transitionIDToFire, options)
  return nil, ""
}

func (tp *pnTransitionPicker) updatePick(tID int, addr string) {
  tp.transitionIDToFire = tID
  tp.addrToFire = addr
}

func initListen(cm *connectionsMap, leader *myTestPeerNode) func() {
  exclude := make(map[string]bool)
  exclude[leader.address] = true
  startListening(cm, exclude)
  return func() { endConnectionsMap(cm) }
}

func TestLeaderFlowOnlyNode(t *testing.T) {
  pn := simpleTestPetriNet(1, "ctx1")
  pList := []*petrinet.PetriNet{pn}
  cm, leader := setUpTestPetriNodes(pList, pn.ID)
  deferFunc := initListen(cm, leader)
  defer deferFunc()
  picker := pnTransitionPicker{1, leader.address}
  leader.rftNode.pNode.transitionPicker = func(options map[string][]*petrinet.Transition) (*petrinet.Transition, string) {
    return picker.pick(options)
  }
  if leader.rftNode.pNode.step != ASK_STEP {
    t.Errorf(stepErrMsg, "ASK_STEP", ASK_STEP, leader.rftNode.pNode.step)
  }
  leader.rftNode.ask()
  if leader.rftNode.pNode.step != PREPARE_FIRE_STEP {
    t.Errorf(stepErrMsg, "PREPARE_FIRE_STEP", PREPARE_FIRE_STEP, leader.rftNode.pNode.step)
  }
  leader.rftNode.prepareFire()
  if leader.rftNode.pNode.step != FIRE_STEP {
    t.Errorf(stepErrMsg, "FIRE_STEP", FIRE_STEP, leader.rftNode.pNode.step)
  }
  leader.rftNode.fire()
  if leader.rftNode.pNode.step != PRINT_STEP {
    t.Errorf(stepErrMsg, "PRINT_STEP", PRINT_STEP, leader.rftNode.pNode.step)
  }
  leader.rftNode.print()
  if leader.rftNode.pNode.step != ASK_STEP {
    t.Errorf(stepErrMsg, "ASK_STEP", ASK_STEP, leader.rftNode.pNode.step)
  }
}

func TestLeaderFlowMultipleNodesNoRemote(t *testing.T) {
  pList := []*petrinet.PetriNet{
    simpleTestPetriNet(1, "ctx1"),
    simpleTestPetriNet(2, "ctx1"),
    simpleTestPetriNet(3, "ctx1")}
  cm, leader := setUpTestPetriNodes(pList, 1)
  deferFunc := initListen(cm, leader)
  defer deferFunc()
  picker := pnTransitionPicker{1, fmt.Sprintf("addr_%v", 2)}
  leader.rftNode.pNode.transitionPicker = func(options map[string][]*petrinet.Transition) (*petrinet.Transition, string) {
    return picker.pick(options)
  }
  if leader.rftNode.pNode.step != ASK_STEP {
    t.Errorf(stepErrMsg, "ASK_STEP", ASK_STEP, leader.rftNode.pNode.step)
  }
  leader.rftNode.ask()
  if leader.rftNode.pNode.step != RECEIVING_TRANSITIONS_STEP {
    t.Errorf(stepErrMsg, "RECEIVING_TRANSITIONS_STEP", RECEIVING_TRANSITIONS_STEP, leader.rftNode.pNode.step)
  }
  // Expects msg from 2 and 3
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  // done, should go to next step
  if leader.rftNode.pNode.step != PREPARE_FIRE_STEP {
    t.Errorf(stepErrMsg, "PREPARE_FIRE_STEP", PREPARE_FIRE_STEP, leader.rftNode.pNode.step)
  }
  leader.rftNode.prepareFire()
  if leader.rftNode.pNode.step != FIRE_STEP { // skip receiving marks because there is no remote transition
    t.Errorf(stepErrMsg, "FIRE_STEP", FIRE_STEP, leader.rftNode.pNode.step)
  }
  leader.rftNode.fire()
  if leader.rftNode.pNode.step != PRINT_STEP {
    t.Errorf(stepErrMsg, "PRINT_STEP", PRINT_STEP, leader.rftNode.pNode.step)
  }
  leader.rftNode.print()
  if leader.rftNode.pNode.step != ASK_STEP {
    t.Errorf(stepErrMsg, "ASK_STEP", ASK_STEP, leader.rftNode.pNode.step)
  }
}

func TestLeaderFlowMultipleNodesWithRemote(t *testing.T) {
  pn := simpleTestPetriNet(1, "ctx1")
  pList := []*petrinet.PetriNet{
    pn,
    simpleTestPetriNet(2, "ctx1"),
    simpleTestPetriNet(3, "ctx1")}
  pn.AddRemoteTransition(1)
  pn.AddRemoteInArc(1, 1, 1, "ctx1")
  pn.AddRemoteOutArc(1, 2, 1, "ctx1")
  cm, leader := setUpTestPetriNodes(pList, 1)
  deferFunc := initListen(cm, leader)
  defer deferFunc()
  picker := pnTransitionPicker{1, fmt.Sprintf("addr_%v", 1)}
  leader.rftNode.pNode.transitionPicker = func(options map[string][]*petrinet.Transition) (*petrinet.Transition, string) {
    return picker.pick(options)
  }
  if leader.rftNode.pNode.step != ASK_STEP {
    t.Errorf(stepErrMsg, "ASK_STEP", ASK_STEP, leader.rftNode.pNode.step)
  }
  leader.rftNode.ask()
  if leader.rftNode.pNode.step != RECEIVING_TRANSITIONS_STEP {
    t.Errorf(stepErrMsg, "RECEIVING_TRANSITIONS_STEP", RECEIVING_TRANSITIONS_STEP, leader.rftNode.pNode.step)
  }
  // Expects msg from 2 and 3
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  // done, should go to next step
  if leader.rftNode.pNode.step != PREPARE_FIRE_STEP {
    t.Errorf(stepErrMsg, "PREPARE_FIRE_STEP", PREPARE_FIRE_STEP, leader.rftNode.pNode.step)
  }
  leader.rftNode.prepareFire()
  if leader.rftNode.pNode.step != RECEIVING_MARKS_STEP {
    t.Errorf(stepErrMsg, "RECEIVING_MARKS_STEP", RECEIVING_MARKS_STEP, leader.rftNode.pNode.step)
  }
  // Expects msg from 2 and 3
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  // done, should go to next step
  if leader.rftNode.pNode.step != FIRE_STEP {
    t.Errorf(stepErrMsg, "FIRE_STEP", FIRE_STEP, leader.rftNode.pNode.step)
  }
  leader.rftNode.fire()
  if leader.rftNode.pNode.step != PRINT_STEP {
    t.Errorf(stepErrMsg, "PRINT_STEP", PRINT_STEP, leader.rftNode.pNode.step)
  }
  leader.rftNode.print()
  if leader.rftNode.pNode.step != ASK_STEP {
    t.Errorf(stepErrMsg, "ASK_STEP", ASK_STEP, leader.rftNode.pNode.step)
  }
}

func TestLocalTransitionOnLeader(t *testing.T) {
  pn := simpleTestPetriNet(1, "ctx1")
  // p1 -1-> t1 -2-> p2 -2-> t2 -3-> p3
  // p1 : inital = 5
  pList := []*petrinet.PetriNet{pn}
  cm, leader := setUpTestPetriNodes(pList, pn.ID)
  deferFunc := initListen(cm, leader)
  defer deferFunc()
  picker := pnTransitionPicker{1, leader.address}
  leader.rftNode.pNode.transitionPicker = func(options map[string][]*petrinet.Transition) (*petrinet.Transition, string) {
    return picker.pick(options)
  }
  leader.rftNode.ask()
  leader.rftNode.prepareFire()
  leader.rftNode.fire()
  leaderPn := leader.rftNode.pNode.petriNet
  if leaderPn.GetPlace(1).GetMarks() != 4 {
    t.Errorf(placeErrMsg, 1, leader.address, 4, leaderPn.GetPlace(1).GetMarks())
  }
  if leaderPn.GetPlace(2).GetMarks() != 2 {
    t.Errorf(placeErrMsg, 2, leader.address, 2, leaderPn.GetPlace(2).GetMarks())
  }
}

func TestLocalTransitionOnOther(t *testing.T) {
  pn := simpleTestPetriNet(1, "ctx1")
  pn2 := simpleTestPetriNet(2, "ctx1")
  pList := []*petrinet.PetriNet{pn, pn2}
  cm, leader := setUpTestPetriNodes(pList, pn.ID)
  deferFunc := initListen(cm, leader)
  defer deferFunc()
  picker := pnTransitionPicker{1, "addr_2"}
  leader.rftNode.pNode.transitionPicker = func(options map[string][]*petrinet.Transition) (*petrinet.Transition, string) {
    pickedT, addr := picker.pick(options)
    t.Logf("Chosen transition: %v\nChosen addr: %v", pickedT, addr)
    t.Logf("Fire options: %v", options)
    return pickedT, addr
  }
  leader.rftNode.ask()
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.prepareFire()
  leader.rftNode.fire()
  leader.rftNode.print()
  otherPn := cm.nodes["addr_2"].rftNode.pNode.petriNet
  if otherPn.GetPlace(1).GetMarks() != 4 {
    t.Errorf(placeErrMsg, 1, "addr_2", 4, otherPn.GetPlace(1).GetMarks())
  }
  if otherPn.GetPlace(2).GetMarks() != 2 {
    t.Errorf(placeErrMsg, 2, "addr_2", 2, otherPn.GetPlace(2).GetMarks())
  }
}

func TestRemoteTransitionInArcsOnLeader(t *testing.T) {
  // remote arcs in leader and other
  pn := simpleTestPetriNet(1, "ctx1")
  pn2 := simpleTestPetriNet(2, "ctx1")
  pn3 := simpleTestPetriNet(3, "ctx1")
  // p1 -1-> t1 -2-> p2 -2-> t2 -3-> p3
  // p1 : inital = 5
  pn.AddRemoteTransition(1)
  pn.AddRemoteInArc(1, 1, 2, "ctx1")
  pList := []*petrinet.PetriNet{pn, pn2, pn3}
  cm, leader := setUpTestPetriNodes(pList, pn.ID)
  deferFunc := initListen(cm, leader)
  defer deferFunc()
  picker := pnTransitionPicker{1, "addr_1"}
  leader.rftNode.pNode.transitionPicker = func(options map[string][]*petrinet.Transition) (*petrinet.Transition, string) {
    pickedT, addr := picker.pick(options)
    t.Logf("Chosen transition: %v\nChosen addr: %v", pickedT, addr)
    t.Logf("Fire options: %v", options)
    return pickedT, addr
  }
  leader.rftNode.ask()
  // should receive transitions from nodes 2 and 3
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.prepareFire()
  // should receive marks from node 2 and 3
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.fire()
  leader.rftNode.print()
  expectedMarks := make(map[string][]int)
  expectedMarks["addr_1"] = []int{4, 2} //from, to
  expectedMarks["addr_2"] = []int{3, 0}
  expectedMarks["addr_3"] = []int{3, 0}
  for addr, marks := range expectedMarks {
    otherPn := cm.nodes[addr].rftNode.pNode.petriNet
    if otherPn.GetPlace(1).GetMarks() != marks[0] {
      t.Errorf(placeErrMsg, 1, addr, marks[0], otherPn.GetPlace(1).GetMarks())
    }
    if otherPn.GetPlace(2).GetMarks() != marks[1] {
      t.Errorf(placeErrMsg, 2, addr, marks[1], otherPn.GetPlace(1).GetMarks())
    }
  }
}

func TestRemoteTransitionInArcsOnOther(t *testing.T) {
  // remote arcs in leader and other
  pn := simpleTestPetriNet(1, "ctx1")
  pn2 := simpleTestPetriNet(2, "ctx1")
  pn3 := simpleTestPetriNet(3, "ctx1")
  // p1 -1-> t1 -2-> p2 -2-> t2 -3-> p3
  // p1 : inital = 5
  pn2.AddRemoteTransition(1)
  pn2.AddRemoteInArc(1, 1, 2, "ctx1")
  pList := []*petrinet.PetriNet{pn, pn2, pn3}
  cm, leader := setUpTestPetriNodes(pList, pn.ID)
  deferFunc := initListen(cm, leader)
  defer deferFunc()
  picker := pnTransitionPicker{1, "addr_2"}
  leader.rftNode.pNode.transitionPicker = func(options map[string][]*petrinet.Transition) (*petrinet.Transition, string) {
    pickedT, addr := picker.pick(options)
    t.Logf("Chosen transition: %v\nChosen addr: %v", pickedT, addr)
    t.Logf("Fire options: %v", options)
    return pickedT, addr
  }
  leader.rftNode.ask()
  // should receive transitions from nodes 2 and 3
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.prepareFire()
  // should receive marks from node 3 (1 is leader so no need)
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.fire()
  leader.rftNode.print()
  expectedMarks := make(map[string][]int)
  expectedMarks["addr_1"] = []int{3, 0} //from, to
  expectedMarks["addr_2"] = []int{4, 2}
  expectedMarks["addr_3"] = []int{3, 0}
  for addr, marks := range expectedMarks {
    otherPn := cm.nodes[addr].rftNode.pNode.petriNet
    if otherPn.GetPlace(1).GetMarks() != marks[0] {
      t.Errorf(placeErrMsg, 1, addr, marks[0], otherPn.GetPlace(1).GetMarks())
    }
    if otherPn.GetPlace(2).GetMarks() != marks[1] {
      t.Errorf(placeErrMsg, 2, addr, marks[1], otherPn.GetPlace(1).GetMarks())
    }
  }

}

func TestRemoteTransitionOutArcsOnLeader(t *testing.T) {
  // remote arcs in leader and other
  pn := simpleTestPetriNet(1, "ctx1")
  pn2 := simpleTestPetriNet(2, "ctx1")
  pn3 := simpleTestPetriNet(3, "ctx1")
  // p1 -1-> t1 -2-> p2 -2-> t2 -3-> p3
  // p1 : inital = 5
  pn.AddRemoteTransition(1)
  pn.AddRemoteOutArc(1, 2, 1, "ctx1")
  pList := []*petrinet.PetriNet{pn, pn2, pn3}
  cm, leader := setUpTestPetriNodes(pList, pn.ID)
  deferFunc := initListen(cm, leader)
  defer deferFunc()
  picker := pnTransitionPicker{1, "addr_1"}
  leader.rftNode.pNode.transitionPicker = func(options map[string][]*petrinet.Transition) (*petrinet.Transition, string) {
    pickedT, addr := picker.pick(options)
    t.Logf("Chosen transition: %v\nChosen addr: %v", pickedT, addr)
    t.Logf("Fire options: %v", options)
    return pickedT, addr
  }
  leader.rftNode.ask()
  // should receive transitions from nodes 2 and 3
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.prepareFire()
  // no need to wait for marks
  leader.rftNode.fire()
  leader.rftNode.print()
  expectedMarks := make(map[string][]int)
  expectedMarks["addr_1"] = []int{4, 2} //from, to
  expectedMarks["addr_2"] = []int{5, 1}
  expectedMarks["addr_3"] = []int{5, 1}
  for addr, marks := range expectedMarks {
    otherPn := cm.nodes[addr].rftNode.pNode.petriNet
    if otherPn.GetPlace(1).GetMarks() != marks[0] {
      t.Errorf(placeErrMsg, 1, addr, marks[0], otherPn.GetPlace(1).GetMarks())
    }
    if otherPn.GetPlace(2).GetMarks() != marks[1] {
      t.Errorf(placeErrMsg, 2, addr, marks[1], otherPn.GetPlace(2).GetMarks())
    }
  }
}

func TestRemoteTransitionOutArcsOnOther(t *testing.T) {
  // remote arcs in leader and other
  pn := simpleTestPetriNet(1, "ctx1")
  pn2 := simpleTestPetriNet(2, "ctx1")
  pn3 := simpleTestPetriNet(3, "ctx1")
  // p1 -1-> t1 -2-> p2 -2-> t2 -3-> p3
  // p1 : inital = 5
  pn2.AddRemoteTransition(1)
  pn2.AddRemoteOutArc(1, 2, 1, "ctx1")
  pList := []*petrinet.PetriNet{pn, pn2, pn3}
  cm, leader := setUpTestPetriNodes(pList, pn.ID)
  deferFunc := initListen(cm, leader)
  defer deferFunc()
  picker := pnTransitionPicker{1, "addr_2"}
  leader.rftNode.pNode.transitionPicker = func(options map[string][]*petrinet.Transition) (*petrinet.Transition, string) {
    pickedT, addr := picker.pick(options)
    t.Logf("Chosen transition: %v\nChosen addr: %v", pickedT, addr)
    t.Logf("Fire options: %v", options)
    return pickedT, addr
  }
  leader.rftNode.ask()
  // should receive transitions from nodes 2 and 3
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.prepareFire()
  // no need to wait for marks
  leader.rftNode.fire()
  leader.rftNode.print()
  expectedMarks := make(map[string][]int)
  expectedMarks["addr_1"] = []int{5, 1} //from, to
  expectedMarks["addr_2"] = []int{4, 2}
  expectedMarks["addr_3"] = []int{5, 1}
  for addr, marks := range expectedMarks {
    otherPn := cm.nodes[addr].rftNode.pNode.petriNet
    if otherPn.GetPlace(1).GetMarks() != marks[0] {
      t.Errorf(placeErrMsg, 1, addr, marks[0], otherPn.GetPlace(1).GetMarks())
    }
    if otherPn.GetPlace(2).GetMarks() != marks[1] {
      t.Errorf(placeErrMsg, 2, addr, marks[1], otherPn.GetPlace(2).GetMarks())
    }
  }
}

func TestRemoteTransitionInhibitorArcsOnLeader(t *testing.T) {
  // remote arcs in leader and other
  pn := simpleTestPetriNet(1, "ctx1")
  pn2 := simpleTestPetriNet(2, "ctx1")
  pn3 := simpleTestPetriNet(3, "ctx1")
  // p1 -1-> t1 -2-> p2 -2-> t2 -3-> p3
  // p1 : inital = 5
  pn.AddRemoteTransition(1)
  pn.AddRemoteInhibitorArc(1, 1, 2, "ctx1")
  pList := []*petrinet.PetriNet{pn, pn2, pn3}
  cm, leader := setUpTestPetriNodes(pList, pn.ID)
  deferFunc := initListen(cm, leader)
  defer deferFunc()
  picker := pnTransitionPicker{1, "addr_1"}
  leader.rftNode.pNode.transitionPicker = func(options map[string][]*petrinet.Transition) (*petrinet.Transition, string) {
    pickedT, addr := picker.pick(options)
    t.Logf("Chosen transition: %v\nChosen addr: %v", pickedT, addr)
    t.Logf("Fire options: %v", options)
    return pickedT, addr
  }
  leader.rftNode.ask()
  // should receive transitions from nodes 2 and 3
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.prepareFire()
  // should receive marks from node 2 and 3
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  // should realice that it must not fire, so should go to PREPARE_FIRE_STEP
  if leader.rftNode.pNode.step != PREPARE_FIRE_STEP {
    t.Errorf(stepErrMsg, "PREPARE_FIRE_STEP", PREPARE_FIRE_STEP, leader.rftNode.pNode.step)
  }
}

func TestRemoteTransitionInhibitorArcsOnOther(t *testing.T) {
  // remote arcs in leader and other
  pn := simpleTestPetriNet(1, "ctx1")
  pn2 := simpleTestPetriNet(2, "ctx1")
  pn3 := simpleTestPetriNet(3, "ctx1")
  // p1 -1-> t1 -2-> p2 -2-> t2 -3-> p3
  // p1 : inital = 5
  pn2.AddRemoteTransition(1)
  pn2.AddRemoteInhibitorArc(1, 1, 2, "ctx1")
  pList := []*petrinet.PetriNet{pn, pn2, pn3}
  cm, leader := setUpTestPetriNodes(pList, pn.ID)
  deferFunc := initListen(cm, leader)
  defer deferFunc()
  picker := pnTransitionPicker{1, "addr_2"}
  leader.rftNode.pNode.transitionPicker = func(options map[string][]*petrinet.Transition) (*petrinet.Transition, string) {
    pickedT, addr := picker.pick(options)
    t.Logf("Chosen transition: %v\nChosen addr: %v", pickedT, addr)
    t.Logf("Fire options: %v", options)
    return pickedT, addr
  }
  leader.rftNode.ask()
  // should receive transitions from nodes 2 and 3
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.prepareFire()
  // should receive marks from node 3 (not 1, hes the leader)
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  // should realice that it must not fire, so should go to PREPARE_FIRE_STEP
  if leader.rftNode.pNode.step != PREPARE_FIRE_STEP {
    t.Errorf(stepErrMsg, "PREPARE_FIRE_STEP", PREPARE_FIRE_STEP, leader.rftNode.pNode.step)
  }
}

func TestPriorityChangeBecauseOfInArcs(t *testing.T) {
  // remote arcs in leader and other
  pn := simpleTestPetriNet(1, "ctx1")
  pn2 := simpleTestPetriNet(2, "ctx1")
  // p1 -1-> t1 -2-> p2 -2-> t2 -3-> p3
  // p1 : inital = 5
  pn.UpdatePriority(2, 1)
  pn2.UpdatePriority(1, 1)
  pn2.UpdatePriority(2, 1)
  pn.AddRemoteTransition(1)
  pn.AddRemoteInArc(1, 1, 4, "ctx1")
  pList := []*petrinet.PetriNet{pn, pn2}
  cm, leader := setUpTestPetriNodes(pList, pn.ID)
  deferFunc := initListen(cm, leader)
  defer deferFunc()
  picker := pnTransitionPicker{1, "addr_1"}
  leader.rftNode.pNode.transitionPicker = func(options map[string][]*petrinet.Transition) (*petrinet.Transition, string) {
    pickedT, addr := picker.pick(options)
    t.Logf("Chosen transition: %v\nChosen addr: %v", pickedT, addr)
    t.Logf("Fire options: %v", options)
    return pickedT, addr
  }
  t.Log("Will fire transition normally")
  leader.rftNode.ask()
  // should receive transitions from node 2
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.prepareFire()
  // should receive marks from node 2
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.fire()
  leader.rftNode.print()
  t.Log("Fired transition normally")
  // should fire leaving 4 -> 2 -> 0 on pn and 1 -> 0 -> 0 on pn2
  expectedMarks := make(map[string][]int)
  expectedMarks["addr_1"] = []int{4, 2} //from, to
  expectedMarks["addr_2"] = []int{1, 0}
  for addr, marks := range expectedMarks {
    otherPn := cm.nodes[addr].rftNode.pNode.petriNet
    if otherPn.GetPlace(1).GetMarks() != marks[0] {
      t.Errorf(placeErrMsg, 1, addr, marks[0], otherPn.GetPlace(1).GetMarks())
    }
    if otherPn.GetPlace(2).GetMarks() != marks[1] {
      t.Errorf(placeErrMsg, 2, addr, marks[1], otherPn.GetPlace(2).GetMarks())
    }
  }
  t.Log("Will try to fire with transition 0")
  fmt.Println("TEST: WILL FIRE WITH TRANSITION 0")
  leader.rftNode.ask()
  // should receive transitions from node 2
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.prepareFire()
  // should receive marks from node 2
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  // should realice that it must not fire, so should go to PREPARE_FIRE_STEP
  if leader.rftNode.pNode.step != PREPARE_FIRE_STEP {
    t.Errorf(stepErrMsg, "PREPARE_FIRE_STEP", PREPARE_FIRE_STEP, leader.rftNode.pNode.step)
  }
  leader.rftNode.prepareFire()
  t.Log("Done trying to fire with transition 0")
  fmt.Println("TEST: DONE FIRE WITH TRANSITION 0 -> SHOULD INCREASE PRIORITY")
  // should realice that there are no transitions of priority 0 to pick, so it should
  // increase used priority
  if leader.rftNode.pNode.step != ASK_STEP {
    t.Errorf(stepErrMsg, "ASK_STEP", ASK_STEP, leader.rftNode.pNode.step)
  }
  if leader.rftNode.pNode.priorityToAsk != 1 {
    t.Errorf(priorityErrMsg, 1, leader.rftNode.pNode.priorityToAsk)
  }
  picker.updatePick(1, "addr_2")
  t.Log("Will fire with transition 1")
  fmt.Println("TEST: WILL FIRE WITH TRANSITION 1")
  leader.rftNode.ask()
  // should receive transitions from node 2
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.prepareFire()
  if leader.rftNode.pNode.step != FIRE_STEP {
    t.Errorf(stepErrMsg, "FIRE_STEP", FIRE_STEP, leader.rftNode.pNode.step)
  }
  // no marks to receive
  leader.rftNode.fire()
  leader.rftNode.print()
  t.Log("Fired with transition 1")
  expectedMarks["addr_1"] = []int{4, 2} //from, to
  expectedMarks["addr_2"] = []int{0, 2}
  for addr, marks := range expectedMarks {
    otherPn := cm.nodes[addr].rftNode.pNode.petriNet
    if otherPn.GetPlace(1).GetMarks() != marks[0] {
      t.Errorf(placeErrMsg, 1, addr, marks[0], otherPn.GetPlace(1).GetMarks())
    }
    if otherPn.GetPlace(2).GetMarks() != marks[1] {
      t.Errorf(placeErrMsg, 2, addr, marks[1], otherPn.GetPlace(2).GetMarks())
    }
  }
}

// TODO: tests for priority, test for experiment which is not working correctly
func TestPriorityChangeBecauseOfInhibitorArcs(t *testing.T) {
  // remote arcs in leader and other
  pn := experiment1TestPetriNet(1, "ctx1")
  pn2 := experiment1TestPetriNet(2, "ctx1")
  pList := []*petrinet.PetriNet{pn, pn2}
  cm, leader := setUpTestPetriNodes(pList, pn.ID)
  deferFunc := initListen(cm, leader)
  defer deferFunc()
  picker := pnTransitionPicker{2, "addr_1"}
  leader.rftNode.pNode.transitionPicker = func(options map[string][]*petrinet.Transition) (*petrinet.Transition, string) {
    pickedT, addr := picker.pick(options)
    fmt.Printf("Chosen transition: %v\nChosen addr: %v\n", pickedT, addr)
    fmt.Printf("Fire options: %v\n", options)
    return pickedT, addr
  }
  fmt.Println("Will fire transition normally")
  leader.rftNode.ask()
  // should receive transitions from node 2
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.prepareFire()
  // should receive marks from node 2
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.fire()
  leader.rftNode.print()
  fmt.Println("Fired transition normally")
  // should fire leaving 0, 1, 0 in pn and 1, 0, 0 in pn2
  expectedMarks := make(map[string][]int)
  expectedMarks["addr_1"] = []int{0, 1} //from, to
  expectedMarks["addr_2"] = []int{1, 0}
  for addr, marks := range expectedMarks {
    otherPn := cm.nodes[addr].rftNode.pNode.petriNet
    if otherPn.GetPlace(1).GetMarks() != marks[0] {
      t.Errorf(placeErrMsg, 1, addr, marks[0], otherPn.GetPlace(1).GetMarks())
    }
    if otherPn.GetPlace(2).GetMarks() != marks[1] {
      t.Errorf(placeErrMsg, 2, addr, marks[1], otherPn.GetPlace(2).GetMarks())
    }
  }
  picker.updatePick(2, "addr_2")
  fmt.Println("Will try to fire with transition 0")
  leader.rftNode.ask()
  // should receive transitions from node 2
  fmt.Println("Will receive msg from node 2")
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  fmt.Println("Received msg from node 2")
  fmt.Println("Will prepare fire")
  leader.rftNode.prepareFire()
  fmt.Println("Done prepare fire -> asdf here")
  // should receive marks from node 1 (aka, will try to fire and realice it cant)
  leader.rftNode.fire()
  // should realice that it must not fire, so should go to PREPARE_FIRE_STEP
  if leader.rftNode.pNode.step != PREPARE_FIRE_STEP {
    t.Errorf(stepErrMsg, "PREPARE_FIRE_STEP", PREPARE_FIRE_STEP, leader.rftNode.pNode.step)
  }
  fmt.Println("Will prepare fire again")
  leader.rftNode.prepareFire()
  fmt.Println("Done prepare fire 2 so should go to ask")
  fmt.Println("Done trying to fire with transition 0")
  // should realice that there are no transitions of priority 0 to pick, so it should
  // increase used priority
  if leader.rftNode.pNode.step != ASK_STEP {
    t.Errorf(stepErrMsg, "ASK_STEP", ASK_STEP, leader.rftNode.pNode.step)
  }
  if leader.rftNode.pNode.priorityToAsk != 1 {
    t.Errorf(priorityErrMsg, 1, leader.rftNode.pNode.priorityToAsk)
  }
  picker.updatePick(3, "addr_2")
  fmt.Println("Will fire with transition 1")
  leader.rftNode.ask()
  // should receive transitions from node 2
  leader.rftNode.processLeader(<- leader.rftNode.pMsg)
  leader.rftNode.prepareFire()
  // no marks to receive
  leader.rftNode.fire()
  leader.rftNode.print()
  fmt.Println("Fired with transition 1")
  expectedMarks["addr_1"] = []int{0} //from, to
  expectedMarks["addr_2"] = []int{1}
  for addr, marks := range expectedMarks {
    otherPn := cm.nodes[addr].rftNode.pNode.petriNet
    if otherPn.GetPlace(3).GetMarks() != marks[0] {
      t.Errorf(placeErrMsg, 1, addr, marks[0], otherPn.GetPlace(1).GetMarks())
    }
  }
}
