package petribuilder

import (
  "os"
  "reflect"
  "testing"

  "github.com/FLAGlab/DCoPN/petrinet"
  "github.com/FLAGlab/DCoPN/conflictsolver"
)

const TEST_CDR string = "test"

func removeNoOrder(s []petrinet.RemoteTransition, i int) []petrinet.RemoteTransition {
  s[len(s)-1], s[i] = s[i], s[len(s)-1]
  return s[:len(s)-1]
}

func TestMain(m *testing.M) {
    cachedCdr = make(map[string]cdr)
    cachedCdr[TEST_CDR] = cdr {
      exclusion: []relation{
        relation{"a", "b"},
      },
      causality: []relation{
        relation{"c", "d"},
      },
      implication: []relation{
        relation{"e", "f"},
      },
      requirement: []relation{
        relation{"g", "h"},
      },
      suggestion: []relation{
        relation{"i", "j"},
      },
    }
    code := m.Run()
    os.Exit(code)
}

func TestCreateContext(t *testing.T) {
  ctx := "test"
  p := CreateContext(ctx)
  if p.Context != ctx {
    t.Errorf("Expected the context to be %v but was %v", ctx, p.Context)
  }
  expected := map[int]int {
    PR_PLACE: 0,
    CTX_PLACE: 0,
    PR_NOT_PLACE: 0,
  }
  checkFunc := func() {
    for place, marks := range expected {
      currMarks := p.GetPlace(place).GetMarks()
      if currMarks != marks {
        t.Errorf("Expected place %v to have %v marks but had %v", place, marks, currMarks)
      }
    }
  }
  checkFunc()
  p.FireTransitionByID(REQ_TR)
  expected[PR_PLACE] = 1
  checkFunc()

  p.FireTransitionByID(ACT_TR)
  expected[PR_PLACE] = 0
  expected[CTX_PLACE] = 1
  checkFunc()

  p.FireTransitionByID(REQ_NOT_TR)
  expected[PR_NOT_PLACE] = 1
  checkFunc()

  p.FireTransitionByID(DEACT_TR)
  expected[PR_NOT_PLACE] = 0
  expected[CTX_PLACE] = 0
  checkFunc()
}

func TestGetUniversalPetriNet(t *testing.T) {
  upn := GetUniversalPetriNet(TEST_CDR)
  _, rmtInternal := upn.GetTransitionOptionsByPriority(INTERNAL_TRANSITION_PRIORITY)
  _, rmtExternal := upn.GetTransitionOptionsByPriority(EXTERNAL_TRANSITION_PRIORITY)
  _, rmtClear := upn.GetTransitionOptionsByPriority(CLEAR_TRANSITION_PRIORITY)
  if len(rmtInternal) != 5 {
    t.Errorf("Expected universal petri net to have %v internal transitions but had %v", 5, rmtInternal)
  }
  if len(rmtExternal) != 0 {
    t.Errorf("Expected universal petri net to have %v internal transitions but had %v", 0, rmtExternal)
  }
  if len(rmtClear) != 1 {
    t.Errorf("Expected universal petri net to have %v internal transitions but had %v", 1, rmtClear)
  }
}

