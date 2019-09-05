package petrinet

import (
	//"reflect"
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
