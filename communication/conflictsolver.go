package communication

import (
	"fmt"

	"github.com/FLAGlab/DCoPN/petrinet"
)

type conflict struct {
	ctxA string
	ctxB string
	placeIdA int
	palceIdB int
}

type ConflictSolver struct {
	conflicts []conflict
}

func InitCS() ConflictSolver {
	return ConflictSolver{nil}
}

func (cs *ConflictSolver) AddConflict(ctxa, ctxb string, pa, pb int) {
	cs.conflicts = append(cs.conflicts, conflict{ctxa, ctxb, pa, pb})
}

func (cs *ConflictSolver) GetRequiredPlacesByAddress(ctx2address map[string][]string) map[string][]int {
	res := make(map[string][]int)
	for _,value := range cs.conflicts {
		ctxA := value.ctxA
		ctxB := value.ctxB
		pa := value.placeIdA
		pb := value.palceIdB
		if len(ctx2address[ctxA]) > 0 && len(ctx2address[ctxB]) > 0 {
			for _,address := range ctx2address[ctxA] {
				res[address] = append(res[address],pa)
			}
			for _,address := range ctx2address[ctxB] {
				res[address] = append(res[address],pb)
			}
		}
	}
	fmt.Printf("possible places %v\n",res)
	return res
}

func (cs *ConflictSolver) GetConflictedAddrs(marks map[string]map[int]*petrinet.RemoteArc, ctx2address map[string][]string) map[string] bool{
	res := make(map[string]bool)
	for _,value := range cs.conflicts {
		ctxA := value.ctxA
		ctxB := value.ctxB
		pa := value.placeIdA
		pb := value.palceIdB
		if len(ctx2address[ctxA]) > 0 && len(ctx2address[ctxB]) > 0 {
			for _,addressA := range ctx2address[ctxA] {
				marksA := marks[addressA][pa].Marks
				for _,addressB := range ctx2address[ctxB] {
					marksB := marks[addressB][pb].Marks
					if marksA > 0 && marksB > 0 {
						res[addressA] = true
						res[addressB] = true
					}
				}
			}
		}
	}
	return res
}
