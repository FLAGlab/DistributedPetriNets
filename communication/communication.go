package communication

import (
	"fmt"
	"flag"
	"strconv"
	"time"
	"math/rand"

	"github.com/FLAGlab/DCoPN/petribuilder"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/cipher/aead"
	"github.com/perlin-network/noise/handshake/ecdh"
	"github.com/perlin-network/noise/log"
	"github.com/perlin-network/noise/protocol"
	"github.com/perlin-network/noise/skademlia"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

//Generates random int as function of range
func getRand(Range int) int {
    return r.Intn(Range)
}

type MyPeerNode struct {
	peer *noise.Peer
}

func (cpeer MyPeerNode) SendMessage(pMsg petriMessage) error {
	return cpeer.peer.SendMessage(pMsg)
}

type MyCommunicationNode struct {
	node *noise.Node
}

func (cn *MyCommunicationNode) CountPeers() int {
	return len(skademlia.Table(cn.node).GetPeers())
}

func (cn *MyCommunicationNode) Broadcast(pMsg petriMessage) []error {
	return skademlia.Broadcast(cn.node, pMsg)
}

func (cn *MyCommunicationNode) ExternalAddress() string {
	return cn.node.ExternalAddress()
}

func (cn *MyCommunicationNode) Dial(address string) (PeerNode, error) {
	peer, err := cn.node.Dial(address)
	return &MyPeerNode{peer}, err
}

/** ENTRY POINT **/
func setup(rn *RaftNode, noiseNode *noise.Node) {
	opcodeChat := noise.RegisterMessage(noise.NextAvailableOpcode(), (*petriMessage)(nil))
	channelCount := 0
	noiseNode.OnPeerConnected(func(node *noise.Node, peer *noise.Peer) error {
		fmt.Printf("PEER CONNECTED -> %v:%v\n", peer.RemoteIP(), peer.RemotePort())
		return nil
	})
	noiseNode.OnPeerInit(func(node *noise.Node, peer *noise.Peer) error {
		// init se llama cuando se conecta un nodo o se le hace dial
		channelCount++
		myChannel := channelCount
		fmt.Printf("Created channel %v\n", channelCount)
		peer.OnConnError(func(node *noise.Node, peer *noise.Peer, err error) error {
			log.Info().Msgf("Got an error: %v", err)
			return nil
		})

		peer.OnDisconnect(func(node *noise.Node, peer *noise.Peer) error {
			log.Info().Msgf("Peer %v has disconnected.",
				peer.RemoteIP().String()+":"+strconv.Itoa(int(peer.RemotePort())))
			return nil
		})
		// acá solo se comunica con el peer que se acaba de inicializar
		go func() {
			timesUsed := 0
			for msg := range peer.Receive(opcodeChat) {
				rn.pMsg <- msg.(petriMessage)
				timesUsed++
				fmt.Printf("HERE: Used channel %v (%v times)\n", myChannel, timesUsed)
			}
			// TODO: usar timeouts grandes para terminar la gorrutina si peer no recibe mas msg
			fmt.Printf("HERE: Closed channel %v\n", myChannel)
		}()
		return nil
	})
}

// Run function that starts everything
func Run() {
	//gob.Register(skademlia.ID{})
	hostFlag := flag.String("h", "127.0.0.1", "host to listen for peers on")
	portFlag := flag.Uint("p", 3000, "port to listen for peers on")
	leaderFlag := flag.Bool("l", false, "is leader node")
	contextName := flag.String("ctx", "", "Context name of this petri node")
	dependencyRelationsFile := flag.String("cdr", "./configuration/cdr.json", "Json with the context dependency relations")
	flag.Parse()

	params := noise.DefaultParams()
	//params.NAT = nat.NewPMP()
	params.Keys = skademlia.RandomKeys()
	params.Host = *hostFlag
	params.Port = uint16(*portFlag)

	node, err := noise.NewNode(params)
	if err != nil {
		panic(err)
	}
	defer node.Kill()

	p := protocol.New()
	p.Register(ecdh.New())
	p.Register(aead.New())
	p.Register(skademlia.New())
	p.Enforce(node)

	// Build my node context petri net
	pnet := petribuilder.CreateContext(*contextName)
	// update pnet with the required remote arcs of the context dependency relations
	petribuilder.UpdatePetriNetWithCDR(pnet, *dependencyRelationsFile)

	// build petri node and update its universal petri net
	pnNode := InitPetriNode(&MyCommunicationNode{node}, pnet)
	pnNode.universalPetriNet = petribuilder.GetUniversalPetriNet(*dependencyRelationsFile)
	// populate the pnNode conflic solver with strategies according to the context dependency relations
	petribuilder.UpdateConflictSolverWithCDR(&pnNode.conflictSolver, *dependencyRelationsFile)

	rn := InitRaftNode(pnNode, *leaderFlag)
	defer rn.close()
	go rn.Listen()
	setup(rn, node)
	go node.Listen()

	log.Info().Msgf("Listening for peers on port %d.", node.ExternalPort())

	if len(flag.Args()) > 0 {
		for _, address := range flag.Args() {
			peer, err := node.Dial(address)
			if err != nil {
				panic(err)
			}
			skademlia.WaitUntilAuthenticated(peer)
		}

		peers := skademlia.FindNode(node, protocol.NodeID(node).(skademlia.ID), skademlia.BucketSize(), 8)
		log.Info().Msgf("Bootstrapped with peers: %+v", peers)
	}

	for {}
}
