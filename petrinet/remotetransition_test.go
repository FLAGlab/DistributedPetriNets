package petrinet

import (
  "testing"
  "reflect"
)

func initTestRemoteTransition() *RemoteTransition {
  return &RemoteTransition{1, nil, nil, nil}
}

func slicesEqual(a, b []int) bool {
    if len(a) != len(b) {
        return false
    }
    for i, v := range a {
        if v != b[i] {
            return false
        }
    }
    return true
}

func sliceContains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func TestAddInArcRemoteTransition(t *testing.T) {
  rt := initTestRemoteTransition()
  rarc := RemoteArc{1,"test","",1,1}
  rt.addInArc(rarc)
  exists := false
  for _, item := range rt.InArcs {
    exists = exists || item == rarc
  }
  if !exists {
    t.Errorf("Couldn't add remote arc %v to remote transition %v.\n", rarc, rt)
  }
}

func TestAddOutArcRemoteTransition(t *testing.T) {
  rt := initTestRemoteTransition()
  rarc := RemoteArc{1,"test","",1,1}
  rt.addOutArc(rarc)
  exists := false
  for _, item := range rt.OutArcs {
    exists = exists || item == rarc
  }
  if !exists {
    t.Errorf("Couldn't add remote arc %v to remote transition %v.\n", rarc, rt)
  }
}

func TestAddInhibArcRemoteTransition(t *testing.T) {
  rt := initTestRemoteTransition()
  rarc := RemoteArc{1,"test","",1,1}
  rt.addInhibitorArc(rarc)
  exists := false
  for _, item := range rt.InhibitorArcs {
    exists = exists || item == rarc
  }
  if !exists {
    t.Errorf("Couldn't add remote arc %v to remote transition %v.\n", rarc, rt)
  }
}

func TestGetPlaceIDsByAddrs(t *testing.T) {
  rt := initTestRemoteTransition()
  rt.InArcs = []RemoteArc{
    {1, "", "127.0.0.1:3000", 1, 1},
    {1, "", "127.0.0.1:3001", 1, 1},
    {1, "", "127.0.0.1:3002", 1, 1},
    {2, "", "127.0.0.1:3002", 1, 1}}
  rt.OutArcs = []RemoteArc{
    {2, "", "127.0.0.1:3000", 1, 1},
    {3, "", "127.0.0.1:3000", 1, 1},
    {2, "", "127.0.0.1:3001", 1, 1},
    {3, "", "127.0.0.1:3002", 1, 1}}
  rt.InhibitorArcs = []RemoteArc{
    {4, "", "127.0.0.1:3000", 1, 1},
    {3, "", "127.0.0.1:3001", 1, 1},
    {4, "", "127.0.0.1:3002", 1, 1}}
  expected := make(map[string][]int)
  expected["127.0.0.1:3000"] = []int{1,2,3,4}
  expected["127.0.0.1:3001"] = []int{1,2,3}
  expected["127.0.0.1:3002"] = []int{1,2,3,4}
  ans := rt.GetPlaceIDsByAddrs()
  if len(ans) != len(expected) {
    t.Errorf("Answer has %v addresses but should have %v.", len(ans), len(expected))
  }
  for key, val := range expected {
    v, ok := ans[key]
    if !ok {
      t.Errorf("The address %v should exist.", key)
    } else if slicesEqual(v, val) {
      t.Errorf("Expected address %v to have values %v but had %v.", key, val, v)
    }
  }
}

