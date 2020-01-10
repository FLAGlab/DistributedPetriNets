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

// OperationType operation being executed on a place
type OperationType int

const (
	//ADDITION of tokens
	ADDITION OperationType = 0
	//SUBSTRACTION of tokens
	SUBSTRACTION OperationType = 1
)

// PetriNet struct, has an id, transitions and places
type PetriNet struct {
	ID           int
	Transitions  map[int]*Transition
	Places       map[int]*Place
	MaxPriority  int
	MarksHistory []map[int]int
}

func (pn PetriNet) String() string {
	s := ""
	keys := make([]int, len(pn.Places))
	i := 0
	for k := range pn.Places {
		keys[i] = k
		i++
	}
	sort.Ints(keys)
	for _, k := range keys {
		s = fmt.Sprintf("%v\n%v", s, pn.Places[k])
	}
	return s + "\n"
}

func (pn *PetriNet) getPlace(id int) *Place {
	return pn.Places[id]
}

func (pn *PetriNet) updatePriority(transitionID, priority int) {
	pn.MaxPriority = -1
	pn.Transitions[transitionID].Priority = priority
}

func (pn *PetriNet) rollBack() error {
	if len(pn.MarksHistory) > 0 {
		currState := pn.MarksHistory[len(pn.MarksHistory)-1]
		pn.MarksHistory = pn.MarksHistory[:len(pn.MarksHistory)-1]
		for idPlace, mark := range currState {
			pn.Places[idPlace].Marks = mark
		}
		return nil
	}

	return errors.New("Invalid initial state")
}

func (pn *PetriNet) getCurrentState() (bool, map[int]int) {
	ans := make(map[int]int)
	for id, place := range pn.Places {
		ans[id] = place.Marks
	}
	return true, ans
}

func (pn *PetriNet) saveHistory() {
	must, state := pn.getCurrentState()
	if must {
		pn.MarksHistory = append(pn.MarksHistory, state)
	}
}

func (pn *PetriNet) Run() {
	for {
		for _, t := range pn.Transitions {
			t.Fire()
		}
		fmt.Printf("%v \n", pn)
	}
}

// FireTransitionByID fires a transition given its ID
func (pn *PetriNet) FireTransitionByID(transitionID int) error {
	pn.saveHistory()
	return pn.Transitions[transitionID].Fire()
}

func (pn *PetriNet) getTransitionOptionsByPriority(priority int) []*Transition {
	priorityOptions := make([]*Transition, 0)
	for _, transition := range pn.Transitions {
		if transition.Priority == priority && transition.CanFire() {
			priorityOptions = append(priorityOptions, transition)
		}
	}
	return priorityOptions
}

// GetTransitionOptions gets all the transitions with min priority that can be
// fired with a map from transition ID to RemoteTransition
func (pn *PetriNet) getTransitionOptions() []*Transition {
	var transitionOptions []*Transition
	currMin := math.MaxInt64
	for _, currTransition := range pn.Transitions {
		if currTransition.CanFire() {
			if currTransition.Priority < currMin {
				currMin = currTransition.Priority
				transitionOptions = []*Transition{currTransition}
			} else if currTransition.Priority == currMin {
				transitionOptions = append(transitionOptions, currTransition)
			}
		}
	}
	return transitionOptions
}

func (pn *PetriNet) AddPlace(_id, _marks int, _label string) {
	pn.Places[_id] = &Place{ID: _id, Marks: _marks, Label: _label}
}

func (pn *PetriNet) AddTransition(_id, _priority int) {
	if _priority > pn.MaxPriority {
		pn.MaxPriority = _priority
	}
	pn.Transitions[_id] = &Transition{
		ID:       _id,
		Priority: _priority,
		InArcs:   make([]Arc, 0),
		OutArcs:  make([]Arc, 0),
	}
}

func (pn *PetriNet) AddInArc(from, _transition, _weight int) {
	pn.Transitions[_transition].AddInArc(
		Arc{
			Place:  pn.Places[from],
			Weight: _weight,
		})
}

func (pn *PetriNet) AddOutArc(_transition, to, _weight int) {
	pn.Transitions[_transition].AddOutArc(
		Arc{
			Place:  pn.Places[to],
			Weight: _weight})
}

//AddRemoteOutArc adds a remote arc
func (pn *PetriNet) AddRemoteOutArc(_transition, weight int, serviceName string) {
	pn.Transitions[_transition].AddRemoteOutArc(
		RemoteArc{
			ServiceName: serviceName,
			Weight:  weight,
		})
}

//InitService
func (pn *PetriNet) InitService() {
	for i := range pn.Places {
		go pn.Places[i].InitService(pn.Places[i].Label)
	}
}

//InitPN Initializes a new Petri net
func InitPN(_id int) *PetriNet {
	return &PetriNet{
		ID:          _id,
		Places:      make(map[int]*Place),
		MaxPriority: -1,
		Transitions: make(map[int]*Transition),
	}
}

func (pn *PetriNet) getMaxPriority() int {
	if pn.MaxPriority == -1 {
		for _, tr := range pn.Transitions {
			if pn.MaxPriority < tr.Priority {
				pn.MaxPriority = tr.Priority
			}
		}
	}
	return pn.MaxPriority
}

/*
Hacer ejercicio de mutual exclution distribuido
que pasa si se conecta en medio de elegir y fire
Definir las reglas
ext 1 -> distribuido y experimentos (con trans normales, despues reactivas) (mail para conectar nodos)
ext 2 -> reacciones reactivas (prioridades)
*/
