package petrinet

import (
  "reflect"
  "testing"
)

func TestGetMaxPriority(t *testing.T) {
  errorMsg := "Expected max priority %v but was %v"
  pn := Init(1, "ctx1")
  pn.AddTransition(1, 2)
  if pn.GetMaxPriority() != 2 {
    t.Errorf(errorMsg, 2, pn.GetMaxPriority())
  }
  pn.AddTransition(2, 5)
  pn.AddTransition(2, 3)
  if pn.GetMaxPriority() != 5 {
    t.Errorf(errorMsg, 5, pn.GetMaxPriority())
  }
}

func TestFireTransitionByID(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddTransition(3, 2)
  pn.AddPlace(1, 1, "")
  pn.AddPlace(2, 2, "")
  pn.AddPlace(3, 5, "")
  pn.AddPlace(4, 0, "")
  pn.AddPlace(5, 1, "")
  pn.AddPlace(6, 0, "")
  pn.AddInArc(1, 3, 1)
  pn.AddInArc(2, 3, 2)
  pn.AddInArc(3, 3, 3)
  pn.AddOutArc(3, 4, 3)
  pn.AddOutArc(3, 5, 2)
  pn.AddOutArc(3, 6, 3)
  pn.FireTransitionByID(3)
  expected := make(map[int]int)
  expected[1] = 0
  expected[2] = 0
  expected[3] = 2
  expected[4] = 3
  expected[5] = 3
  expected[6] = 3
  for key, value := range expected {
    if pn.places[key].marks != value {
      t.Errorf(
        "Place %v should have been affected by transition %v, expected it to have %v marks but had %v",
        pn.places[key], pn.transitions[3], value, pn.places[key].marks)
    }
  }
}

func TestCopyPlaceMarksToRemoteArc(t *testing.T) {
  pn := Init(1, "ctx1")
  expectedMarks := []int{5, 3, 1, 0}
  pn.AddPlace(1, expectedMarks[0], "")
  pn.AddPlace(2, expectedMarks[1], "")
  pn.AddPlace(3, expectedMarks[2], "")
  pn.AddPlace(4, expectedMarks[3], "")
  rmtArc := []*RemoteArc{
    {1, "", "", 1, 0},
    {2, "", "", 1, 0},
    {3, "", "", 1, 0},
    {4, "", "", 1, 0}}
  pn.CopyPlaceMarksToRemoteArc(rmtArc)
  for index, rmt := range rmtArc {
    if rmt.Marks != expectedMarks[index] {
      t.Errorf("Expected rmt arc %v to have %v marks", rmt, expectedMarks[index])
    }
  }
}

func TestAddMarksToPlaces(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddPlace(1,0,"")
  pn.AddPlace(2,5,"")
  pn.AddPlace(3,2,"")
  rmtArcsAdd := []*RemoteArc{
    {1, "ctx1", "", 2, 0},
    {2, "ctx1", "", 3, 0},
    {3, "ctx1", "", 2, 0}}
  expectedMarks := []int{2, 8, 4}
  pn.AddMarksToPlaces(ADDITION, rmtArcsAdd, false)
  for index, value := range expectedMarks {
    if pn.places[index + 1].marks != value {
      t.Errorf("Expected place %v to have %v marks", pn.places[index + 1], value)
    }
  }
  pn.places = make(map[int]*Place)
  pn.AddPlace(1,1,"")
  pn.AddPlace(2,5,"")
  pn.AddPlace(3,2,"")
  rmtArcsAdd = []*RemoteArc{
    {1, "ctx1", "", 1, 0},
    {2, "ctx1", "", 3, 0},
    {3, "ctx1", "", 2, 0}}
  expectedMarks = []int{0, 2, 0}
  pn.AddMarksToPlaces(SUBSTRACTION, rmtArcsAdd, false)
  for index, value := range expectedMarks {
    if pn.places[index + 1].marks != value {
      t.Errorf("Expected place %v to have %v marks", pn.places[index + 1], value)
    }
  }
}

