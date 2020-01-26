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
	"math/rand"
	"sort"
	"time"
)

// PetriNet struct, has an id, transitions and places
type PetriNet struct {
	ID          int
	Transitions map[int]*Transition
	Places      map[int]*Place
	MaxPriority int
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
	/*for _, k := range keys {
		s = fmt.Sprintf("%v\n%v", s, pn.Places[k])
	}*/
	return s + "\n"
}

func (pn *PetriNet) AddPlace(_id int, _label, _nombre string) {
	pn.Places[_id] = &Place{ID: _id, Label: _label, Nombre: _nombre}
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
			Weight:      weight,
		})
}

func (pn *PetriNet) run() {
	for {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		for _, i := range r.Perm(len(pn.Transitions)) {
			i = i + 1
			pn.Transitions[i].Fire()
		}
		time.Sleep(10 * time.Second)
	}

}

//InitService
func (pn *PetriNet) InitService(interf string) {
	for i := range pn.Places {
		go pn.Places[i].InitService(pn.Places[i].Label, interf)
	}
	time.Sleep(6 * time.Second)
	pn.run()
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

/*
Hacer ejercicio de mutual exclution distribuido
que pasa si se conecta en medio de elegir y fire
Definir las reglas
ext 1 -> distribuido y experimentos (con trans normales, despues reactivas) (mail para conectar nodos)
ext 2 -> reacciones reactivas (prioridades)
*/
