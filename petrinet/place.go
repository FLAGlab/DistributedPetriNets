package petrinet

import (
	"fmt"
)

type Place struct {
	ID int
	marks int
	label string
}

func (a Place) String() string {
	return fmt.Sprintf("{id: %v, marks: %v, label: %v}", a.ID, a.marks, a.label)
}

// GetMarks gets the marks on the place
func (a *Place) GetMarks() int {
	return a.marks
}