func TestGetTransitionOptionsByPriority(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddTransition(1, 0)
  pn.AddTransition(2, 0)
  pn.AddTransition(3, 1)
  pn.AddTransition(4, 0)
  pn.AddTransition(5, 2)
  pn.AddTransition(6, 2)
  pn.AddTransition(7, 2)
  tList, rmtList := pn.GetTransitionOptionsByPriority(0)

  if len(rmtList) != 0 {
    t.Errorf("There should not me any remote transitions on the option: %v", rmtList)
  }
  if len(tList) != 3 {
    t.Errorf("Transitions 1, 2 and 4 should be on the list: %v", tList)
  }
  expectedFuncCheck := func (list []*Transition, expected map[int]bool) {
    for _, tr := range list {
      _, ok := expected[tr.ID]
      if !ok {
        t.Errorf("Got transition %v that wasn't expected", tr)
      }
      delete(expected, tr.ID)
    }
    if len(expected) > 0 {
      t.Errorf("Expected to find %v but didnt", expected)
    }
  }
  expectedIds := make(map[int]bool)
  expectedIds[1] = true
  expectedIds[2] = true
  expectedIds[4] = true
  expectedFuncCheck(tList, expectedIds)

  tList, rmtList = pn.GetTransitionOptionsByPriority(1)
  expectedIds[3] = true
  expectedFuncCheck(tList, expectedIds)
  // Add remote arc for transition 5 and 6
  pn.AddRemoteTransition(5)
  pn.AddRemoteTransition(6)
  pn.AddRemoteInArc(1, 5, 1, "testCtx")
  pn.AddRemoteInArc(1, 6, 1, "testCtx")
  tList, rmtList = pn.GetTransitionOptionsByPriority(2)
  expectedIds[5] = true
  expectedIds[6] = true
  expectedIds[7] = true
  expectedFuncCheck(tList, expectedIds)
  expectedIds[5] = true
  expectedIds[6] = true
  if len(rmtList) != 2 {
    t.Errorf("Remote arcs list should have 2 remote arcs but had %v", rmtList)
  } else {
    _, ok := rmtList[5]
    if !ok {
      t.Error("Remote transition 5 should exist")
    }
    _, ok = rmtList[6]
    if !ok {
      t.Error("Remote transition 6 should exist")
    }
  }
}

func TestGetTransitionOptions(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddTransition(1, 0)
  pn.AddTransition(2, 0)
  pn.AddTransition(3, 1)
  pn.AddTransition(4, 0)
  pn.AddTransition(5, 2)
  pn.AddTransition(6, 2)
  pn.AddTransition(7, 2)
  pn.AddPlace(4, 0, "")
  pn.AddInArc(4, 4, 2)

  tList, rmtList := pn.GetTransitionOptions()
  if len(rmtList) != 0 {
    t.Errorf("There should not me any remote transitions on the options: %v", rmtList)
  }
  if len(tList) != 2 {
    t.Errorf("Transitions 1, and 2 should be on the list: %v", tList)
  }
  expectedFuncCheck := func (list []*Transition, expected map[int]bool) {
    for _, tr := range list {
      _, ok := expected[tr.ID]
      if !ok {
        t.Errorf("Got transition %v that wasn't expected", tr)
      }
      delete(expected, tr.ID)
    }
    if len(expected) > 0 {
      t.Errorf("Expected to find %v but didnt", expected)
    }
  }
  expectedIds := make(map[int]bool)
  expectedIds[1] = true
  expectedIds[2] = true
  expectedFuncCheck(tList, expectedIds)
  // make all of priority 0 not able to fire
  pn.AddPlace(1, 0, "")
  pn.AddPlace(2, 0, "")
  pn.AddInArc(1, 1, 2)
  pn.AddInArc(2, 2, 2)
  tList, rmtList = pn.GetTransitionOptions()
  expectedIds[3] = true
  expectedFuncCheck(tList, expectedIds)
  // make all of priority 1 not able to fire and add remote arc for transition 5 and 6
  pn.AddInArc(1, 3, 2)
  pn.AddRemoteTransition(5)
  pn.AddRemoteTransition(6)
  pn.AddRemoteInArc(1, 5, 1, "testCtx")
  pn.AddRemoteInArc(1, 6, 1, "testCtx")
  tList, rmtList = pn.GetTransitionOptions()
  expectedIds[5] = true
  expectedIds[6] = true
  expectedIds[7] = true
  expectedFuncCheck(tList, expectedIds)
  expectedIds[5] = true
  expectedIds[6] = true
  if len(rmtList) != 2 {
    t.Errorf("Remote arcs list should have 2 remote arcs but had %v", rmtList)
  } else {
    _, ok := rmtList[5]
    if !ok {
      t.Error("Remote transition 5 should exist")
    }
    _, ok = rmtList[6]
    if !ok {
      t.Error("Remote transition 6 should exist")
    }
  }
}

