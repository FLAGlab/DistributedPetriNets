package petrinet

import (
	"fmt"
	"os"
	"log"
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
	f, err := os.OpenFile(p.Label+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
        log.Println(err)
	}
	defer f.Close()

	logger := log.New(f, "", log.LstdFlags)
	logger.Printf(", %v\n", t)
	p.Marks = append(p.Marks, t...)
}

func (p *Place) GetMark(l int) []Token {
	x := p.Marks[0:l]
	p.Marks = p.Marks[l:]
	fmt.Printf("%v %v/n", l, x)
	return x
}

//InitService creates and runs the node containing the net
func (p *Place) InitService(_serviceName, interf string) {
	srv := &ServiceNode{
		Interface:   interf,
		PetriPlace:  p,
		ServiceName: _serviceName, //strconv.Itoa(p.ID),
	}
	srv.RunService()
}
