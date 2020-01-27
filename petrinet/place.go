package petrinet

import (
	"fmt"
	"log"
	"os"
)

// Place of the Petri net
type Place struct {
	ID    int
	Marks []Token
	Label string
	Name  string
}

func (p Place) String() string {
	return fmt.Sprintf("{id: %v, marks: %v, label: %v}", p.ID, p.Marks, p.Label)
}

//GetNumMarks gets the quantity of tokens in the place
func (p *Place) GetNumMarks() int {
	return len(p.Marks)
}

//AddMarks adds an array of tokens to the current token
func (p *Place) AddMarks(t []Token) {
	f, err := os.OpenFile(p.Label+"_"+p.Name+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	logger := log.New(f, "", log.LstdFlags)
	//logger.Printf(", Service Name, Token Id")
	for _, val := range t {
		logger.Printf(", %v, %v\n", p.Name, val.ID)
	}
	defer f.Close()
	p.Marks = append(p.Marks, t...)
}

//GetMark gets the mark of the place
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
