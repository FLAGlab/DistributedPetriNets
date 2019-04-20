package communication

import (
  "fmt"
  "testing"

  "github.com/FLAGlab/DCoPN/petrinet"
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
  return nil, ""
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
  stepErrMsg := "Should be at %v(%v) but was %v"
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
  stepErrMsg := "Should be at %v(%v) but was %v"
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
  stepErrMsg := "Should be at %v(%v) but was %v"
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
  placeErr := "Expected place %v to have %v marks, but had %v"
  if leaderPn.GetPlace(1).GetMarks() != 4 {
    t.Errorf(placeErr, 1, 4, leaderPn.GetPlace(1).GetMarks())
  }
  if leaderPn.GetPlace(2).GetMarks() != 2 {
    t.Errorf(placeErr, 2, 2, leaderPn.GetPlace(2).GetMarks())
  }
}

func TestLocalTransitionOnOther(t *testing.T) {

}

func TestRemoteTransitionInArcsOnLeader(t *testing.T) {

}

func TestRemoteTransitionInArcsOnOther(t *testing.T) {

}

func TestRemoteTransitionOutArcsOnLeader(t *testing.T) {

}

func TestRemoteTransitionOutArcsOnOther(t *testing.T) {

}

func TestRemoteTransitionInhibitorArcsOnLeader(t *testing.T) {

}

func TestRemoteTransitionInhibitorArcsOnOther(t *testing.T) {

}
