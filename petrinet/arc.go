package petrinet

import (
	"fmt"
)

// Arc between Petri net components
type Arc struct {
	Place  *Place
	Weight int
}

func (a Arc) String() string {
	return fmt.Sprintf("{place: %v, weight: %v}", a.Place, a.Weight)
}
