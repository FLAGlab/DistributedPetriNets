package petrinet

import (
	"fmt"
)

type place struct {
	id int
	marks int
	label string
}

func (a place) String() string {
	return fmt.Sprintf("{id: %v, marks: %v, label: %v}", a.id, a.marks, a.label)
}