func TestUpdateAddressByContext(t *testing.T) {
  rt := initTestRemoteTransition()
  rt.InArcs = []RemoteArc{
    {1, "ctx1", "", 1, 1},
    {1, "ctx2", "", 1, 1},
    {1, "ctx3", "", 1, 1},
    {2, "ctx3", "", 1, 1}}
  rt.OutArcs = []RemoteArc{
    {2, "ctx1", "", 1, 1},
    {3, "ctx1", "", 1, 1},
    {2, "ctx2", "", 1, 1},
    {3, "ctx3", "", 1, 1}}
  rt.InhibitorArcs = []RemoteArc{
    {4, "ctx1", "", 1, 1},
    {3, "ctx2", "", 1, 1},
    {4, "ctx3", "", 1, 1}}
  ctxToAddrs := make(map[string][]string)
  ctxToAddrs["ctx1"] = []string{"addr1", "addr2", "addr4"}
  ctxToAddrs["ctx2"] = []string{}
  ctxToAddrs["ctx3"] = []string{"addr3"}
  rt.UpdateAddressByContext(ctxToAddrs, "addr2")

  expectedInArc := make(map[string]map[string]int)
  expectedInArc["ctx1"] = make(map[string]int)
  expectedInArc["ctx1"]["addr1"] = 1
  expectedInArc["ctx1"]["addr4"] = 1
  expectedInArc["ctx3"] = make(map[string]int)
  expectedInArc["ctx3"]["addr3"] = 2

  expectedOutArc := make(map[string]map[string]int)
  expectedOutArc["ctx1"] = make(map[string]int)
  expectedOutArc["ctx1"]["addr1"] = 2
  expectedOutArc["ctx1"]["addr4"] = 2
  expectedOutArc["ctx3"] = make(map[string]int)
  expectedOutArc["ctx3"]["addr3"] = 1

  expectedInhibArc := make(map[string]map[string]int)
  expectedInhibArc["ctx1"] = make(map[string]int)
  expectedInhibArc["ctx1"]["addr1"] = 1
  expectedInhibArc["ctx1"]["addr4"] = 1
  expectedInhibArc["ctx3"] = make(map[string]int)
  expectedInhibArc["ctx3"]["addr3"] = 1
  // helper checks that all list remote arcs have correct address by
  // checking it from the expected map.
  helper := func (list []RemoteArc, expected map[string]map[string]int) {
    for _, val := range list {
      v, ok := expected[val.Context]
      if !ok {
        t.Errorf("Context %v should not exist or already used.", val.Context)
      } else if v2, ok2 := v[val.Address]; ok2 {
        v[val.Address] = v2 - 1
        if v[val.Address] == 0 {
          delete(v, val.Address)
          if len(v) == 0 {
            delete(expected, val.Context)
          }
        }
      } else {
        t.Errorf("Address %v should not exist or already used.", val.Address)
      }
    }
    if len(expected) != 0 {
      t.Errorf("Didn't use all addresses: %v", expected)
    }
  }
  helper(rt.InArcs, expectedInArc)
  helper(rt.OutArcs, expectedOutArc)
  helper(rt.InhibitorArcs, expectedInhibArc)
}

func TestGetInArcsByAddrs(t *testing.T) {
  rt := initTestRemoteTransition()
  addrToRemoteAddr := rt.GetInArcsByAddrs()
  if len(addrToRemoteAddr) != 0 {
    t.Errorf("Should have returned empty map %v", addrToRemoteAddr)
  }
  rt.InArcs = []RemoteArc{
    {1, "", "addr1", 1, 1},
    {1, "", "addr2", 1, 1},
    {1, "", "addr3", 1, 1},
    {2, "", "addr3", 1, 1}}
  addrToRemoteAddr = rt.GetInArcsByAddrs()
  expectedMap := make(map[string][]*RemoteArc)
  expectedMap["addr1"] = []*RemoteArc{{1, "", "addr1", 1, 1}}
  expectedMap["addr2"] = []*RemoteArc{{1, "", "addr2", 1, 1}}
  expectedMap["addr3"] = []*RemoteArc{{1, "", "addr3", 1, 1}, {2, "", "addr3", 1, 1}}
  eq := reflect.DeepEqual(addrToRemoteAddr, expectedMap)
  if !eq {
    t.Errorf("Expected %v but was %v", expectedMap, addrToRemoteAddr)
  }
}

func TestGetOutArcsByAddrs(t *testing.T) {
  rt := initTestRemoteTransition()
  addrToRemoteAddr := rt.GetOutArcsByAddrs()
  if len(addrToRemoteAddr) != 0 {
    t.Errorf("Should have returned empty map %v", addrToRemoteAddr)
  }
  rt.OutArcs = []RemoteArc{
    {1, "", "addr1", 1, 1},
    {1, "", "addr2", 1, 1},
    {1, "", "addr3", 1, 1},
    {2, "", "addr3", 1, 1}}
  addrToRemoteAddr = rt.GetOutArcsByAddrs()
  expectedMap := make(map[string][]*RemoteArc)
  expectedMap["addr1"] = []*RemoteArc{{1, "", "addr1", 1, 1}}
  expectedMap["addr2"] = []*RemoteArc{{1, "", "addr2", 1, 1}}
  expectedMap["addr3"] = []*RemoteArc{{1, "", "addr3", 1, 1}, {2, "", "addr3", 1, 1}}
  eq := reflect.DeepEqual(addrToRemoteAddr, expectedMap)
  if !eq {
    t.Errorf("Expected %v but was %v", expectedMap, addrToRemoteAddr)
  }
}