func TestGetRemoteTransitionsFromTransitions(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddTransition(5, 2)
  pn.AddTransition(6, 2)
  pn.AddRemoteTransition(5)
  pn.AddRemoteTransition(6)
  pn.AddRemoteInArc(1, 5, 1, "testCtx")
  pn.AddRemoteInArc(1, 6, 1, "testCtx")
  rmtList := pn.getRemoteTransitionsFromTransitions([]*Transition{pn.transitions[5], pn.transitions[6]})
  expectedIds := make(map[int]bool)
  expectedIds[5] = true
  expectedIds[6] = true
  if len(rmtList) != 2 {
    t.Errorf("Remote arcs list should have 2 remote arcs but had %v", rmtList)
  } else {
    _, ok := rmtList[5]
    if !ok {
      t.Error("Remote transition 5 should exist")
    }
    _, ok = rmtList[6]
    if !ok {
      t.Error("Remote transition 6 should exist")
    }
  }
}

func TestAddPlace(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddPlace(1, 2, "")
  pn.AddPlace(2, 2, "")
  if len(pn.places) != 2 {
    t.Errorf("Petrinet %v should have 2 places", pn)
  }
  p1, ok1 := pn.places[1]
  if !ok1 {
    t.Error("Place with id 1 should exist")
  } else if p1.marks != 2 {
    t.Errorf("Marks of place %v should be 2", p1)
  }
  p2, ok2 := pn.places[2]
  if !ok2 {
    t.Error("Place with id 2 should exist")
  } else if p2.marks != 2 {
    t.Errorf("Marks of place %v should be 2", p2)
  }
}

func TestAddTransition(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddTransition(1, 3)
  pn.AddTransition(2, 1)
  if len(pn.transitions) != 2 {
    t.Errorf("Petrinet %v should have 2 transitions", pn)
  }
  p1, ok1 := pn.transitions[1]
  if !ok1 {
    t.Error("Transition with id 1 should exist")
  } else if p1.Priority != 3 {
    t.Errorf("Priority of transition %v should be 3", p1)
  }
  p2, ok2 := pn.transitions[2]
  if !ok2 {
    t.Error("Transition with id 2 should exist")
  } else if p2.Priority != 1 {
    t.Errorf("Priority of transition %v should be 1", p2)
  }
}

