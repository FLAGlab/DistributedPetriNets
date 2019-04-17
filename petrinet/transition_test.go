package petrinet

import (
  "testing"
  "reflect"
)

func initTestTransition() *Transition {
  return &Transition{1, 1, nil, nil, nil}
}

func TestAddInArc(t *testing.T) {
  tr := initTestTransition()
  p := Place{1,1,""}
  newArc := arc{&p, 1}
  tr.addInArc(newArc)
  exists := false
  for _, item := range tr.inArcs {
    exists = exists || reflect.DeepEqual(newArc, item)
  }
  if !exists {
    t.Errorf("Couldn't add in arc %v to transition %v.\n", newArc, tr)
  }
}

func TestAddOutArc(t *testing.T) {
  tr := initTestTransition()
  p := Place{1,1,""}
  newArc := arc{&p, 1}
  tr.addOutArc(newArc)
  exists := false
  for _, item := range tr.outArcs {
    exists = exists || reflect.DeepEqual(newArc, item)
  }
  if !exists {
    t.Errorf("Couldn't add in arc %v to transition %v.\n", newArc, tr)
  }
}

func TestAddInhibArc(t *testing.T) {
  tr := initTestTransition()
  p := Place{1,1,""}
  newArc := arc{&p, 1}
  tr.addInhibitorArc(newArc)
  exists := false
  for _, item := range tr.inhibitorArcs {
    exists = exists || reflect.DeepEqual(newArc, item)
  }
  if !exists {
    t.Errorf("Couldn't add in arc %v to transition %v.\n", newArc, tr)
  }
}

func TestCanFire(t *testing.T) {
  tr := initTestTransition()
  p := Place{1,1,""}
  newArc := arc{&p, 1}
  tr.addInArc(newArc)
  // can fire
  if !tr.canFire() {
    t.Errorf("Transition %v should be able to fire with arc %v", tr, newArc)
  }
  // can't fire because of in Arcs
  p.marks = 0
  if tr.canFire() {
    t.Errorf("Transition %v should NOT be able to fire with arc %v", tr, newArc)
  }
  p.marks = 1
  //cant fire because of inhib arcs
  p2 := Place{1,1,""}
  inhibArc := arc{&p2, 1}
  tr.addInhibitorArc(inhibArc)
  if tr.canFire() {
    t.Errorf("Transition %v should NOT be able to fire with inhib arc %v", tr, inhibArc)
  }
}

func TestFire(t *testing.T) {
  tr := initTestTransition()
  inPlaces := []Place{{1,1,""}, {2,2,""}, {3,5,""}}
  outPlaces := []Place{{4,0,""}, {5,1,""}, {6,0,""}}
  inArcs := []arc{{&inPlaces[0], 1}, {&inPlaces[1], 2}, {&inPlaces[2], 3}}
  outArcs := []arc{{&outPlaces[0], 3}, {&outPlaces[1], 2}, {&outPlaces[2], 3}}
  tr.inArcs = inArcs
  tr.outArcs = outArcs
  tr.fire()
  expectedIn := []int{0, 0, 2}
  for index, value := range expectedIn {
    if inPlaces[index].marks != value {
      t.Errorf("Place %v should have been fired on transition %v, expected it to have %v marks but had %v", inPlaces[index], tr, value, inPlaces[index].marks)
    }
  }
  expectedOut := []int{3, 3, 3}
  for index, value := range expectedOut {
    if outPlaces[index].marks != value {
      t.Errorf("Place %v should have received marks on transition %v, expected it to have %v marks but had %v", outPlaces[index], tr, value, outPlaces[index].marks)
    }
  }
}
