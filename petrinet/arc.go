package petrinet

import (
  "fmt"
)

type arc struct {
	_place *place
	weight int
}

func (a arc) String() string {
	return fmt.Sprintf("{place: %v, weight: %v}", a._place, a.weight)
}