func TestAddInArcPetrinet(t *testing.T){
  pn := Init(1, "ctx1")
  pn.AddTransition(1, 1)
  pn.AddTransition(2, 1)
  pn.AddPlace(1, 2, "")
  pn.AddPlace(2, 2, "")
  pn.AddInArc(1, 1, 3) // from, transition, weight
  pn.AddInArc(1, 2, 3)
  pn.AddInArc(2, 2, 3)
  tr1 := pn.transitions[1]
  if len(tr1.inArcs) != 1 {
    t.Errorf("Transition 1 should only have 1 in arc but have: %v", tr1.inArcs)
  } else {
    p := Place{1,2,"",false}
    expectedArc := arc{&p, 3}
    if !reflect.DeepEqual(expectedArc, tr1.inArcs[0]) {
      t.Errorf("In arc is wrong, was %v but expected %v", tr1.inArcs[0], expectedArc)
    }
  }
  tr2 := pn.transitions[2]
  if len(tr2.inArcs) != 2 {
    t.Errorf("Transition 2 should have 2 in arc but have: %v", tr2.inArcs)
  } else {
    p := Place{1,2,"",false}
    expectedArc1 := arc{&p, 3}
    p2 := Place{2,2,"",false}
    expectedArc2 := arc{&p2, 3}
    if !((reflect.DeepEqual(expectedArc1, tr2.inArcs[0]) && reflect.DeepEqual(expectedArc2, tr2.inArcs[1])) ||
        (reflect.DeepEqual(expectedArc1, tr2.inArcs[1]) && reflect.DeepEqual(expectedArc2, tr2.inArcs[0]))) {
      t.Errorf("In arc is wrong, was %v but expected %v", tr2.inArcs[0], expectedArc1)
      t.Errorf("In arc is wrong, was %v but expected %v", tr2.inArcs[1], expectedArc2)
    }
  }
}

func TestAddOutArcPetrinet(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddTransition(1, 1)
  pn.AddTransition(2, 1)
  pn.AddPlace(1, 2, "")
  pn.AddPlace(2, 2, "")
  pn.AddOutArc(1, 1, 3) // transition, to, weight
  pn.AddOutArc(2, 1, 3)
  pn.AddOutArc(2, 2, 3)
  tr1 := pn.transitions[1]
  if len(tr1.outArcs) != 1 {
    t.Errorf("Transition 1 should only have 1 out arcs but have: %v", tr1.outArcs)
  } else {
    p := Place{1,2,"",false}
    expectedArc := arc{&p, 3}
    if !reflect.DeepEqual(expectedArc, tr1.outArcs[0]) {
      t.Errorf("Out arc is wrong, was %v but expected %v", tr1.outArcs[0], expectedArc)
    }
  }
  tr2 := pn.transitions[2]
  if len(tr2.outArcs) != 2 {
    t.Errorf("Transition 2 should have 2 out arcs but have: %v", tr2.outArcs)
  } else {
    p := Place{1,2,"",false}
    expectedArc1 := arc{&p, 3}
    p2 := Place{2,2,"",false}
    expectedArc2 := arc{&p2, 3}
    if !((reflect.DeepEqual(expectedArc1, tr2.outArcs[0]) && reflect.DeepEqual(expectedArc2, tr2.outArcs[1])) ||
        (reflect.DeepEqual(expectedArc1, tr2.outArcs[1]) && reflect.DeepEqual(expectedArc2, tr2.outArcs[0]))) {
      t.Errorf("Out arc is wrong, was %v but expected %v", tr2.outArcs[0], expectedArc1)
      t.Errorf("Out arc is wrong, was %v but expected %v", tr2.outArcs[1], expectedArc2)
    }
  }
}

func TestAddInhibitorArcPetrinet(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddTransition(1, 1)
  pn.AddTransition(2, 1)
  pn.AddPlace(1, 2, "")
  pn.AddPlace(2, 2, "")
  pn.AddInhibitorArc(1, 1, 3) // from, transition, weight
  pn.AddInhibitorArc(1, 2, 3)
  pn.AddInhibitorArc(2, 2, 3)
  tr1 := pn.transitions[1]
  if len(tr1.inhibitorArcs) != 1 {
    t.Errorf("Transition 1 should only have 1 inhib arcs but have: %v", tr1.inhibitorArcs)
  } else {
    p := Place{1,2,"",false}
    expectedArc := arc{&p, 3}
    if !reflect.DeepEqual(expectedArc, tr1.inhibitorArcs[0]) {
      t.Errorf("Inhib arc is wrong, was %v but expected %v", tr1.inhibitorArcs[0], expectedArc)
    }
  }
  tr2 := pn.transitions[2]
  if len(tr2.inhibitorArcs) != 2 {
    t.Errorf("Transition 2 should have 2 inhib arcs but have: %v", tr2.inhibitorArcs)
  } else {
    p := Place{1,2,"",false}
    expectedArc1 := arc{&p, 3}
    p2 := Place{2,2,"",false}
    expectedArc2 := arc{&p2, 3}
    if !((reflect.DeepEqual(expectedArc1, tr2.inhibitorArcs[0]) && reflect.DeepEqual(expectedArc2, tr2.inhibitorArcs[1])) ||
        (reflect.DeepEqual(expectedArc1, tr2.inhibitorArcs[1]) && reflect.DeepEqual(expectedArc2, tr2.inhibitorArcs[0]))) {
      t.Errorf("Inhib arc is wrong, was %v but expected %v", tr2.inhibitorArcs[0], expectedArc1)
      t.Errorf("Inhib arc is wrong, was %v but expected %v", tr2.inhibitorArcs[1], expectedArc2)
    }
  }
}

