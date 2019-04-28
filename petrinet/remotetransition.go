package petrinet

/*
Cases:
- Local transition to remote place (IN, inhib or normal)
  - Transition hability to fire must be checked by leader and locally
  - Fires with no problem locall
- Local transition to remote place (OUT, inhib or normal)
  - Transition hability to fire can be done locally
  - Fires with no problem locally, leader fires the remote if available
- Local transition to remote place (IN N' OUT, inhib or normal)
  - Transition hability to fire must be checked by leader and locally
  - Fires with no problem locally, leader fires the remote if available
- Blue transition from and to remote Place (different nodes, inhib or normal)
  - Transition hability to fire must be checked by leader
  - Fires with no problem remotely by leader
- Blue transition from and to remote Place (same node but requiring X node to be connected, inhib or normal)
*/
// RemoteTransition of a PetriNet
type RemoteTransition struct {
	ID int
	InArcs []RemoteArc
	OutArcs []RemoteArc
	InhibitorArcs []RemoteArc
}

func convertToAddressArcsMap(raList []RemoteArc) map[string][]*RemoteArc {
	ans := make(map[string][]*RemoteArc)
	for _, currArc := range raList {
		copy := currArc
		if _, ok := ans[currArc.Address]; !ok {
			ans[currArc.Address] = []*RemoteArc{&copy}
		} else {
			ans[currArc.Address] = append(ans[currArc.Address], &copy)
		}
	}
	return ans
}

func createArcsWithAddress(arcList []RemoteArc, contextToAddrs map[string][]string, myAddr string) []RemoteArc {
	var newArcs []RemoteArc
	for _, item := range arcList {
		copy := item
		addrs := contextToAddrs[copy.Context]
		for _, addr := range addrs {
			if myAddr != addr { // avoid connecting with self
				copy.Address = addr
				newArcs = append(newArcs, copy)
			}
		}
	}
	return newArcs
}

func (t *RemoteTransition) UpdateAddressByContext(contextToAddrs map[string][]string, myAddr string) {
	t.InArcs = createArcsWithAddress(t.InArcs, contextToAddrs, myAddr)
	t.OutArcs = createArcsWithAddress(t.OutArcs, contextToAddrs, myAddr)
	t.InhibitorArcs = createArcsWithAddress(t.InhibitorArcs, contextToAddrs, myAddr)
}

func (t *RemoteTransition) GetInArcsByAddrs() map[string][]*RemoteArc {
	return convertToAddressArcsMap(t.InArcs)
}

func (t *RemoteTransition) GetOutArcsByAddrs() map[string][]*RemoteArc {
	return convertToAddressArcsMap(t.OutArcs)
}

func remoteArcListToPlaceIDsByAddrsMap(arcList []RemoteArc, ans map[string][]int) {
	for _, inarc := range arcList {
		list, ok := ans[inarc.Address]
		if !ok {
			ans[inarc.Address] = []int{inarc.PlaceID}
		} else {
			ans[inarc.Address] = append(list, inarc.PlaceID)
		}
	}
}

// GetAllPlaceIDsByAddrs gets a map from address to list of all place ids from that address
// (in arcs, inhib arcs and out arcs)
func (t *RemoteTransition) GetAllPlaceIDsByAddrs() map[string][]int {
	ans := make(map[string][]int)
	remoteArcListToPlaceIDsByAddrsMap(t.InArcs, ans)
	remoteArcListToPlaceIDsByAddrsMap(t.InhibitorArcs, ans)
	remoteArcListToPlaceIDsByAddrsMap(t.OutArcs, ans)
	return ans
}

// GetPlaceIDsByAddrs gets a map from address to list of place ids needed from that address
// (NEEDED measn INARCS and INHIBITOR ARCS [i.e., we dont need to know out arcs marks])
func (t *RemoteTransition) GetPlaceIDsByAddrs() map[string][]int {
	ans := make(map[string][]int)
	remoteArcListToPlaceIDsByAddrsMap(t.InArcs, ans)
	remoteArcListToPlaceIDsByAddrsMap(t.InhibitorArcs, ans)
	return ans
}

func (t *RemoteTransition) addInArc(remoteArc RemoteArc) {
	t.InArcs = append(t.InArcs, remoteArc)
}

func (t *RemoteTransition) addOutArc(remoteArc RemoteArc) {
	t.OutArcs = append(t.OutArcs, remoteArc)
}

func (t *RemoteTransition) addInhibitorArc(remoteArc RemoteArc) {
	t.InhibitorArcs = append(t.InhibitorArcs, remoteArc)
}

func (t *RemoteTransition) generateTransitionsByContext(contextToAddrs map[string][]string) []RemoteTransition {
	myContexts := make(map[string]bool)
	for _, inarc := range t.InArcs {
		myContexts[inarc.Context] = true
	}
	for _, outarc := range t.OutArcs {
		myContexts[outarc.Context] = true
	}
	for _, inhibarc := range t.InhibitorArcs {
		myContexts[inhibarc.Context] = true
	}
	myCntxtToAddrs := make(map[string][]string)
	for ctx := range myContexts {
		addrs := contextToAddrs[ctx]
		cpy := make([]string, len(addrs))
		copy(cpy, addrs)
		myCntxtToAddrs[ctx] = cpy
	}
	currConfig := make(map[string]string) // map from context to ONE address (current addr)
	indAddrs := make([][]string, len(myCntxtToAddrs))
	indToCtxt := make(map[int]string)
	currIndex := 0
	expectedConfigNum := 1
	for ctx, addrs := range myCntxtToAddrs {
		expectedConfigNum *= len(addrs)
		indAddrs[currIndex] = addrs
		indToCtxt[currIndex] = ctx
		currIndex++
	}
	ans := make([]RemoteTransition, expectedConfigNum)
	if expectedConfigNum == 0 {
		return ans
	}
	currIndex = 0
	doneF := func() {
		cf := make(map[string][]string)
		for ctx, addr := range currConfig {
			cf[ctx] = []string{addr}
		}
		ans[currIndex] = RemoteTransition{
			t.ID,
			createArcsWithAddress(t.InArcs, cf, ""),
			createArcsWithAddress(t.OutArcs, cf, ""),
			createArcsWithAddress(t.InhibitorArcs, cf, "")}
		currIndex++
	}

	generateConfigurations(0, indAddrs, indToCtxt, currConfig, doneF)
	return ans
}

func generateConfigurations(currIndex int, addrMatrix [][]string, indToCtxt map[int]string, currConfig map[string]string, doneF func()) {
	if currIndex == len(addrMatrix) {
		doneF()
	} else {
		myAddrs := addrMatrix[currIndex]
		myCtx := indToCtxt[currIndex]
		for _, addr := range myAddrs {
			currConfig[myCtx] = addr
			generateConfigurations(currIndex + 1, addrMatrix, indToCtxt, currConfig, doneF)
		}
	}
}
