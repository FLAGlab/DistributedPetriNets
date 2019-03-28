package communication

import (
	"fmt"
	"time"

	"github.com/perlin-network/noise/skademlia"
)

// MinTimeout the minimum timeout for raft
const MinTimeout = 1000 // milliseconds
// MaxTimeout the max timeout for raft
const MaxTimeout = 5000 // milliseconds
// LeaderTimeout to wait for ask response
const LeaderTimeout = 500 // milliseconds
const humanTimeout = 5000

// NodeType enums for PetriNodes communication
type NodeType string

const (
	// Leader is leader node
	Leader    NodeType = "leader"
	// Follower is follower node
	Follower  NodeType = "follower"
	// Candidate is candidate node
	Candidate NodeType = "candidate"
)

// RaftNode in charge of raft and manipulating the given petriNode
type RaftNode struct {
	currentTerm int
  myVotes map[string]string
  nodeType NodeType
  pMsg chan petriMessage
  pNode *petriNode
	timeoutCount int
	votedFor string
}

// InitRaftNode gets a new raft node that contains the petri node
func InitRaftNode(pNode *petriNode, isLeader bool) *RaftNode {
  rn := &RaftNode {pNode: pNode, pMsg: make(chan petriMessage)}
	if isLeader {
		rn.setNodeType(Leader)
	} else {
		rn.setNodeType(Follower)
	}
  return rn
}

func (rn *RaftNode) close() {
	close(rn.pMsg)
}

func (rn *RaftNode) setNodeType(nodeType NodeType) {
	fmt.Printf("_TYPE %v\n", nodeType)
  rn.nodeType = nodeType
	currTimeout := LeaderTimeout
	if nodeType != Leader {
		currTimeout = MinTimeout + getRand(MaxTimeout - MinTimeout)
    rn.pNode.resetStep()
	}
	rn.myVotes = make(map[string]string)
	rn.timeoutCount = currTimeout
	rn.pNode.timeoutCount = currTimeout
	fmt.Printf("TIMEOUT %v FOR TYPE %v\n", rn.timeoutCount + humanTimeout, nodeType)
}

// Listen Function that listens to the channel
func (rn *RaftNode) Listen() {
  for {
		fmt.Printf("STARTED ITERATION AS %v\n", rn.nodeType)
    if rn.nodeType == Leader {
      for rn.pNode.step != RECEIVING_TRANSITIONS_STEP && rn.pNode.step != RECEIVING_MARKS_STEP {
				fmt.Printf("IS LEADER AT STEP: %v\n", rn.pNode.step)
        switch rn.pNode.step {
        case ASK_STEP:
					fmt.Println("WILL ASK")
					fmt.Println("_ASK")
          rn.pNode.ask(rn.generateBaseMessage())
					total := len(skademlia.Table(rn.pNode.node).GetPeers()) + 1 // plus me
					if total == 1 && rn.pNode.step == RECEIVING_TRANSITIONS_STEP {
						rn.pNode.incStep() // skip RECEIVING_TRANSITIONS_STEP if Im the only one
					}
        case PREPARE_FIRE_STEP:
					fmt.Println("WILL PREPARE FIRE TRANSITION")
					fmt.Println("_PREPARE FIRE")
					rn.pNode.prepareFire(rn.generateBaseMessage())
					total := len(skademlia.Table(rn.pNode.node).GetPeers()) + 1 // plus me
					if total == 1 && rn.pNode.step == RECEIVING_MARKS_STEP {
						rn.pNode.incStep() // skip RECEIVING_MARKS_STEP if Im the only one
					}
				case FIRE_STEP:
					fmt.Println("WILL FIRE TRANSITION")
					fmt.Println("_FIRE")
					rn.pNode.fireTransition(rn.generateBaseMessage())
        case PRINT_STEP:
					fmt.Println("WILL PRINT")
					fmt.Println("_PRINT")
          rn.pNode.printPetriNet(rn.generateBaseMessage())
        }
      } // TODO take into account timeout
    }
		fmt.Println("Will wait for msg...")
		fmt.Printf("_WAITING AS %v\n", rn.nodeType)
		if rn.nodeType == Leader {
			fmt.Printf("_WAITING STEP %v\n", rn.pNode.step)
		}
    select {
    case pMsg := <- rn.pMsg:
			fmt.Printf("Msg received: %v\n", pMsg)
			fmt.Printf("_RECEIVED: %v\n", pMsg)
      switch rn.nodeType {
      case Leader:
        rn.processLeader(pMsg)
      case Follower:
        rn.processFollower(pMsg)
      case Candidate:
        rn.processCandidate(pMsg)
      }
    case <- time.After(time.Duration(rn.timeoutCount + humanTimeout) * time.Millisecond):
			fmt.Println("_TIMEOUT")
			fmt.Printf("%v milliseconds already passed...", rn.timeoutCount + humanTimeout)
      if rn.nodeType == Leader {
				fmt.Println("leader timed out")
        rn.pNode.resetStep()
      } else { // Candidate and Follower
				fmt.Println("will do election !!")
        rn.assembleElection()
      }
    }

  }
}

