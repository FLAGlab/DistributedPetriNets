/*
P => places
T => transitions
F subset of (P X T) U (T X P) => arcs
F0 subset of (P X T) => inhibitor arcs
W: F -> {1,2,3,...} => weights of arcs
W0: F0 -> {1,2,3,...} => weights of inhibitor arcs
M_0: P -> {0, 1, 2, 3, ...} => Initial marking
{
  P: Number => 1...P places
  T: Number => 1...T transitions
  Fpt: [{p: Number, t: Number, w: Number}] => Arcs from P to T list
  Ftp: [{p: Number, t: Number, w: Number}] => Arcs from T to P list
  Inhibitors: [{p: Number, t: Number, w: Number}] => Inhibitor Arcs from P to T list
  M: [{p: Number, m: Number}] => Initial marking list,
}

eg.
{
  P: 4,
  T: 2,
  Fpt: [{p: 1, t: 1, w: 1}, {p: 2, t: 2, w: 1}, {p: 3, t: 2, w: 1}],
  Ftp: [{t: 1, p: 2, w: 1}, {t: 1, p: 3, w: 1}, {t: 2, p: 4, w: 1}],
  Inhibitors: [{t: 2, p: 4, w: 1 }],
  M: [{p: 1, m: 1}, {p: 3, m: 2}, {p: 4, m: 1}],
}
*/
package petrinet

import (
  "fmt"
  "math/rand"
)

type PetriNet struct {
  p int
  t int

  tIn map[int][]Arc
  tOut map[int][]Arc
  inhibitors map[int][]Arc
  m map[int]int
  //it int
}
func (pn PetriNet) String() string {
	return fmt.Sprintf("{\n\tp: %v, \n\tt: %v, \n\ttIn: %v, \n\ttOut: %v, \n\tm: %v\n}", pn.p, pn.t, pn.tIn, pn.tOut, pn.m)
}

func InitPetriNet(p, t int, fpt, ftp, inhi []Arc, m0 map[int]int) PetriNet {
  tIn  := make(map[int][]Arc)
  tOut := make(map[int][]Arc)
  inhibitors := make(map[int][]Arc)

  for _, arc := range fpt {
    tIn[arc.t] = append(tIn[arc.t], arc)
  }
  for _, arc := range ftp {
    tOut[arc.t] = append(tOut[arc.t], arc)
  }
  for _, arc :=range inhi {
    inhibitors[arc.t] = append(inhibitors[arc.t],arc)
  }
  return PetriNet{p, t, tIn, tOut, inhibitors, m0}
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
  inhibArcs := pn.inhibitors[currT]
  for _, value := range inhibArcs {
    ans = ans && pn.m[value.p] < value.w
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

func Test(){
  fpt := []Arc{Arc{p: 1, t: 1, w: 1}, Arc{p: 2, t: 2, w: 1}, Arc{p: 3, t: 2, w: 1}}
	ftp := []Arc{Arc{t: 1, p: 2, w: 1}, Arc{t: 1, p: 3, w: 1}, Arc{t: 2, p: 4, w: 1}}
	m := make(map[int]int)
	inhi := []Arc{Arc{t: 2, p: 4, w:1}}
	m[1] = 1
	m[3] = 2
	m[4] = 1
	p := InitPetriNet(4, 2, fpt, ftp, inhi, m)
  fmt.Printf("%v", p)
  p.Run()
}
/*
Hacer ejercicio de mutual exclution distribuido
que pasa si se conecta en medio de elegir y fire
Definir las reglas
ext 1 -> distribuido y experimentos (con trans normales, despues reactivas) (mail para conectar nodos)
ext 2 -> reacciones reactivas (prioridades)
*/
