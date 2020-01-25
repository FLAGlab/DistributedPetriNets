package petrinet

import (
	"fmt"
)

// Token for a petri net
type Token struct {
	Id int
}

func (a Token) String() string {
	return fmt.Sprintf("{Id: %v}", a.Id)
}