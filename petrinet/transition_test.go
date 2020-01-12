package petrinet

/* import (
	"reflect"
	"testing"
)

func initTestTransition() *Transition {
	return &Transition{1, 1, nil, nil, nil}
}

func TestAddInArc(t *testing.T) {
	tr := initTestTransition()
	p := Place{1, 1, ""}
	newArc := Arc{&p, 1}
	tr.AddInArc(newArc)
	exists := false
	for _, item := range tr.InArcs {
		exists = exists || reflect.DeepEqual(newArc, item)
	}
	if !exists {
		t.Errorf("Couldn't add in arc %v to transition %v.\n", newArc, tr)
	}
}

func TestAddOutArc(t *testing.T) {
	tr := initTestTransition()
	p := Place{1, 1, ""}
	newArc := Arc{&p, 1}
	tr.AddOutArc(newArc)
	exists := false
	for _, item := range tr.OutArcs {
		exists = exists || reflect.DeepEqual(newArc, item)
	}
	if !exists {
		t.Errorf("Couldn't add in arc %v to transition %v.\n", newArc, tr)
	}
}

func TestAddOutRemoteArc(t *testing.T) {
	tr := initTestTransition()
	newrarc := RemoteArc{1, "127.0.0.1", 1, 1}
	tr.AddRemoteOutArc(newrarc)
	exists := false
	for _, item := range tr.RemoteOutArcs {
		exists = exists || reflect.DeepEqual(newrarc, item)
	}
	if !exists {
		t.Errorf("Couldn't add remote out arc %v to transition %v.\n", newrarc, tr)
	}
}
 */