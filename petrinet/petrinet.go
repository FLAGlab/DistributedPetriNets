/*
P => places
T => transitions
F subset of (P X T) U (T X P) => arcs
W: F -> {1,2,3,...} => weights
M_0: P -> {0, 1, 2, 3, ...} => Initial marking
{
  P: Number => 1...P places
  T: Number => 1...T transitions
  Fpt: [{p: Number, t: Number, w: Number}] => Arcs from P to T list
  Ftp: [{p: Number, t: Number, w: Number}] => Arcs from T to P list
  M: [{p: Number, m: Number}] => Initial marking list,
  I: Number => max number of transitions to fire
}

eg.
{
  P: 4,
  T: 2,
  Fpt: [{p: 1, t: 1, w: 1}, {p: 2, t: 2, w: 1}, {p: 3, t: 2, w: 1}],
  Ftp: [{t: 1, p: 2, w: 1}, {t: 1, p: 3, w: 1}, {t: 2, p: 4, w: 1}, {t: 2, p: 1, w: 1}],
  M: [{p: 1, m: 1}, {p: 3, m: 2}, {p: 4, m: 1}],
  I: 10
}
*/
package main

import (
  "fmt"
  "math/rand"
)

type Arc struct {
  p int
  t int
  w int
}
func (a Arc) String() string {
	return fmt.Sprintf("{p: %v, t: %v, w: %v}", a.p, a.t, a.w)
}

type PetriNet struct {
  p int
  t int
  tIn map[int][]Arc
  tOut map[int][]Arc
  m map[int]int
  //it int
}
func (pn PetriNet) String() string {
	return fmt.Sprintf("{\n\tp: %v, \n\tt: %v, \n\ttIn: %v, \n\ttOut: %v, \n\tm: %v\n}", pn.p, pn.t, pn.tIn, pn.tOut, pn.m)
}

func initPetriNet(p, t int, fpt, ftp []Arc, m0 map[int]int) PetriNet {
  tIn  := make(map[int][]Arc)
  tOut := make(map[int][]Arc)

  for _, arc := range fpt {
    tIn[arc.t] = append(tIn[arc.t], arc)
  }
  for _, arc := range ftp {
    tOut[arc.t] = append(tOut[arc.t], arc)
  }
  return PetriNet{p, t, tIn, tOut, m0}
}

func (pn *PetriNet) Run() {
  transitionOptions := pn.getTransitionOptions()
  for len(transitionOptions) > 0 {
    pn.fire(transitionOptions[rand.Intn(len(transitionOptions))])
    fmt.Printf("%v\n", pn)
    transitionOptions = pn.getTransitionOptions()
  }
}

func (pn *PetriNet) getTransitionOptions() []int {
  var transitionOptions []int
  for i := 1; i <= pn.t; i++ {
    if (pn.canTransition(i)) {
        transitionOptions = append(transitionOptions, i)
    }
  }
  return transitionOptions
}

func (pn *PetriNet) canTransition(currT int) bool {
  ans := true
  inArcs := pn.tIn[currT]
  for _, value := range inArcs {
    ans = ans && pn.m[value.p] >= value.w
  }
  return ans
}

func (pn *PetriNet) fire(currT int) {
  inArcs := pn.tIn[currT]
  for _, value := range inArcs {
    pn.m[value.p] -= value.w
  }
  outArcs := pn.tOut[currT]
  for _, value := range outArcs {
    pn.m[value.p] += value.w
  }
}

func main() {
  fpt := []Arc{Arc{1, 1, 1}, Arc{2, 2, 1}, Arc{3, 2, 1}}
  ftp := []Arc{Arc{t: 1, p: 2, w: 1}, Arc{t: 1, p: 3, w: 1}, Arc{t: 2, p: 4, w: 1}}
  m := make(map[int]int)
  m[1] = 1
  m[3] = 2
  m[4] = 1
  p := initPetriNet(4, 2, fpt, ftp, m)
  fmt.Printf("%v", p)
  p.Run()
}