func TestAddRemoteTransition(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddRemoteTransition(1)
  if len(pn.remoteTransitions) != 1 {
    t.Errorf("Remote transitions should have length 1 %v", pn.remoteTransitions)
  } else if rt, ok := pn.remoteTransitions[1]; !ok || rt.ID != 1 {
    t.Errorf("Remote transitions should have id 1 %v", pn.remoteTransitions[1])
  }
}

func TestAddRemoteInArc(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddRemoteTransition(1)
  pn.AddRemoteInArc(1, 1, 2, "testCtx")
  rarc := RemoteArc{1,"testCtx","",2,0}
  if len(pn.remoteTransitions[1].InArcs) != 1 {
    t.Errorf("Wrong number of in arcs on remote transition: %v", pn.remoteTransitions[1].InArcs)
  } else if !reflect.DeepEqual(pn.remoteTransitions[1].InArcs[0], rarc) {
    t.Errorf("Wrong remote in arc, expected %v but was %v", rarc, pn.remoteTransitions[1].InArcs[0])
  }
}

func TestAddRemoteOutArc(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddRemoteTransition(1)
  pn.AddRemoteOutArc(1, 1, 2, "testCtx")
  rarc := RemoteArc{1,"testCtx","",2,0}
  if len(pn.remoteTransitions[1].OutArcs) != 1 {
    t.Errorf("Wrong number of out arcs on remote transition: %v", pn.remoteTransitions[1].OutArcs)
  } else if !reflect.DeepEqual(pn.remoteTransitions[1].OutArcs[0], rarc) {
    t.Errorf("Wrong remote out arc, expected %v but was %v", rarc, pn.remoteTransitions[1].OutArcs[0])
  }
}

func TestAddRemoteInhibitorArc(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddRemoteTransition(1)
  pn.AddRemoteInhibitorArc(1, 1, 2, "testCtx")
  rarc := RemoteArc{1,"testCtx","",2,0}
  if len(pn.remoteTransitions[1].InhibitorArcs) != 1 {
    t.Errorf("Wrong number of out arcs on remote transition: %v", pn.remoteTransitions[1].InhibitorArcs)
  } else if !reflect.DeepEqual(pn.remoteTransitions[1].InhibitorArcs[0], rarc) {
    t.Errorf("Wrong remote out arc, expected %v but was %v", rarc, pn.remoteTransitions[1].InhibitorArcs[0])
  }
}

func TestGetCurrentState(t *testing.T) {
  pn := Init(1,"ctx1")
  pn.AddPlace(1,1,"")
  pn.AddPlace(2,0,"")
  pn.AddPlace(3,1,"")
  expected := make(map[int]int)
  expected[1]=1
  expected[2]=0
  expected[3]=1
  must,result := pn.getCurrentState()
  if !must {
    t.Error("Wrong result there should not be a temporal place with marks")
  }
  if len(expected)!=len(result) {
    t.Errorf("Expected length %v but was %v",len(expected), len(result))
  }
  for key, value := range expected {
    if value != result[key] {
      t.Errorf("Expected marks in %v to be %v but was %v",key,value,result[key])
    }
  }
}

