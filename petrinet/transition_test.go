package petrinet

import (
	"reflect"
	"testing"
)

func initTestTransition() *Transition {
	return &Transition{1, 1, nil, nil, nil}
}

func TestAddInArc(t *testing.T) {
	tr := initTestTransition()
	p := Place{1, 1, ""}
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
	p := Place{1, 1, ""}
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

func TestAddOutRemoteArc(t *testing.T) {
	tr := initTestTransition()
	newrarc := RemoteArc{1, "127.0.0.1", 1, 1}
	tr.addRemoteOutArc(newrarc)
	exists := false
	for _, item := range tr.remoteOutArcs {
		exists = exists || reflect.DeepEqual(newrarc, item)
	}
	if !exists {
		t.Errorf("Couldn't add remote out arc %v to transition %v.\n", newrarc, tr)
	}
}
