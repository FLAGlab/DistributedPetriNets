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
	for _, inarc := range raList {
		if _, ok := ans[inarc.Address]; !ok {
			ans[inarc.Address] = []*RemoteArc{&inarc}
		} else {
			ans[inarc.Address] = append(ans[inarc.Address], &inarc)
		}
	}
	return ans
}

func createArcsWithAddress(arcList []RemoteArc, contextToAddrs map[string][]string) []RemoteArc {
	var newArcs []RemoteArc
	for _, item := range arcList {
		copy := item
		addrs := contextToAddrs[copy.Context]
		for _, addr := range addrs {
			copy.Address = addr
			newArcs = append(newArcs, copy)
		}
	}
	return newArcs
}

func (t *RemoteTransition) UpdateAddressByContext(contextToAddrs map[string][]string) {
	t.InArcs = createArcsWithAddress(t.InArcs, contextToAddrs)
	t.OutArcs = createArcsWithAddress(t.OutArcs, contextToAddrs)
	t.InhibitorArcs = createArcsWithAddress(t.InhibitorArcs, contextToAddrs)
}

func (t *RemoteTransition) GetInArcsByAddrs() map[string][]*RemoteArc {
	return convertToAddressArcsMap(t.InArcs)
}

func (t *RemoteTransition) GetOutArcsByAddrs() map[string][]*RemoteArc {
	return convertToAddressArcsMap(t.OutArcs)
}


// GetPlaceIDsByAddrs gets a map from address to list of place ids needed from that address
func (t *RemoteTransition) GetPlaceIDsByAddrs() map[string][]int {
	ans := make(map[string][]int)
	for _, inarc := range t.InArcs {
		_, ok := ans[inarc.Address]
		if !ok {
			ans[inarc.Address] = []int{inarc.PlaceID}
		} else {
			ans[inarc.Address] = append(ans[inarc.Address], inarc.PlaceID)
		}
	}
	for _, inarc := range t.InhibitorArcs {
		_, ok := ans[inarc.Address]
		if !ok {
			ans[inarc.Address] = []int{inarc.PlaceID}
		} else {
			ans[inarc.Address] = append(ans[inarc.Address], inarc.PlaceID)
		}
	}
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
