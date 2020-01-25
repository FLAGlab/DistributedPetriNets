package petrinet

import (
	"fmt"
)

// Place of the Petri net
type Place struct {
	ID    int
	Marks []Token
	Label string
}

func (p Place) String() string {
	return fmt.Sprintf("{id: %v, marks: %v, label: %v}", p.ID, p.Marks, p.Label)
}

// GetMarks gets the marks on the place
func (p *Place) GetNumMarks() int {
	return len(p.Marks)
}

func (p *Place) AddMarks(t []Token) {
	p.Marks = append(p.Marks, t...)
}

func (p *Place) GetMark(l int) []Token {
	fmt.Printf("%v %v/n",l,p.Marks)
	x := p.Marks[0:l]
	p.Marks = p.Marks[l:]
	fmt.Printf("%v %v/n",l,x)
	return x
}
//InitService creates and runs the node containing the net
func (p *Place) InitService(_serviceName, interf string) {
	srv := &ServiceNode{
		Interface: interf,
		PetriPlace:  p,
		ServiceName: _serviceName, //strconv.Itoa(p.ID),
	}
	srv.RunService()
}
