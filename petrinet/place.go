package petrinet

import (
	"fmt"
	"strconv"
)

type Place struct {
	ID    int
	Marks int
	Label string
}

func (a Place) String() string {
	return fmt.Sprintf("{id: %v, marks: %v, label: %v}", a.ID, a.Marks, a.Label)
}

// GetMarks gets the marks on the place
func (a *Place) GetMarks() int {
	return a.Marks
}

func (p *Place) InitService() {
	srv := &ServiceNode{
		PetriPlace:  p,
		ServiceName: strconv.Itoa(p.ID),
	}
	srv.RunService()
}
