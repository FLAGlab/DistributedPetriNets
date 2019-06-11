package petribuilder

import (
  "os"
  "testing"
)

const TEST_CDR string = "test"

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
  // TODO: check that each set of rmt transitions is as expected
}

func TestUpdatePetriNetWithCDR(t *testing.T) {
  // a and b exclusion
  a := CreateContext("a")
  b := CreateContext("b")
  // c and d causality
  c := CreateContext("c")
  d := CreateContext("d")
  // e and f implication
  e := CreateContext("e")
  f := CreateContext("f")
  // g and h requirement
  g := CreateContext("g")
  h := CreateContext("h")
  // i and j suggestion
  i := CreateContext("i")
  j := CreateContext("j")
  UpdatePetriNetWithCDR(a, TEST_CDR)
  UpdatePetriNetWithCDR(b, TEST_CDR)
  UpdatePetriNetWithCDR(c, TEST_CDR)
  UpdatePetriNetWithCDR(d, TEST_CDR)
  UpdatePetriNetWithCDR(e, TEST_CDR)
  UpdatePetriNetWithCDR(f, TEST_CDR)
  UpdatePetriNetWithCDR(g, TEST_CDR)
  UpdatePetriNetWithCDR(h, TEST_CDR)
  UpdatePetriNetWithCDR(i, TEST_CDR)
  UpdatePetriNetWithCDR(j, TEST_CDR)
  // TODO: check that each set of rmt transitions is as expected
}

func TestUpdateConflictSolverWithCDR(t *testing.T) {
  // TODO: complete test
}