func TestGetCurrentStateFail(t *testing.T) {
  pn := Init(1,"ctx1")
  pn.AddPlace(1,1,"")
  pn.AddPlace(2,0,"")
  pn.AddPlace(3,1,"")
  pn.SetPlaceTemporal(1)
  must, _:= pn.getCurrentState()
  if must {
    t.Error("Wrong result there should exist a temporal place with marks")
  }
}

func TestMarksHistory(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddTransition(1, 2)
  pn.AddTransition(2, 2)
  pn.AddPlace(1, 0, "")
  pn.AddPlace(2, 2, "")
  pn.AddInArc(1, 2, 1)
  pn.AddOutArc(1, 1, 1)
  pn.AddOutArc(2, 2, 1)
  pn.SetPlaceTemporal(1)
  pn.FireTransitionByID(1)
  expected := make(map[int]int)
  expected[1] = 0
  expected[2] = 2
  helper := func(expectedLen int, result []map[int]int, expectedLast map[int]int) {
    if len(result) != expectedLen {
    t.Errorf("Expected history to have %v item but have %v", expectedLen, result)
    } else {
      if len(expected)!=len(result[expectedLen - 1]) {
        t.Errorf("Expected length %v but was %v",len(expected), len(result[expectedLen - 1]))
      }
      for key, value := range expected {
        if value != result[expectedLen - 1][key] {
          t.Errorf("Expected marks in %v to be %v but was %v",key,value,result[expectedLen - 1][key])
        }
      }
    }
  }
  helper(1, pn.marksHistory, expected)

  pn.FireTransitionByID(2)
  helper(1, pn.marksHistory, expected)

  pn.FireTransitionByID(1)
  expected[2] = 3
  helper(2, pn.marksHistory, expected)
}

func TestRollBack(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddTransition(1, 2)
  pn.AddTransition(2, 2)
  pn.AddPlace(1, 0, "")
  pn.AddPlace(2, 2, "")
  pn.AddInArc(1, 2, 1)
  pn.AddOutArc(1, 1, 1)
  pn.AddOutArc(2, 2, 1)
  pn.SetPlaceTemporal(1)
  pn.FireTransitionByID(1)
  pn.FireTransitionByID(2)
  pn.FireTransitionByID(1)
  expected := make(map[int]int)
  expected[1] = 0
  expected[2] = 3
  helper := func() {
    for id, place := range pn.places {
      if expected[id] != place.marks {
        t.Errorf("Error on rollback expected state %v but is %v",expected,pn.places)
        break
      }
    }
  }
  pn.RollBack()
  helper()
  expected[2]=2
  pn.RollBack()
  helper()
  err := pn.RollBack()
  if err == nil {
    t.Error("There must be an error")
  }
}

func TestRollBackTemporal(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddTransition(1, 2)
  pn.AddTransition(2, 2)
  pn.AddPlace(1, 0, "")
  pn.AddPlace(2, 2, "")
  pn.AddInArc(1, 2, 1)
  pn.AddOutArc(1, 1, 1)
  pn.AddOutArc(2, 2, 1)
  pn.SetPlaceTemporal(1)
  pn.FireTransitionByID(1)
  pn.FireTransitionByID(2)
  pn.FireTransitionByID(1)
  expected := make(map[int]int)
  expected[1] = 0
  expected[2] = 3
  helper := func() {
    for id, place := range pn.places {
      if expected[id] != place.marks {
        t.Errorf("Error on rollback expected state %v but is %v",expected,pn.places)
        break
      }
    }
  }
  pn.RollBackTemporal()
  helper() // should rollback to expected
  pn.RollBackTemporal()
  helper() // shouldn't rollback because there is no mark on temporal
}

