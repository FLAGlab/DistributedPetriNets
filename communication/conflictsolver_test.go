package communication

import (
  "testing"

  //"github.com/FLAGlab/DCoPN/petrinet"
)

func TestAddConflict(t *testing.T) {
  cs := InitCS()
  cs.AddConflict("ctx1","ctx2",1,1)
  expected := conflict{"ctx1","ctx2",1,1}
  exists := false
  for _, item := range cs.conflicts {
    exists = exists || item == expected
  }
  if !exists {
    t.Errorf("Couldn't add conflict %v to conflictSolver %v.\n", expected, cs.conflicts)
  }
}