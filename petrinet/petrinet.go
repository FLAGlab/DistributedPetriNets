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
  "errors"
  "fmt"
  "math"
  "sort"
)

type OperationType int

const (
  ADDITION     OperationType = 0
  SUBSTRACTION OperationType = 1
)

// PetriNet struct, has an id, transitions and places
type PetriNet struct {
 id int
 transitions map[int]*Transition
 places map[int]*Place
 maxPriority int
 marksHistory [] map[int]int
}

func (pn PetriNet) String() string {
  s := ""
  keys := make([]int, len(pn.places))
  i := 0
  for k := range pn.places {
      keys[i] = k
      i++
  }
  sort.Ints(keys)
  for _, k := range keys {
    s = fmt.Sprintf("%v\n%v", s, pn.places[k])
  }
  return s + "\n"
}

func (pn *PetriNet) GetPlace(id int) *Place {
  return pn.places[id]
}

func (pn *PetriNet) UpdatePriority(transitionID, priority int) {
  pn.maxPriority = -1
  pn.transitions[transitionID].priority = priority
}

func (pn *PetriNet) RollBack() error {
  if len(pn.marksHistory) > 0 {
    currState := pn.marksHistory[len(pn.marksHistory) - 1]
    pn.marksHistory = pn.marksHistory[:len(pn.marksHistory) - 1]
    for idPlace, mark := range currState {
      pn.places[idPlace].marks=mark
    }
    return nil
  }

  return errors.New("Invalid initial state")

}

func (pn *PetriNet) getCurrentState() (bool, map[int]int) {
  ans := make(map[int]int)
  for id, place := range pn.places {
    ans[id] = place.marks
  }
  return true, ans
}

func (pn *PetriNet) saveHistory() {
  must, state := pn.getCurrentState()
  if must {
    pn.marksHistory = append(pn.marksHistory, state)
  }
}
// FireTransitionByID fires a transition given its ID
func (pn *PetriNet) FireTransitionByID(transitionID int) error {
  pn.saveHistory()
  return pn.transitions[transitionID].fire()
}

func (pn *PetriNet) CopyPlaceMarksToRemoteArc(remoteArcs []*RemoteArc) {
  for i, rmtArc := range remoteArcs {
    remoteArcs[i].marks = pn.places[rmtArc.placeID].marks
  }
}

// AddMarksToPlaces adds weight (pos or neg) to specified places
func (pn *PetriNet) AddMarksToPlaces(opType OperationType, remoteArcs []*RemoteArc, saveHistory bool) {
  if saveHistory {
    pn.saveHistory()
  }
  for _, rmtArc := range remoteArcs {
    toAdd := rmtArc.weight
    if opType == SUBSTRACTION {
      toAdd = -toAdd
    }
    pn.places[rmtArc.placeID].marks += toAdd
  }
}

func (pn *PetriNet) GetTransitionOptionsByPriority(priority int) ([]*Transition) {
  priorityOptions := make([]*Transition, 0)
  for _, transition := range pn.transitions {
    if transition.priority == priority && transition.canFire() {
      priorityOptions = append(priorityOptions, transition)
    }
  }
  return priorityOptions
}

// GetTransitionOptions gets all the transitions with min priority that can be
// fired with a map from transition ID to RemoteTransition
func (pn *PetriNet) GetTransitionOptions() ([]*Transition) {
  var transitionOptions []*Transition
  currMin := math.MaxInt64
  for _, currTransition := range pn.transitions {
    if (currTransition.canFire()) {
      if (currTransition.priority < currMin) {
        currMin = currTransition.priority
        transitionOptions = []*Transition{currTransition}
      } else if (currTransition.priority == currMin) {
        transitionOptions = append(transitionOptions, currTransition)
      }
    }
  }
  return transitionOptions
}

func (pn *PetriNet) AddPlace(_id, _marks int, _label string) {
  pn.places[_id] = &Place{ID: _id, marks: _marks, label: _label}
}

func (pn *PetriNet) AddTransition(_id, _priority int) {
  if _priority > pn.maxPriority {
    pn.maxPriority = _priority
  }
  pn.transitions[_id] = &Transition {
    ID: _id,
    priority: _priority,
    inArcs: make([]arc,0),
    outArcs: make([]arc,0),
  }
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

func (pn *PetriNet) AddRemoteOutArc(_transition, to, weight int) {
  pn.transitions[_transition].addRemoteOutArc(
    RemoteArc {
      placeID: to,
      weight: weight})
}

func Init(_id int) *PetriNet {
  return &PetriNet{
    id: _id,
    places: make(map[int]*Place),
    maxPriority: -1,
    transitions: make(map[int]*Transition),
  }
}

func (pn *PetriNet) GetMaxPriority() int {
  if pn.maxPriority == -1 {
    for _, tr := range pn.transitions {
      if pn.maxPriority < tr.priority {
        pn.maxPriority = tr.priority
      }
    }
  }
  return pn.maxPriority
}

/*
Hacer ejercicio de mutual exclution distribuido
que pasa si se conecta en medio de elegir y fire
Definir las reglas
ext 1 -> distribuido y experimentos (con trans normales, despues reactivas) (mail para conectar nodos)
ext 2 -> reacciones reactivas (prioridades)
*/
