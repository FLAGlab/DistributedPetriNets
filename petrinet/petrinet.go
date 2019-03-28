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
)

type OperationType int

const (
  ADDITION     OperationType = 0
  SUBSTRACTION OperationType = 1
)

// PetriNet struct, has an id, transitions and places
type PetriNet struct {
 id int
 Context string
 transitions map[int]*Transition
 places map[int]*Place
 remoteTransitions map[int]*RemoteTransition
}

func (pn PetriNet) String() string {
  s := ""
  for _, _place := range pn.places{
    s = fmt.Sprintf("%v\n%v", s, _place)
  }
  return s + "\n"
}

// FireTransitionByID fires a transition given its ID
func (pn *PetriNet) FireTransitionByID(transitionID int) error {
  return pn.transitions[transitionID].fire()
}

func (pn *PetriNet) CopyPlaceMarksToRemoteArc(remoteArcs []*RemoteArc) {
  fmt.Println("WILL COPY PLACE MARKS TO REMOTE ARC METH")
  fmt.Printf("BEFORE: %v\n", remoteArcs)
  for i, rmtArc := range remoteArcs {
    fmt.Printf("CURRENT PLACE: %v, PLACE MARKS: %v\n", rmtArc.PlaceID, pn.places[rmtArc.PlaceID].marks)
    fmt.Printf("CURR: %v\n", remoteArcs[i])
    remoteArcs[i].Marks = pn.places[rmtArc.PlaceID].marks
    fmt.Printf("CURR AFTER: %v\n", remoteArcs[i])
  }
  fmt.Printf("AFTER: %v\n", remoteArcs)
}

// AddMarksToPlaces adds weight (pos or neg) to specified places
func (pn *PetriNet) AddMarksToPlaces(opType OperationType, remoteArcs []*RemoteArc) {
  fmt.Println("WILL ADD MARKS TO PLACES")
  fmt.Printf("OLD MARKS: %v\n", pn)
  for _, rmtArc := range remoteArcs {
    toAdd := rmtArc.Weight
    if opType == SUBSTRACTION {
      toAdd = -toAdd
    }
    fmt.Printf("WILL ADD: %v\n", toAdd)
    pn.places[rmtArc.PlaceID].marks += toAdd
  }
  fmt.Printf("NEW MARKS: %v\n", pn)
}

// GetTransitionOptions gets all the transitions with min priority that can be
// fired with a map from transition ID to RemoteTransition
func (pn *PetriNet) GetTransitionOptions() ([]*Transition, map[int]*RemoteTransition) {
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
  remoteTransitions := make(map[int]*RemoteTransition)
  for _, option := range transitionOptions {
    if pn.remoteTransitions[option.ID] != nil {
      remoteTransitions[option.ID] = pn.remoteTransitions[option.ID]
    }
  }
  return transitionOptions, remoteTransitions
}

func (pn *PetriNet) AddPlace(_id, _marks int, _label string) {
  pn.places[_id] = &Place{ID: _id, marks: _marks, label: _label}
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
      place: pn.places[from],
      weight: _weight})
}
func (pn *PetriNet) AddOutArc(_transition, to, _weight int) {
  pn.transitions[_transition].addOutArc(
    arc {
      place: pn.places[to],
      weight: _weight})
}
func (pn *PetriNet) AddInhibitorArc(from,_transition,_weight int) {
  pn.transitions[_transition].addInhibitorArc(
    arc {
      place: pn.places[from],
      weight: _weight})
}

func (pn *PetriNet) AddRemoteTransition(_id int) {
  fmt.Println("will add remote transition")
  fmt.Println(_id)
  pn.remoteTransitions[_id] = &RemoteTransition {
    ID: _id,
    InArcs: make([]RemoteArc,0),
    OutArcs: make([]RemoteArc,0),
    InhibitorArcs: make([]RemoteArc,0)}
}

func (pn *PetriNet) AddRemoteInArc(from,_transition, weight int, context string) {
  fmt.Println("will add remote in arc")
  fmt.Printf("%d %d %v\n", from, _transition, pn.remoteTransitions[_transition])
  pn.remoteTransitions[_transition].addInArc(
    RemoteArc {
      PlaceID: from,
      Context: context,
      Weight: weight})
}
func (pn *PetriNet) AddRemoteOutArc(_transition, to, weight int, context string) {
  pn.remoteTransitions[_transition].addOutArc(
    RemoteArc {
      PlaceID: to,
      Context: context,
      Weight: weight})
}
func (pn *PetriNet) AddRemoteInhibitorArc(from,_transition, weight int, context string) {
  pn.remoteTransitions[_transition].addInhibitorArc(
    RemoteArc {
      PlaceID: from,
      Context: context,
      Weight: weight})
}

func Init(_id int, context string) *PetriNet {
  return &PetriNet{
    id: _id,
    places: make(map[int]*Place),
    Context: context,
    transitions: make(map[int]*Transition),
    remoteTransitions: make(map[int]*RemoteTransition)}
}

/*
Hacer ejercicio de mutual exclution distribuido
que pasa si se conecta en medio de elegir y fire
Definir las reglas
ext 1 -> distribuido y experimentos (con trans normales, despues reactivas) (mail para conectar nodos)
ext 2 -> reacciones reactivas (prioridades)
*/
