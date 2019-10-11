package petrinet

import (
	"testing"
)

func TestAddPlace(t *testing.T) {
	pn := Init(1)
	pn.AddPlace(1, 2, "")
	pn.AddPlace(2, 2, "")
	if len(pn.places) != 2 {
		t.Errorf("Petrinet %v should have 2 places", pn)
	}
	p1, ok1 := pn.places[1]
	if !ok1 {
		t.Error("Place with id 1 should exist")
	} else if p1.Marks != 2 {
		t.Errorf("Marks of place %v should be 2", p1)
	}
	p2, ok2 := pn.places[2]
	if !ok2 {
		t.Error("Place with id 2 should exist")
	} else if p2.Marks != 2 {
		t.Errorf("Marks of place %v should be 2", p2)
	}
}

func TestCurrentState(t *testing.T) {
	pn := Init(1)
	pn.AddPlace(1, 0, "p1")
	pn.AddPlace(2, 1, "p2")
	pn.AddPlace(3, 3, "p3")
	_, resMap := pn.getCurrentState()
	expected := make(map[int]int)
	expected[1] = 0
	expected[2] = 1
	expected[3] = 3
	for key, value := range expected {
		if resMap[key] != value {
			t.Errorf("The Place %v has an incorrect marking, should be %v but is %v",
				key, value, resMap[key])
		}
	}
}

func TestFireLocalTransition(t *testing.T) {
	pn := Init(1)
	pn.AddPlace(1, 2, "p1")
	pn.AddPlace(2, 1, "p2")
	pn.AddTransition(1, 1)
	pn.AddInArc(1, 1, 1)
	pn.AddOutArc(1, 2, 1)
	pn.FireTransitionByID(1)
	expected := make(map[int]int)
	expected[1] = 1
	expected[2] = 2
	for key, value := range expected {
		if pn.places[key].Marks != value {
			t.Errorf(
				"Place %v should have been affected by transition %v, expected it to have %v marks but had %v",
				pn.places[key], pn.transitions[1], value, pn.places[key].Marks)
		}
	}
}
