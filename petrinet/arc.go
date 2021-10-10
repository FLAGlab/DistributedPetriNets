package petrinet

import (
	"fmt"
)

// Arc between Petri net components
type Arc struct {
	Place   *Place `json:"place"`
	Weight  int    `json:"weight"`
	PlaceID int    `json:"placeId"`
}

func (a Arc) String() string {
	return fmt.Sprintf("{place: %v, weight: %v}", a.Place, a.Weight)
}
