package petrinet

import (
	"fmt"
)

// Place of the Petri net
type Place struct {
	ID    int
	Marks int
	Label string
}

func (p Place) String() string {
	return fmt.Sprintf("{id: %v, marks: %v, label: %v}", p.ID, p.Marks, p.Label)
}

// GetMarks gets the marks on the place
func (p *Place) GetMarks() int {
	return p.Marks
}

//InitService creates and runs the node containing the net
func (p *Place) InitService(_serviceName string) {
	srv := &ServiceNode{
		PetriPlace:  p,
		ServiceName: _serviceName, //strconv.Itoa(p.ID),
	}
	srv.RunService()
}
