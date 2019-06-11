package petribuilder

import (
  "encoding/json"
  "io/ioutil"

  "github.com/FLAGlab/DCoPN/petrinet"
  "github.com/FLAGlab/DCoPN/conflictsolver"
)

// [0 - 2] is the range of priorities from high(0) to low(2)
const (
  INTERNAL_TRANSITION_PRIORITY int = 0
  CLEAR_TRANSITION_PRIORITY int = 1
  EXTERNAL_TRANSITION_PRIORITY int = 2
  UNIVERSAL_PN string = "universal"

  REQ_TR int = 1
  ACT_TR int = 2
  REQ_NOT_TR int = 3
  DEACT_TR int = 4

  PR_PLACE int = 1
  CTX_PLACE int = 2
  PR_NOT_PLACE int = 3
)

type relation struct {
  a string `json: "A"`
  b string `json: "B"`
}

type cdr struct {
  exclusion []relation `json: "exclusion"`
  causality []relation `json: "causality"`
  implication []relation `json: "implication"`
  requirement []relation `json: "requirement"`
  suggestion []relation `json: "suggestion"`
}

var cachedCdr map[string]cdr

func getCdrStruct(cdrFile string) cdr {
  if cachedCdr == nil {
    cachedCdr = make(map[string]cdr)
  }
  cached, exists := cachedCdr[cdrFile]
  if !exists {
    file, _ := ioutil.ReadFile(cdrFile)
    data := cdr{}
    _ = json.Unmarshal([]byte(file), &data)
    cachedCdr[cdrFile] = data
    cached = data
  }
  return cached
}

func GetUniversalPetriNet(cdrFile string) *petrinet.PetriNet {
  relations := getCdrStruct(cdrFile)
  p := petrinet.Init(1, "universal")
  id := 1
  // for _, rel := range relations.exclusion {} -> nothing on exclusion
  for _, rel := range relations.causality {
    p.AddTransition(id, INTERNAL_TRANSITION_PRIORITY)
    p.AddRemoteTransition(id)
    p.AddRemoteInArc(CTX_PLACE, id, 1, rel.a)
    p.AddRemoteInArc(PR_NOT_PLACE, id, 1, rel.a)
    p.AddRemoteInhibitorArc(CTX_PLACE, id, 1, rel.b)
    id++
  }
  for _, rel := range relations.implication {
    p.AddTransition(id, INTERNAL_TRANSITION_PRIORITY)
    p.AddRemoteTransition(id)
    p.AddRemoteInArc(CTX_PLACE, id, 1, rel.a)
    p.AddRemoteOutArc(id, CTX_PLACE, 1, rel.a)
    p.AddRemoteOutArc(id, PR_NOT_PLACE, 1, rel.a)
    p.AddRemoteInhibitorArc(PR_NOT_PLACE, id, 1, rel.a)
    p.AddRemoteInhibitorArc(PR_PLACE, id, 1, rel.b)
    p.AddRemoteInhibitorArc(CTX_PLACE, id, 1, rel.b)
    id++
    p.AddTransition(id, INTERNAL_TRANSITION_PRIORITY)
    p.AddRemoteTransition(id)
    p.AddRemoteInArc(PR_NOT_PLACE, id, 1, rel.b)
    p.AddRemoteInhibitorArc(CTX_PLACE, id, 1, rel.b)
    id++
  }
  for _, rel := range relations.requirement {
    p.AddTransition(id, INTERNAL_TRANSITION_PRIORITY)
    p.AddRemoteTransition(id)
    p.AddRemoteInArc(CTX_PLACE, id, 1, rel.a)
    p.AddRemoteOutArc(id, CTX_PLACE, 1, rel.a)
    p.AddRemoteOutArc(id, PR_NOT_PLACE, 1, rel.a)
    p.AddRemoteInhibitorArc(PR_NOT_PLACE, id, 1, rel.a)
    p.AddRemoteInhibitorArc(CTX_PLACE, id, 1, rel.b)
    id++
  }
  for _, rel := range relations.suggestion {
    p.AddTransition(id, INTERNAL_TRANSITION_PRIORITY)
    p.AddRemoteTransition(id)
    p.AddRemoteInArc(CTX_PLACE, id, 1, rel.a)
    p.AddRemoteInArc(PR_NOT_PLACE, id, 1, rel.a)
    p.AddRemoteInhibitorArc(CTX_PLACE, id, 1, rel.b)
    id++
    p.AddTransition(id, CLEAR_TRANSITION_PRIORITY)
    p.AddRemoteTransition(id)
    p.AddRemoteInArc(PR_PLACE, id, 1, rel.b)
    id++
  }
  return p
}

