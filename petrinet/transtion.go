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
  for _, value := range t.inArcs {
		ans = ans && value.place.marks >= value.weight
  }
  for _, value := range t.inhibitorArcs {
    ans = ans && value.place.marks < value.weight
  }
  return ans
}

func (t *transition) fire() {
	for _, currArc := range t.inArcs {
    currArc.place.marks -= currArc.weight
  }
  for _, currArc := range t.outArcs {
    currArc.place.marks += currArc.weight
  }
}
