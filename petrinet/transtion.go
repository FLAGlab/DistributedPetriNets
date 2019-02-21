package petrinet

type transition struct {
	id int
	priority int
	inArcs []Arc
	outArcs []Arc
	inhibitorArcs []Arc
}

func (t *transition) canFire() bool {
	ans := true
  for _, currArc := range t.inArcs {
		ans = ans && currArc._place.marks >= value.weight
  }
  for _, value := range t.inhibitorArcs {
    ans = ans && currArc._place.marks < value.weight
  }
  return ans
}

func (t *transition) fire() {
	for _, currArc := range t.inArcs {
    currArc._place.marks -= currArc.weight
  }
  for _, currArc := range t.outArcs {
    currArc._place.marks += currArc.weight
  }
}