func TestGetUniversalPetriNetCausality(t *testing.T) {
  upn := GetUniversalPetriNet(TEST_CDR)
  _, rmtInternal := upn.GetTransitionOptionsByPriority(INTERNAL_TRANSITION_PRIORITY)
  expectedInternal := []petrinet.RemoteTransition {
    petrinet.RemoteTransition { // from c d causality
      ID: 0,
      InArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{
          PlaceID: CTX_PLACE,
          Context: "c",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
        petrinet.RemoteArc{
          PlaceID: PR_NOT_PLACE,
          Context: "c",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
      },
      OutArcs: []petrinet.RemoteArc{},
      InhibitorArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{
          PlaceID: CTX_PLACE,
          Context: "d",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
      },
    },
  }
  helperTestGetUniversalPetriNet(rmtInternal, expectedInternal, t)
}

func TestGetUniversalPetriNetImplication(t *testing.T) {
  upn := GetUniversalPetriNet(TEST_CDR)
  _, rmtInternal := upn.GetTransitionOptionsByPriority(INTERNAL_TRANSITION_PRIORITY)
  expectedInternal := []petrinet.RemoteTransition {
    petrinet.RemoteTransition { // from e f implication
      ID: 0,
      InArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{
          PlaceID: CTX_PLACE,
          Context: "e",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
      },
      OutArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{
          PlaceID: CTX_PLACE,
          Context: "e",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
        petrinet.RemoteArc{
          PlaceID: PR_NOT_PLACE,
          Context: "e",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
      },
      InhibitorArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{
          PlaceID: PR_NOT_PLACE,
          Context: "e",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
        petrinet.RemoteArc{
          PlaceID: PR_PLACE,
          Context: "f",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
        petrinet.RemoteArc{
          PlaceID: CTX_PLACE,
          Context: "f",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
      },
    },
    petrinet.RemoteTransition { // from e f implication (2)
      ID: 0,
      InArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{
          PlaceID: PR_NOT_PLACE,
          Context: "f",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
      },
      OutArcs: []petrinet.RemoteArc{},
      InhibitorArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{
          PlaceID: CTX_PLACE,
          Context: "f",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
      },
    },
  }
  helperTestGetUniversalPetriNet(rmtInternal, expectedInternal, t)
}

func TestGetUniversalPetriNetRequirement(t *testing.T) {
  upn := GetUniversalPetriNet(TEST_CDR)
  _, rmtInternal := upn.GetTransitionOptionsByPriority(INTERNAL_TRANSITION_PRIORITY)
  expectedInternal := []petrinet.RemoteTransition {
    petrinet.RemoteTransition { // from g h requirement
      ID: 0,
      InArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{
          PlaceID: CTX_PLACE,
          Context: "g",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
      },
      OutArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{
          PlaceID: CTX_PLACE,
          Context: "g",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
        petrinet.RemoteArc{
          PlaceID: PR_NOT_PLACE,
          Context: "g",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
      },
      InhibitorArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{
          PlaceID: PR_NOT_PLACE,
          Context: "g",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
        petrinet.RemoteArc{
          PlaceID: CTX_PLACE,
          Context: "h",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
      },
    },
  }
  helperTestGetUniversalPetriNet(rmtInternal, expectedInternal, t)
}

func TestGetUniversalPetriNetSuggestion(t *testing.T) {
  upn := GetUniversalPetriNet(TEST_CDR)
  _, rmtInternal := upn.GetTransitionOptionsByPriority(INTERNAL_TRANSITION_PRIORITY)
  _, rmtClear := upn.GetTransitionOptionsByPriority(CLEAR_TRANSITION_PRIORITY)
  expectedInternal := []petrinet.RemoteTransition {
    petrinet.RemoteTransition { // from i suggest j
      ID: 0,
      InArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{
          PlaceID: CTX_PLACE,
          Context: "i",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
        petrinet.RemoteArc{
          PlaceID: PR_NOT_PLACE,
          Context: "i",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
      },
      OutArcs: []petrinet.RemoteArc{},
      InhibitorArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{
          PlaceID: CTX_PLACE,
          Context: "j",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
      },
    },
  }
  helperTestGetUniversalPetriNet(rmtInternal, expectedInternal, t)

  expectedClear := []petrinet.RemoteTransition {
    petrinet.RemoteTransition {
      ID: 0,
      InArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{
          PlaceID: PR_PLACE,
          Context: "j",
          Address: "",
          Weight: 1,
          Marks: 0,
        },
      },
      OutArcs: []petrinet.RemoteArc{},
      InhibitorArcs: []petrinet.RemoteArc{},
    },
  }
  helperTestGetUniversalPetriNet(rmtClear, expectedClear, t)
}

func helperTestGetUniversalPetriNet(
    obtainedTrsRef map[int]*petrinet.RemoteTransition,
    expectedTrs []petrinet.RemoteTransition,
    t *testing.T) {
  obtainedTrs := make(map[int]petrinet.RemoteTransition)
  for k, v := range obtainedTrsRef {
    obtainedTrs[k] = *v
  }
  for id, tr := range obtainedTrs {
    foundIndex := -1
    for index, expectedTr := range expectedTrs {
      expectedTr.ID = id
      if reflect.DeepEqual(expectedTr, tr) {
        foundIndex = index
        break
      }
    }
    if foundIndex != -1 {
      expectedTrs = removeNoOrder(expectedTrs, foundIndex)
    }
  }
  if len(expectedTrs) > 0 {
    t.Errorf("Expected to find %v but didn't. Obtained map: %v", expectedTrs, obtainedTrs)
  }
}

func helperUpdatePetriNetWithCDRfunc(pn *petrinet.PetriNet, expectedInternal map[int]petrinet.RemoteTransition, t *testing.T) {
  // make sure every transition can fire
  pn.FireTransitionByID(REQ_TR)
  pn.FireTransitionByID(REQ_TR)
  pn.FireTransitionByID(ACT_TR)
  pn.FireTransitionByID(REQ_NOT_TR)

  _, rmtInternalRef := pn.GetTransitionOptionsByPriority(INTERNAL_TRANSITION_PRIORITY)
  _, rmtExternalRef := pn.GetTransitionOptionsByPriority(EXTERNAL_TRANSITION_PRIORITY)
  _, rmtClearRef := pn.GetTransitionOptionsByPriority(CLEAR_TRANSITION_PRIORITY)
  rmtInternal := make(map[int]petrinet.RemoteTransition)
  for id, rmt := range rmtInternalRef {
    rmtInternal[id] = *rmt
  }
  rmtExternal := make(map[int]petrinet.RemoteTransition)
  for id, rmt := range rmtExternalRef {
    rmtExternal[id] = *rmt
  }
  rmtClear := make(map[int]petrinet.RemoteTransition)
  for id, rmt := range rmtClearRef {
    rmtClear[id] = *rmt
  }
  t.Logf("Rmt expected internal: %v", expectedInternal)
  t.Logf("Rmt internal: %v", rmtInternal)
  t.Logf("Rmt external: %v", rmtExternal)
  t.Logf("Rmt clear: %v", rmtClear)
  if len(rmtExternal) != 0 {
    t.Errorf("Expected no remote external transitions on context %v but had %v", pn.Context, rmtExternal)
  }
  if len(rmtClear) != 0 {
    t.Errorf("Expected no remote clear transitions on context %v but had %v", pn.Context, rmtExternal)
  }
  if len(rmtInternal) != len(expectedInternal) {
    t.Errorf("Expected %v remote internal transitions but were %v", len(expectedInternal), len(rmtInternal))
  }
  for id, expectedTr := range expectedInternal {
    t.Logf("Current transition id: %v", id)
    t.Logf("Map value of that id in expected internal: %v", expectedInternal[id])
    t.Logf("Map value of that id in existing internal: %v", rmtInternal[id])
    existing, exists := rmtInternal[id]
    if !exists {
      t.Errorf("Expect universal petri net to have internal transition %v but didn't exist", expectedTr)
    } else if !reflect.DeepEqual(expectedTr, existing) {
      t.Errorf("Expect universal petri net to have internal transition %v but was %v", expectedTr, existing)
    }
  }
}

func TestUpdatePetriNetWithCDRExclusion(t *testing.T) {
  // a and b exclusion
  a := CreateContext("a")
  b := CreateContext("b")
  UpdatePetriNetWithCDR(a, TEST_CDR)
  UpdatePetriNetWithCDR(b, TEST_CDR)
  expected := map[int]petrinet.RemoteTransition {
    ACT_TR: petrinet.RemoteTransition{
      ID: ACT_TR,
      InArcs: []petrinet.RemoteArc{},
      OutArcs: []petrinet.RemoteArc{},
      InhibitorArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{PlaceID: CTX_PLACE,Context: "b",Address: "",Weight: 1,Marks: 0},
      },
    },
  }
  helperUpdatePetriNetWithCDRfunc(a, expected, t)

  expected = map[int]petrinet.RemoteTransition {
    ACT_TR: petrinet.RemoteTransition{
      ID: ACT_TR,
      InArcs: []petrinet.RemoteArc{},
      OutArcs: []petrinet.RemoteArc{},
      InhibitorArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{PlaceID: CTX_PLACE,Context: "a",Address: "",Weight: 1,Marks: 0},
      },
    },
  }
  helperUpdatePetriNetWithCDRfunc(b, expected, t)
}

func TestUpdatePetriNetWithCDRCausality(t *testing.T) {
  // c and d causality
  c := CreateContext("c")
  d := CreateContext("d")
  UpdatePetriNetWithCDR(c, TEST_CDR)
  UpdatePetriNetWithCDR(d, TEST_CDR)
  expected := map[int]petrinet.RemoteTransition {
    ACT_TR: petrinet.RemoteTransition{
      ID: ACT_TR,
      InArcs: []petrinet.RemoteArc{},
      OutArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{PlaceID: PR_PLACE,Context: "d",Address: "",Weight: 1,Marks: 0},
      },
      InhibitorArcs: []petrinet.RemoteArc{},
    },
    DEACT_TR: petrinet.RemoteTransition{
      ID: DEACT_TR,
      InArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{PlaceID: CTX_PLACE,Context: "d",Address: "",Weight: 1,Marks: 0},
      },
      OutArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{PlaceID: PR_NOT_PLACE,Context: "d",Address: "",Weight: 1,Marks: 0},
        petrinet.RemoteArc{PlaceID: CTX_PLACE,Context: "d",Address: "",Weight: 1,Marks: 0},
      },
      InhibitorArcs: []petrinet.RemoteArc{},
    },
  }
  helperUpdatePetriNetWithCDRfunc(c, expected, t)

  expected = map[int]petrinet.RemoteTransition {}
  helperUpdatePetriNetWithCDRfunc(d, expected, t)
}

func TestUpdatePetriNetWithCDRImplication(t *testing.T) {
  // e and f implication
  e := CreateContext("e")
  f := CreateContext("f")
  UpdatePetriNetWithCDR(e, TEST_CDR)
  UpdatePetriNetWithCDR(f, TEST_CDR)

  expected := map[int]petrinet.RemoteTransition {
    ACT_TR: petrinet.RemoteTransition{
      ID: ACT_TR,
      InArcs: []petrinet.RemoteArc{},
      OutArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{PlaceID: PR_PLACE,Context: "f",Address: "",Weight: 1,Marks: 0},
      },
      InhibitorArcs: []petrinet.RemoteArc{},
    },
    DEACT_TR: petrinet.RemoteTransition{
      ID: DEACT_TR,
      InArcs: []petrinet.RemoteArc{},
      OutArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{PlaceID: PR_NOT_PLACE,Context: "f",Address: "",Weight: 1,Marks: 0},
      },
      InhibitorArcs: []petrinet.RemoteArc{},
    },
  }
  helperUpdatePetriNetWithCDRfunc(e, expected, t)

  expected = map[int]petrinet.RemoteTransition {}
  helperUpdatePetriNetWithCDRfunc(f, expected, t)
}

