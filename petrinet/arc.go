package petrinet

import (
  "fmt"
)

type arc struct {
	place *Place
	weight int
}

func (a arc) String() string {
	return fmt.Sprintf("{place: %v, weight: %v}", a.place, a.weight)
}
