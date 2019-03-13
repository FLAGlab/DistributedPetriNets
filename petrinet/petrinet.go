/*Package petrinet where:
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
  "math"
  "math/rand"
)

// PetriNet struct, has an id, transitions and places
type PetriNet struct {
 id int
 transitions map[int]*Transition
 places map[int]*place
}

func (pn PetriNet) String() string {
  s := ""
  for _, _place := range pn.places{
    s = fmt.Sprintf("%v\n%v", s, _place)
  }
  return s + "\n"
}

func (pn *PetriNet) Run() {
  transitionOptions := pn.GetTransitionOptions()
  for len(transitionOptions) > 0 {
    transitionOptions[rand.Intn(len(transitionOptions))].fire()
    fmt.Printf("%v\n", pn)
    transitionOptions = pn.GetTransitionOptions()
  }
}

// FireTransitionByID fires a transition given its ID
func (pn *PetriNet) FireTransitionByID(transitionID int) error {
  return pn.transitions[transitionID].fire()
}

// GetTransitionOptions gets all the transitions with min priority that can be
// fired
func (pn *PetriNet) GetTransitionOptions() []*Transition {
  var transitionOptions []*Transition
  currMin := math.MaxInt64
  for _, currTransition := range pn.transitions {
    if (currTransition.canFire()) {
      if (currTransition.Priority < currMin) {
        currMin = currTransition.Priority
        transitionOptions = []*Transition{currTransition}
      } else if (currTransition.Priority == currMin) {
        transitionOptions = append(transitionOptions, currTransition)
      }
    }
  }
  return transitionOptions
}

func (pn *PetriNet) AddPlace(_id, _marks int, _label string) {
  pn.places[_id] = &place{id: _id, marks: _marks, label: _label}
}

func (pn *PetriNet) AddTransition(_id, _priority int) {
  pn.transitions[_id] = &Transition {
    ID: _id,
    Priority: _priority,
    inArcs: make([]arc,0),
    outArcs: make([]arc,0),
    inhibitorArcs: make([]arc,0)}
}
func (pn *PetriNet) AddInArc(from,_transition,_weight int){
  pn.transitions[_transition].addInArc(
    arc {
      _place: pn.places[from],
      weight: _weight})
}
func (pn *PetriNet) AddOutArc(_transition, to, _weight int){

  pn.transitions[_transition].addOutArc(
    arc {
      _place: pn.places[to],
      weight: _weight})
}
func (pn *PetriNet) AddInhibitorArc(from,_transition,_weight int){
  pn.transitions[_transition].addInhibitorArc(
    arc {
      _place: pn.places[from],
      weight: _weight})
}

func Init(_id int) *PetriNet {
  return &PetriNet{
    id: _id,
    places: make(map[int]*place),
    transitions: make(map[int]*Transition)}
}

func Build() *PetriNet{
  /*fpt := []Arc{Arc{p: 1, t: 1, w: 1}, Arc{p: 2, t: 2, w: 1}, Arc{p: 3, t: 2, w: 1}}
	ftp := []Arc{Arc{t: 1, p: 2, w: 1}, Arc{t: 1, p: 3, w: 1}, Arc{t: 2, p: 4, w: 1}}
	m := make(map[int]int)
	inhi := []Arc{Arc{t: 2, p: 4, w:1}}
	m[1] = 1
	m[3] = 2
	m[4] = 1
	p := InitPetriNet(4, 2, fpt, ftp, inhi, m)
  fmt.Printf("%v", p)
  p.Run()
  */

  p := Init(1)
  p.AddPlace(1, 1, "")
  p.AddPlace(2, 1, "")
  p.AddPlace(3, 2, "")
  p.AddPlace(4, 1, "")
  p.AddTransition(1,1)
  p.AddTransition(2,0)
  p.AddInArc(1,1,1)
  p.AddInArc(2,2,1)
  p.AddInArc(3,2,1)
  p.AddOutArc(1,2,1)
  p.AddOutArc(1,3,1)
  p.AddOutArc(2,4,1)
  //p.addInhibitorArc(4,2,1)
  //fmt.Printf("%v\n", p)
  return p
}
/*
Hacer ejercicio de mutual exclution distribuido
que pasa si se conecta en medio de elegir y fire
Definir las reglas
ext 1 -> distribuido y experimentos (con trans normales, despues reactivas) (mail para conectar nodos)
ext 2 -> reacciones reactivas (prioridades)
*/
