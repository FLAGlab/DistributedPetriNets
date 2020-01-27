package petrinet

import (
	"fmt"
)

// Token for a petri net
type Token struct {
	ID int
}

func (t Token) String() string {
	return fmt.Sprintf("{Id: %v}", t.ID)
}