func (rn *RaftNode) processLeader(pMsg petriMessage) {
  if (pMsg.Command == RequestVoteCommand || pMsg.FromType == Leader) &&
      pMsg.Term >= rn.currentTerm {
		rn.currentTerm = pMsg.Term
    rn.setNodeType(Follower)
		if pMsg.Command == RequestVoteCommand {
			rn.vote(pMsg.Term, pMsg.Address)
		}
  } else if pMsg.Term == rn.currentTerm {
		rn.pNode.updateCtx(pMsg)
		switch rn.pNode.step {
		case RECEIVING_TRANSITIONS_STEP:
    	rn.pNode.getTransition(pMsg)
		case RECEIVING_MARKS_STEP:
			rn.pNode.getPlaceMarks(pMsg)
		}
  }
}

func (rn *RaftNode) processFollower(pMsg petriMessage) {
  fmt.Printf("Received msg: %v\n", pMsg)
  if pMsg.Term >= rn.currentTerm {
    rn.currentTerm = pMsg.Term
    if pMsg.Command == RequestVoteCommand {
  		fmt.Println("WILL VOTE")
  		rn.vote(pMsg.Term, pMsg.Address)
    } else {
      rn.pNode.processMessage(pMsg, rn.generateBaseMessage())
    }
    if pMsg.Command != RequestVoteCommand {
      rn.votedFor = "" // theres a leader, I'll be ready for new elections TODO revisar
    }
  }
}

func (rn *RaftNode) closePolls() {
	fmt.Println("POLLS ARE CLOSED!!!")
	countMap := make(map[string]int)
	maxVotes := 0
	maxVoteAddress := ""
	for _, voteAddr := range rn.myVotes {
		countMap[voteAddr]++
		if countMap[voteAddr] > maxVotes {
			maxVotes = countMap[voteAddr]
			maxVoteAddress = voteAddr
		}
	}
	fmt.Printf("WINNER: %v, COUNT: %v\n", maxVoteAddress, maxVotes)
	if maxVoteAddress == rn.pNode.node.ExternalAddress() { // I won!!
		fmt.Println("LEADER SETTED AS ME !!! >:v")
		rn.setNodeType(Leader)
	} else {
		rn.setNodeType(Follower)
	}
}

func (rn *RaftNode) processCandidate(pMsg petriMessage) {
  fmt.Printf("Processing msg as candidate: %v", pMsg)
  if pMsg.Term > rn.currentTerm {
    // theres a leader !!
    rn.currentTerm = pMsg.Term
    rn.setNodeType(Follower)
    if pMsg.FromType == Leader { // in case it is a leader
      rn.processFollower(pMsg)
    } // else -> not a leader and is voting, thou
      // his vote will not be for me because he has a bigger Term
  } else if pMsg.Command == RequestVoteCommand { // someone else wants me to vote
    fmt.Println("someone else wants my vote D:")
    rn.vote(pMsg.Term, pMsg.Address)
  } else if pMsg.Command == VoteCommand { // its a vote
    fmt.Printf("Received %v vote from: %v\n", pMsg.VoteGranted, pMsg.Address)
    rn.myVotes[pMsg.Address] = pMsg.VoteGranted
    total := len(skademlia.Table(rn.pNode.node).GetPeers()) + 1 // plus me
    fmt.Printf("Total of votes: %v\n", rn.myVotes)
    fmt.Printf("Total of peers: %v\n", total)
    if len(rn.myVotes) == total { // polls are closed!
      rn.closePolls()
    }
  } // else is an old leader msg, ignore
}

func (rn *RaftNode) assembleElection() {
  rn.setNodeType(Candidate)
  myAddr := rn.pNode.node.ExternalAddress()
  rn.myVotes[myAddr] = myAddr
  rn.currentTerm++
  rn.votedFor = myAddr
	total := len(skademlia.Table(rn.pNode.node).GetPeers()) + 1 // plus me
	if len(rn.myVotes) == total { // polls are closed!
		rn.closePolls()
	} else {
		successCallback := func() {
			fmt.Println("assembleElection done correclty")
		}
	  timeoutCallback := func () {
			fmt.Println("assembleElection didnt wor")
	    rn.setNodeType(Candidate)
	  }
	  rn.pNode.broadcastWithTimeout(rn.generateMessageWithCommand(RequestVoteCommand),
	    successCallback, timeoutCallback)
	}
}

func (rn *RaftNode) vote(candidateTerm int, candidateAddress string) {
  ans := rn.generateMessageWithCommand(VoteCommand)
	fmt.Println("WILL VOTE")
	fmt.Printf("my term: %v, msg term: %v, my last vote for: %v\n",
    rn.currentTerm, candidateTerm, rn.votedFor)
	if rn.currentTerm <= candidateTerm &&
      (rn.votedFor == "" || rn.votedFor == candidateAddress) {
		ans.VoteGranted = candidateAddress
		rn.currentTerm = candidateTerm
		rn.votedFor = candidateAddress
	} else {
		ans.VoteGranted = rn.votedFor
	}
	fmt.Printf("My vote: %v\n", ans.VoteGranted)
	rn.pNode.SendMessageByAddress(ans, candidateAddress)
}

func (rn *RaftNode) generateBaseMessage() petriMessage {
	return petriMessage {
		Address: rn.pNode.node.ExternalAddress(),
		Term: rn.currentTerm,
		PetriContext: rn.pNode.petriNet.Context,
		FromType: rn.nodeType}
}

func (rn *RaftNode) generateMessageWithCommand(command CommandType) petriMessage {
  baseMsg := rn.generateBaseMessage()
  baseMsg.Command = command
	return baseMsg
}
