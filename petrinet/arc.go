package petrinet

import (
  "fmt"
)

type Arc struct {
	p int
	t int
	w int
}
func (a Arc) String() string {
	return fmt.Sprintf("{p: %v, t: %v, w: %v}", a.p, a.t, a.w)
}