func TestGenerateUniversalTransitions(t *testing.T) {
    pn := Init(1, "universalCtx")
    pn.AddTransition(1, 0)
    pn.AddTransition(2, 0)
    pn.AddTransition(3, 1)
    pn.AddRemoteTransition(1)
    pn.AddRemoteTransition(2)
    pn.AddRemoteTransition(3)
    pn.AddRemoteInArc(1, 1, 3, "ctx1")
    pn.AddRemoteInArc(2, 1, 3, "ctx1")
    pn.AddRemoteOutArc(1, 1, 3, "ctx4")
    pn.AddRemoteOutArc(1, 1, 3, "ctx3")
    pn.AddRemoteInhibitorArc(3, 1, 3, "ctx1")

    pn.AddRemoteInhibitorArc(3, 3, 3, "ctx1") // priority 1

    pn.AddRemoteOutArc(2, 2, 3, "ctx3")
    pn.AddRemoteOutArc(2, 2, 3, "ctx5") // doesnt exist, should add no remote transitions of 2
    ctxToAddrs := make(map[string][]string)
    ctxToAddrs["ctx1"] = []string{"addr1", "addr2"}
    ctxToAddrs["ctx2"] = []string{"addr3", "addr4", "addr5"}
    ctxToAddrs["ctx3"] = []string{"addr6"}
    ctxToAddrs["ctx4"] = []string{"addr7", "addr8"}

    transitions, rmtTransitions := pn.GenerateUniversalTransitionsByPriority(ctxToAddrs, 1)
    t.Logf("Transitions: %v, Remote transitions: %v", transitions, rmtTransitions)
    t.Logf("Rmt1: %v, Rmt2: %v", rmtTransitions[0], rmtTransitions[1])
    if len(transitions) != 2 {
      t.Errorf("Expected to generate 2 transitions but did %v", len(transitions))
    }
    expected1 := RemoteTransition{0,
      []RemoteArc{},
      []RemoteArc{},
      []RemoteArc{{3, "ctx1", "addr1", 3, 0}}}
    expected2 := RemoteTransition{0,
      []RemoteArc{},
      []RemoteArc{},
      []RemoteArc{{3, "ctx1", "addr2", 3, 0}}}
    found1 := false
    found2 := false
    for ind, tr := range transitions {
      foundOne := false
      if !found1 && len(rmtTransitions[tr.ID].InhibitorArcs) == 1 && rmtTransitions[tr.ID].InhibitorArcs[0] == expected1.InhibitorArcs[0] {
        foundOne = true
        found1 = true
      } else if !found2 && len(rmtTransitions[tr.ID].InhibitorArcs) == 1 && rmtTransitions[tr.ID].InhibitorArcs[0] == expected2.InhibitorArcs[0] {
        foundOne = true
        found2 = true
      }
      if !foundOne {
        t.Errorf("Wrong remote transition %v found at %v, expected it to be %v or %v", rmtTransitions[tr.ID], ind, expected1, expected2)
      }
    }

    transitions, rmtTransitions = pn.GenerateUniversalTransitionsByPriority(ctxToAddrs, 0)
    t.Logf("Transitions: %v, Remote transitions: %v", transitions, rmtTransitions)
    for ind, rmtT := range rmtTransitions {
      t.Logf("Rmt Transition %v: %v", ind, rmtT)
    }
    if len(transitions) != 4 {
      t.Errorf("Expected to generate 4 transitions but did %v", len(transitions))
    }
    expectedTr1 := RemoteTransition{1,
      []RemoteArc{{1, "ctx1", "addr1", 3, 0}, {2, "ctx1", "addr1", 3, 0}},
      []RemoteArc{{1, "ctx4", "addr7", 3, 0}, {1, "ctx3", "addr6", 3, 0}},
      []RemoteArc{{3, "ctx1", "addr1", 3, 0}}}
    expectedTr2 := RemoteTransition{1,
      []RemoteArc{{1, "ctx1", "addr1", 3, 0}, {2, "ctx1", "addr1", 3, 0}},
      []RemoteArc{{1, "ctx4", "addr8", 3, 0}, {1, "ctx3", "addr6", 3, 0}},
      []RemoteArc{{3, "ctx1", "addr1", 3, 0}}}
    expectedTr3 := RemoteTransition{1,
      []RemoteArc{{1, "ctx1", "addr2", 3, 0}, {2, "ctx1", "addr2", 3, 0}},
      []RemoteArc{{1, "ctx4", "addr7", 3, 0}, {1, "ctx3", "addr6", 3, 0}},
      []RemoteArc{{3, "ctx1", "addr2", 3, 0}}}
    expectedTr4 := RemoteTransition{1,
      []RemoteArc{{1, "ctx1", "addr2", 3, 0}, {2, "ctx1", "addr2", 3, 0}},
      []RemoteArc{{1, "ctx4", "addr8", 3, 0}, {1, "ctx3", "addr6", 3, 0}},
      []RemoteArc{{3, "ctx1", "addr2", 3, 0}}}
    found1 = false
    found2 = false
    found3 := false
    found4 := false
    for ind, tr := range transitions {
      foundOne := false
      if !found1 && sliceContainsAllRemoteArcs(rmtTransitions[tr.ID].InhibitorArcs, expectedTr1.InhibitorArcs) &&
      sliceContainsAllRemoteArcs(rmtTransitions[tr.ID].InArcs, expectedTr1.InArcs) &&
      sliceContainsAllRemoteArcs(rmtTransitions[tr.ID].OutArcs, expectedTr1.OutArcs) {
        t.Log("Found expected tr1")
        foundOne = true
        found1 = true
      } else if !found2 && sliceContainsAllRemoteArcs(rmtTransitions[tr.ID].InhibitorArcs, expectedTr2.InhibitorArcs) &&
      sliceContainsAllRemoteArcs(rmtTransitions[tr.ID].InArcs, expectedTr2.InArcs) &&
      sliceContainsAllRemoteArcs(rmtTransitions[tr.ID].OutArcs, expectedTr2.OutArcs) {
        t.Log("Found expected tr2")
        foundOne = true
        found2 = true
      } else if !found3 && sliceContainsAllRemoteArcs(rmtTransitions[tr.ID].InhibitorArcs, expectedTr3.InhibitorArcs) &&
      sliceContainsAllRemoteArcs(rmtTransitions[tr.ID].InArcs, expectedTr3.InArcs) &&
      sliceContainsAllRemoteArcs(rmtTransitions[tr.ID].OutArcs, expectedTr3.OutArcs) {
        t.Log("Found expected tr3")
        foundOne = true
        found3 = true
      } else if !found4 && sliceContainsAllRemoteArcs(rmtTransitions[tr.ID].InhibitorArcs, expectedTr4.InhibitorArcs) &&
      sliceContainsAllRemoteArcs(rmtTransitions[tr.ID].InArcs, expectedTr4.InArcs) &&
      sliceContainsAllRemoteArcs(rmtTransitions[tr.ID].OutArcs, expectedTr4.OutArcs) {
        t.Log("Found expected tr4")
        foundOne = true
        found4 = true
      }
      if !foundOne {
        t.Errorf("Wrong remote transition %v found at %v, expected it to be %v, %v, %v or %v", rmtTransitions[tr.ID], ind, expectedTr1, expectedTr2, expectedTr3, expectedTr4)
      }
    }
}

func TestHasBeenFired(t *testing.T) {
  pn := Init(1, "ctx1")
  pn.AddTransition(3, 2)
  pn.AddPlace(1, 1, "")
  pn.AddPlace(2, 2, "")
  pn.AddPlace(3, 5, "")
  pn.AddPlace(4, 0, "")
  pn.AddPlace(5, 1, "")
  pn.AddPlace(6, 0, "")
  pn.AddInArc(1, 3, 1)
  pn.AddInArc(2, 3, 2)
  pn.AddInArc(3, 3, 3)
  pn.AddOutArc(3, 4, 3)
  pn.AddOutArc(3, 5, 2)
  pn.AddOutArc(3, 6, 3)
  pn.FireTransitionByID(3)
  if !pn.HasBeenFired() {
    t.Error("Should have been fired")
  }
}