func UpdatePetriNetWithCDR(pn *petrinet.PetriNet, cdrFile string) {
  relations := getCdrStruct(cdrFile)
  for _, rel := range relations.exclusion {
    if pn.Context == rel.a {
      pn.AddRemoteTransition(ACT_TR)
      pn.AddRemoteInhibitorArc(CTX_PLACE, ACT_TR, 1, rel.b)
    } else if pn.Context == rel.b {
      pn.AddRemoteTransition(ACT_TR)
      pn.AddRemoteInhibitorArc(CTX_PLACE, ACT_TR, 1, rel.a)
    }
  }
  for _, rel := range relations.causality {
    if pn.Context == rel.a {
      pn.AddRemoteTransition(ACT_TR)
      pn.AddRemoteOutArc(ACT_TR, PR_PLACE, 1, rel.b)
      pn.AddRemoteTransition(DEACT_TR)
      pn.AddRemoteOutArc(DEACT_TR, PR_NOT_PLACE, 1, rel.b)
      pn.AddRemoteOutArc(DEACT_TR, CTX_PLACE, 1, rel.b)
      pn.AddRemoteInArc(CTX_PLACE, DEACT_TR, 1, rel.b)
    } // no need to add arcs if pn.Context == rel.b
  }
  for _, rel := range relations.implication {
    if pn.Context == rel.a {
      pn.AddRemoteTransition(ACT_TR)
      pn.AddRemoteOutArc(ACT_TR, PR_PLACE, 1, rel.b)
      pn.AddRemoteTransition(DEACT_TR)
      pn.AddRemoteOutArc(DEACT_TR, PR_NOT_PLACE, 1, rel.b)
    } // no need to add arcs if pn.Context == rel.b
  }
  for _, rel := range relations.requirement {
    if pn.Context == rel.a {
      pn.AddRemoteTransition(ACT_TR)
      pn.AddRemoteOutArc(ACT_TR, CTX_PLACE, 1, rel.b)
      pn.AddRemoteInArc(CTX_PLACE, ACT_TR, 1, rel.b)
    } // no need to add arcs if pn.Context == rel.b
  }
  for _, rel := range relations.suggestion {
    if pn.Context == rel.a {
      pn.AddRemoteTransition(ACT_TR)
      pn.AddRemoteOutArc(ACT_TR, PR_PLACE, 1, rel.b)
      pn.AddRemoteTransition(DEACT_TR)
      pn.AddRemoteOutArc(DEACT_TR, PR_NOT_PLACE, 1, rel.b)
      pn.AddRemoteOutArc(DEACT_TR, CTX_PLACE, 1, rel.b)
      pn.AddRemoteInArc(CTX_PLACE, DEACT_TR, 1, rel.b)
    } // no need to add arcs if pn.Context == rel.b
  }
}

func UpdateConflictSolverWithCDR(cs *conflictsolver.ConflictSolver, cdrFile string) {
  // TODO: Complete function
  // relations := getCdrStruct(cdrFile)
  // for _, rel := range relations.exclusion {
  //
  // }
  // for _, rel := range relations.causality {
  //
  // }
  // for _, rel := range relations.implication {
  //
  // }
  // for _, rel := range relations.requirement {
  //
  // }
  // for _, rel := range relations.suggestion {
  //
  // }
}

func CreateContext(contextName string) *petrinet.PetriNet {
  p := petrinet.Init(1, contextName)
  p.AddPlace(PR_PLACE, 0, "Pr(" + contextName + ")")
  p.SetPlaceTemporal(PR_PLACE)
  p.AddPlace(CTX_PLACE, 0, contextName)
  p.AddPlace(PR_NOT_PLACE, 0, "Pr(¬" + contextName + ")")
  p.SetPlaceTemporal(PR_NOT_PLACE)
  p.AddTransition(REQ_TR, EXTERNAL_TRANSITION_PRIORITY) // req(contextName)
  p.AddTransition(ACT_TR, INTERNAL_TRANSITION_PRIORITY) // act(contextName)
  p.AddTransition(REQ_NOT_TR, EXTERNAL_TRANSITION_PRIORITY) // req(¬contextName)
  p.AddTransition(DEACT_TR, INTERNAL_TRANSITION_PRIORITY) // act(contextName)
  p.AddOutArc(REQ_TR, PR_PLACE, 1)
  p.AddInArc(PR_PLACE, ACT_TR, 1)
  p.AddOutArc(ACT_TR, CTX_PLACE, 1)
  p.AddOutArc(REQ_NOT_TR, PR_NOT_PLACE, 1)
  p.AddInArc(PR_NOT_PLACE, DEACT_TR, 1)
  p.AddInArc(CTX_PLACE, DEACT_TR, 1)
  return p
}
