package conflictsolver

import (
  "testing"

  "github.com/FLAGlab/DCoPN/petrinet"
)

func TestAddConflict(t *testing.T) {
  cs := InitCS()
  cs.AddConflict("ctx1","ctx2",1,1)
  expected := conflict{"ctx1","ctx2",1,1}
  exists := false
  for _, item := range cs.conflicts {
    exists = exists || item == expected
  }
  if !exists {
    t.Errorf("Couldn't add conflict %v to conflictSolver %v.\n", expected, cs.conflicts)
  }
}

func TestGetRequiredPlacesByAddress(t *testing.T) {
	cs := InitCS()
	cs.AddConflict("ctx1","ctx2",1,2)
	cs.AddConflict("ctx1","ctx3",2,3)
	ctx2address := make(map[string][]string)
	ctx2address["ctx1"] = []string{"add1","add2"}
	ctx2address["ctx2"] = []string{"add3","add4"}
	expected := make(map[string][]int)
	expected["add1"] = []int{1}
	expected["add2"] = []int{1}
	expected["add3"] = []int{2}
	expected["add4"] = []int{2}
	result := cs.GetRequiredPlacesByAddress(ctx2address)
	if len(result) != len(expected) {
		t.Errorf("Expected %v but result is %v",expected,result)
	}
	for key,value := range expected {
		if len(result[key]) != 1 || result[key][0] != value[0] {
			t.Errorf("Expected address %v to have %v but had %v", key, value, result[key])
		}
	}
}

func TestGetConflictedAddrs(t *testing.T) {
	cs := InitCS()
	cs.AddConflict("ctx1","ctx2",1,2)
	cs.AddConflict("ctx1","ctx3",2,3)
	ctx2address := make(map[string][]string)
	ctx2address["ctx1"] = []string{"add1","add2"}
	ctx2address["ctx2"] = []string{"add3","add4"}
	marks:= make(map[string]map[int]*petrinet.RemoteArc)
	marks["add1"] = make(map[int]*petrinet.RemoteArc)
	marks["add2"] = make(map[int]*petrinet.RemoteArc)
	marks["add3"] = make(map[int]*petrinet.RemoteArc)
	marks["add4"] = make(map[int]*petrinet.RemoteArc)
	marks["add1"][1] = &petrinet.RemoteArc{Marks: 0}
	marks["add2"][1] = &petrinet.RemoteArc{Marks: 1}
	marks["add3"][2] = &petrinet.RemoteArc{Marks: 0}
	marks["add4"][2] = &petrinet.RemoteArc{Marks: 1}
	result := cs.GetConflictedAddrs(marks,ctx2address)
	expected := make(map[string]bool)
	expected["add2"] = true
	expected["add4"] = true
	if len(result) != len(expected) {
		t.Errorf("Expected %v but result is %v",expected,result)
	}
	for key,value := range expected {
		if result[key] != value {
			t.Errorf("Expected address %v to have %v but had %v", key, value, result[key])
		}
	}

}
