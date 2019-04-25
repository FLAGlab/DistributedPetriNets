package petrinet

import (
	"fmt"
)

type Place struct {
	ID int
	marks int
	label string
	temporal bool
}

func (a Place) String() string {
	return fmt.Sprintf("{id: %v, marks: %v, label: %v, temporal: %v}", a.ID, a.marks, a.label, a.temporal)
}

// GetMarks gets the marks on the place
func (a *Place) GetMarks() int {
	return a.marks
}

func (a *Place) setTemporal(isTemporal bool) {
	a.temporal = isTemporal
}