func TestUpdatePetriNetWithCDRRequirement(t *testing.T) {
  // g and h requirement
  g := CreateContext("g")
  h := CreateContext("h")
  UpdatePetriNetWithCDR(g, TEST_CDR)
  UpdatePetriNetWithCDR(h, TEST_CDR)

  expected := map[int]petrinet.RemoteTransition {
    ACT_TR: petrinet.RemoteTransition{
      ID: ACT_TR,
      InArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{PlaceID: CTX_PLACE,Context: "h",Address: "",Weight: 1,Marks: 0},
      },
      OutArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{PlaceID: CTX_PLACE,Context: "h",Address: "",Weight: 1,Marks: 0},
      },
      InhibitorArcs: []petrinet.RemoteArc{},
    },
  }
  helperUpdatePetriNetWithCDRfunc(g, expected, t)

  expected = map[int]petrinet.RemoteTransition {}
  helperUpdatePetriNetWithCDRfunc(h, expected, t)
}

func TestUpdatePetriNetWithCDRSuggestion(t *testing.T) {
  // i and j suggestion
  i := CreateContext("i")
  j := CreateContext("j")
  UpdatePetriNetWithCDR(i, TEST_CDR)
  UpdatePetriNetWithCDR(j, TEST_CDR)

  expected := map[int]petrinet.RemoteTransition {
    ACT_TR: petrinet.RemoteTransition{
      ID: ACT_TR,
      InArcs: []petrinet.RemoteArc{},
      OutArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{PlaceID: PR_PLACE,Context: "j",Address: "",Weight: 1,Marks: 0},
      },
      InhibitorArcs: []petrinet.RemoteArc{},
    },
    DEACT_TR: petrinet.RemoteTransition{
      ID: DEACT_TR,
      InArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{PlaceID: CTX_PLACE,Context: "j",Address: "",Weight: 1,Marks: 0},
      },
      OutArcs: []petrinet.RemoteArc{
        petrinet.RemoteArc{PlaceID: PR_NOT_PLACE,Context: "j",Address: "",Weight: 1,Marks: 0},
        petrinet.RemoteArc{PlaceID: CTX_PLACE,Context: "j",Address: "",Weight: 1,Marks: 0},
      },
      InhibitorArcs: []petrinet.RemoteArc{},
    },
  }
  helperUpdatePetriNetWithCDRfunc(i, expected, t)

  expected = make(map[int]petrinet.RemoteTransition)
  helperUpdatePetriNetWithCDRfunc(j, expected, t)
}

func TestUpdateConflictSolverWithCDR(t *testing.T) {
  // TODO: complete test
  cs := conflictsolver.InitCS()
  UpdateConflictSolverWithCDR(&cs, TEST_CDR)
  expected := conflictsolver.InitCS()
  expected.AddConflict("a", "b", CTX_PLACE, CTX_PLACE, 1, 1, true, true)
  expected.AddConflict("e", "f", CTX_PLACE, CTX_PLACE, 1, 0, true, false)
  expected.AddConflict("g", "h", CTX_PLACE, CTX_PLACE, 1, 0, true, false)
  if !cs.Equals(&expected) {
    t.Errorf("Expected conflicts %v but was %v", expected, cs)
  }
}
